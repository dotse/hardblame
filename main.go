package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
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

	str2points := map[string]int{
		"good":    1,
		"neutral": 0,
		"warning": -1,
		"error":   -2,
	}

	// Command line parameters
	config := getConfig()

	// login to web interface
	if config.Verbose {
		log.Println("Login to web interface")
	}
	hardclient := GetHardenizeClient(config.HardenizeUser, config.HardenizePasswd, config.HardenizeWebUser, config.HardenizeWebPasswd)

	// Get group list from hardenize
	if config.Verbose {
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
			if config.Verbose {
				log.Printf("Skip group %s\n", group.Name)
			}
			continue
		}
		if config.Verbose {
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
				continue
			}

			// compute host points
			host := HostStat{Name: record[0]}
			// nameServers
			host.DNSpoints += str2points[record[31]]
			// dnssec
			host.DNSpoints += str2points[record[32]]
			// emailTls
			host.EMAILpoints += str2points[record[33]]
			// emailDane
			host.EMAILpoints += str2points[record[34]]
			// spf
			host.EMAILpoints += str2points[record[35]]
			// dmarc
			host.EMAILpoints += str2points[record[36]]
			// wwwTls
			host.WEBpoints += str2points[record[37]]
			// wwwDane
			// host.WEBpoints += str2points[record[38]]
			// hsts
			// host.WEBpoints += str2points[record[39]]
			// hpkp
			// host.WEBpoints += str2points[record[40]]
			// csp
			host.WEBpoints += str2points[record[41]]
			// securityHeaders
			host.WEBpoints += str2points[record[42]]
			// cookies
			host.WEBpoints += str2points[record[43]]
			// mixedContent
			host.WEBpoints += str2points[record[44]]
			// wwwXssProtection
			host.WEBpoints += str2points[record[45]]

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
	if config.Verbose {
		fmt.Println("\n\nRANKED")
		for n := range groupstats {
			fmt.Printf("%2d %-25s %3d\n", n, groupstats[n].Name, groupstats[n].RANK)
		}
	}

	// compute rank hosts
	for g := range groupstats {
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
		if config.Verbose {
			fmt.Println("\n\nRANKED ", groupstats[g].Name)
			for n := range groupstats[g].HostStats {
				fmt.Printf("%2d %-25s %3d\n", n, groupstats[g].HostStats[n].Name, groupstats[g].HostStats[n].RANK)
			}
		}

	}
	if !config.Dryrun {

		today := fmt.Sprintf("%4d-%02d-%02d", time.Now().Year(), time.Now().Month(), time.Now().Day())

		if config.Verbose {
			log.Printf("Marshal %s data to json\n", today)
		}

		data := map[string][]GroupStat{today: groupstats}

		// compute data
		b, err := json.Marshal(data)
		if err != nil {
			panic(err)
		}

		jsonfilename := fmt.Sprintf("data-%d-%02d-%02d.json", time.Now().Year(), time.Now().Month(), time.Now().Day())
		if config.Verbose {
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
	}

	// done
	if config.Verbose {
		log.Println("Done")
	}
}
