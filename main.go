package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/net/idna"
	"github.com/spf13/viper"
)

func mainloop(conf *Config) {
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)

	for {
	    select {
	    	   case <-exit:
		   os.Exit(0)
	    }
	}     
}

func main() {
        var conf Config

	// Command line parameters
	// config := getConfig()

	viper.SetConfigFile(DefaultCfgFile)

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Could not load config (%s)", err)
	}

	ValidateConfig(nil, DefaultCfgFile, false) // will terminate on error
	viper.Unmarshal(&conf)

	conf.HardDB = NewDB(viper.GetString("db.file"), false) // Don't drop status tables if they exist

	verbose := viper.GetBool("log.verbose")

	go APIdispatcher(&conf)

	mainloop(&conf)

	// login to web interface
	if conf.Log.Verbose == true {
		log.Println("Login to web interface")
	}
	h := conf.Hardenize
	conf.HClient = GetHardenizeClient(h.APIUrl, h.Organisation, 
					 h.User, h.Passwd, h.WebUser,
					 h.WebPasswd)
					 
	// Get group list from hardenize
	if verbose {
		log.Println("Get Groups")
	}

	// endpoint := fmt.Sprintf("%s", "groups")
	body := conf.HClient.GetAPIData("groups")
	err = WriteJsonGroups("data", "groups", body)
	var groups hgroups
	err = json.Unmarshal(body, &groups)
	if err != nil {
		log.Fatal(err)
	}

	err, groupstats, alldata := FetchAllData(&conf, groups)
	if err != nil {
	   log.Fatalf("Error from FetchAllData: %v", err)
	}

	UpdateCounters(&groupstats, &alldata)

	// Nothing to do if we just investigate one domain
