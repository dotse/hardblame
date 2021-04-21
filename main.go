package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/net/idna"
)

const SECTION_SMTP = "smtp"
const SECTION_HTTP = "http"
const SECTION_NONE = "none"

type HostStat struct {
	Name        string `json:"name"`
	DNSpoints   int    `json:"dns"`
	WEBpoints   int    `json:"web"`
	EMAILpoints int    `json:"email"`
	TOTALpoints int    `json:"total"`
	RANK        int    `json:"rank"`
}

type GroupStat struct {
	Name        string     `json:"name"`
	Id          string     `json:"id"`
	DNSpoints   int        `json:"dns"`
	WEBpoints   int        `json:"web"`
	EMAILpoints int        `json:"email"`
	TOTALpoints int        `json:"total"`
	HostStats   []HostStat `json:"hosts"`
	RANK        int        `json:"rank"`
}

type entry struct {
	name   string
	points float32
}

func main() {

	var url string

	// Command line parameters
	config := getConfig()

	// login to web interface
	if config.Verbose > 0 {
		log.Println("Login to web interface")
	}
	hardclient := GetHardenizeClient(config.HardenizeUser, config.HardenizePasswd, config.HardenizeWebUser, config.HardenizeWebPasswd)

	// Get group list from hardenize
	if config.Verbose > 0 {
		log.Println("Get Groups")
	}

	url = fmt.Sprintf("%s/%s/%s", config.HardenizeRoot, config.Organization, "groups")
	body := hardclient.GetAPIData(url)
	var groups hgroups
	err := json.Unmarshal(body, &groups)
	if err != nil {
		log.Fatal(err)
	}

	groupstats := make([]GroupStat, 0)
	for _, group := range groups.Groups {
		if string(group.Name[0]) == "#" {
			if config.Verbose > 0 {
				log.Printf("Skip group %s\n", group.Name)
			}
			continue
		}
		if config.Verbose > 0 {
			log.Printf("Process group %s\n", group.Name)
		}

		// keep statistics
		stat := GroupStat{Name: group.Name, Id: group.Id}

		// get CSV file
		// example: https: //www.hardenize.com/org/sweden-health-status/hosts/statligtgdabolag?format=csv
		url = fmt.Sprintf("%s/org/%s/hosts/%s?format=csv", config.HardenizeWebRoot, config.Organization, group.Id)
		csv := hardclient.GetCSV(url)
		for _, record := range csv {
			// jump over header
			if record[0] == "hostname" {
				if record[30] != "nameServers" {
					log.Printf("record index 30 is %s", record[30])
					log.Fatal("Index to CSV is broken")
				}
				continue
			}

			// compute host points
			if record[0][0:4] == "xn--" {
				newr, err := idna.ToUnicode(record[0])
				if err != nil {
					panic(err)
				}
				record[0] = newr
			}
			host := HostStat{Name: record[0]}
			// nameServers
			host.DNSpoints += str2points(record[30])
			// dnssec
			host.DNSpoints += str2points(record[31])
			// emailTls
			host.EMAILpoints += str2points(record[32])
			// emailDane
			host.EMAILpoints += str2points(record[33])
			// spf
			host.EMAILpoints += str2points(record[34])
			// dmarc
			host.EMAILpoints += str2points(record[35])
			// wwwTls
			host.WEBpoints += str2points(record[36])
			// wwwDane
			// host.WEBpoints += str2points[record[37]]
			// hsts
			// host.WEBpoints += str2points[record[38]]
			// hpkp
			// host.WEBpoints += str2points[record[39]]
			// csp
			host.WEBpoints += str2points(record[40])
			// securityHeaders
			host.WEBpoints += str2points(record[41])
			// cookies
			host.WEBpoints += str2points(record[42])
			// mixedContent
			host.WEBpoints += str2points(record[43])
			// wwwXssProtection
			host.WEBpoints += str2points(record[44])

			// host TOTAL
			host.TOTALpoints = host.DNSpoints + host.EMAILpoints + host.WEBpoints

			// now keep group stats
			stat.DNSpoints += host.DNSpoints
			stat.EMAILpoints += host.EMAILpoints
			stat.WEBpoints += host.WEBpoints
			stat.TOTALpoints += host.TOTALpoints
			stat.HostStats = append(stat.HostStats, host)
		}

		// store group stats
		groupstats = append(groupstats, stat)
	}

	// compute rank groups
	if config.Verbose > 0 {
		log.Printf("Compute ranks for groups\n")
	}

	list := make([]entry, 0)
	for _, group := range groupstats {
		entry := entry{name: group.Name, points: float32(group.TOTALpoints) / float32(len(group.HostStats))}
		list = append(list, entry)
	}
	for n := 0; n < len(list)-1; n++ {
		for m := n + 1; m < len(list); m++ {
			if list[n].points < list[m].points {
				list[n], list[m] = list[m], list[n]
			}
		}
	}
	for n := range list {
		for m := range groupstats {
			if groupstats[m].Name == list[n].name {
				groupstats[m].RANK = n + 1
			}
		}
	}

	// compute rank hosts
	for g := range groupstats {
		if config.Verbose > 0 {
			log.Printf("Compute ranks for group %s\n", groupstats[g].Name)
		}

		hosts := make([]entry, 0)
		for n := range groupstats[g].HostStats {
			hosts = append(hosts, entry{name: groupstats[g].HostStats[n].Name, points: float32(groupstats[g].HostStats[n].TOTALpoints)})
		}
		for n := 0; n < len(hosts)-1; n++ {
			for m := n + 1; m < len(hosts); m++ {
				if hosts[n].points < hosts[m].points {
					hosts[n], hosts[m] = hosts[m], hosts[n]
				}
			}
		}
		for n := range hosts {
			for m := range groupstats[g].HostStats {
				if groupstats[g].HostStats[m].Name == hosts[n].name {
					groupstats[g].HostStats[m].RANK = n + 1
				}
			}
		}

	}

	today := fmt.Sprintf("%4d-%02d-%02d", time.Now().Year(), time.Now().Month(), time.Now().Day())

	if config.Verbose > 0 {
		log.Printf("Marshal %s data to json\n", today)
	}

	data := map[string][]GroupStat{today: groupstats}

	// compute data
	b, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	jsonfilename := fmt.Sprintf("data-%d-%02d-%02d.json", time.Now().Year(), time.Now().Month(), time.Now().Day())
	if config.Verbose > 0 {
		log.Printf("Write json file %s\n", jsonfilename)
	}

	f, err := os.Create(jsonfilename)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	n, err := f.Write(b)
	if err != nil {
		panic(err)
	}
	if n != len(b) {
		panic(fmt.Errorf("Marshaled data is %d bytes, written only %d bytes", len(b), n))
	}
	f.Sync()

	// done
	if config.Verbose > 0 {
		log.Println("Done")
	}
}

var data2points map[string]int = map[string]int{
	"good":    1,
	"neutral": 0,
	"warning": -1,
	"error":   -2,
}

func str2points(value string) int {
	return data2points[value]
}