//	if config.Domain != "" {
//		log.Printf("Domain %s not found.\n", config.Domain)
//		os.Exit(0)
//	}

	// compute rank groups
	if verbose {
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
		if verbose {
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

	if verbose {
		log.Printf("Marshal %s data to json\n", today)
	}

	data := map[string][]GroupStat{today: groupstats}

	// compute data
	b, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	jsonfilename := fmt.Sprintf("data-%d-%02d-%02d.json", time.Now().Year(), time.Now().Month(), time.Now().Day())
	if verbose {
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
	if verbose {
		log.Println("Done")
	}
}

func FetchAllData(conf *Config, groups hgroups) (error, []GroupStat,
					 		 	 map[string]Group) {
	log.Printf("Enter FetchAllData")
	h := conf.Hardenize
	url := fmt.Sprintf("%s/%s/%s", h.APIUrl, h.Organisation, "groups")
	basenametmpl := "data"

	verbose := viper.GetBool("log.verbose")
	loglevel := viper.GetInt("log.level")
	alldata := map[string]Group{}

	groupstats := make([]GroupStat, 10)

	for _, group := range groups.Groups {
		if string(group.Name[0]) == "#" {
			if verbose && loglevel > 1 {
				log.Printf("Skip group %s\n", group.Name)
			}
			continue
		}
		if verbose {
			log.Printf("Process group %s\n", group.Name)
		}

		// keep statistics
		stat := GroupStat{Name: group.Name, Id: group.Id}

		// get CSV file
		// example: https: //www.hardenize.com/org/sweden-health-status/hosts/statligtgdabolag?format=csv
		url = fmt.Sprintf("%s/org/%s/hosts/%s?format=csv", h.WebUrl,
		      						   h.Organisation,
								   group.Id)
		jsonurl := fmt.Sprintf("reports0?group=%s&format=json", group.Id)
		jsondata := conf.HClient.GetAPIData(jsonurl)
		err := WriteJsonData(conf, basenametmpl, group.Id, jsondata)
		if err != nil {
		   log.Fatalf("Error writing json blob received from Hardenize: %v", err)
		}

		jsongroup := Group{ Name: group.Id }
		err = json.Unmarshal(jsondata, &jsongroup)
		if err != nil {
		   log.Fatalf("json.Unmarshal says kaboom: %v", err)
		}
		alldata[group.Id] = jsongroup

		// fmt.Printf("Our group struct: %v\n\n\n", jsongroup)		

		csv := conf.HClient.GetCSV(url)
		err = WriteCSVInput(basenametmpl, group.Id, csv)
		if err != nil {
		   log.Fatalf("Error writing json blob received from Hardenize: %v", err)
		}
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

			// we are just investigating one domain
//			if config.Domain != "" {
//				if verbose {
//					log.Printf("Config.Domain: %s    Record[0]: %s    Host: %s\n", config.Domain, record[0], host.Name)
//				}
//				if config.Domain != record[0] && config.Domain != host.Name {
//					continue
//				}
//
//				// Domain found, print values
//				log.Println("Hälsoläget för ", config.Domain)
//			}
			// nameServers
			host.DNSpoints += str2points(record[30])
//			if config.Domain != "" {
//				log.Printf("%-20s%3d\n", "DNS", str2points(record[30]))
//			}
			// dnssec
			host.DNSpoints += str2points(record[31])
//			if config.Domain != "" {
//				log.Printf("%-20s%3d\n", "DNSSEC", str2points(record[31]))
//			}
			// emailTls
			host.EMAILpoints += str2points(record[32])
//			if config.Domain != "" {
//				log.Printf("%-20s%3d\n", "EMAIL TLS", str2points(record[32]))
//			}
			// emailDane
			host.EMAILpoints += str2points(record[33])
//			if config.Domain != "" {
//				log.Printf("%-20s%3d\n", "EMAIL DANE", str2points(record[33]))
//			}
			// spf
			host.EMAILpoints += str2points(record[34])
//			if config.Domain != "" {
//				log.Printf("%-20s%3d\n", "EMAIL SPF", str2points(record[34]))
//			}
			// dmarc
			host.EMAILpoints += str2points(record[35])
//			if config.Domain != "" {
//				log.Printf("%-20s%3d\n", "EMAIL DMARC", str2points(record[35]))
//			}
			// wwwTls
			host.WEBpoints += str2points(record[36])
//			if config.Domain != "" {
//				log.Printf("%-20s%3d\n", "WEB TLS", str2points(record[36]))
//			}
			// wwwDane
			// host.WEBpoints += str2points[record[37]]
			// if config.Verbose==7 {
			//	log.Printf("%-20s%3d\n","WEB DANE", str2points[record[37]))
			// }
			// hsts
			// host.WEBpoints += str2points[record[38]]
			// if config.Verbose==7 {
			//	log.Printf("%-20s%3d\n","WEB HSTS", str2points[record[38]))
			// }
			// hpkp
			// host.WEBpoints += str2points[record[39]]
			// if config.Verbose==7 {
			// 	log.Printf("%-20s%3d\n","WEB HPKP", str2points[record[39]))
			// }
			// csp
			host.WEBpoints += str2points(record[40])
//			if config.Domain != "" {
//				log.Printf("%-20s%3d\n", "WEB CSP", str2points(record[40]))
//			}
			// securityHeaders
			host.WEBpoints += str2points(record[41])
//			if config.Domain != "" {
//				log.Printf("%-20s%3d\n", "WEB Security headers", str2points(record[41]))
//			}
			// cookies
			host.WEBpoints += str2points(record[42])
//			if config.Domain != "" {
//				log.Printf("%-20s%3d\n", "WEB Cookies", str2points(record[42]))
//			}
			// mixedContent
			host.WEBpoints += str2points(record[43])
//			if config.Domain != "" {
//				log.Printf("%-20s%3d\n", "WEB mixed content", str2points(record[43]))
//			}
			// wwwXssProtection
			host.WEBpoints += str2points(record[44])
//			if config.Domain != "" {
//				log.Printf("%-20s%3d\n", "WEB XSS", str2points(record[44]))
//			}

			// host TOTAL
			host.TOTALpoints = host.DNSpoints + host.EMAILpoints + host.WEBpoints
//			if config.Domain != "" {
//				log.Println("\n\n")
//				log.Printf("DNS :  %3d\n", host.DNSpoints)
//				log.Printf("EMAIL: %3d\n", host.EMAILpoints)
//				log.Printf("WEB:   %3d\n", host.WEBpoints)
//				log.Println("----------")
//				log.Printf("TOTAL: %3d\n", host.TOTALpoints)
//
//				// Done
//				os.Exit(0)
//			}

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
	return nil, groupstats, alldata
}