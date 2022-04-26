package main

import (
        "fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

const datadir = "2006-01-02"
const datafile = "15:04:05"

var data2points map[string]int = map[string]int{
	"good":    1,
	"neutral": 0,
	"warning": -1,
	"error":   -2,
}

func str2points(value string) int {
	return data2points[value]
}

func WriteJsonGroups(filetmpl, group string, json []byte) error {
     day := time.Now().Format(datadir)
     outdir := filetmpl + "/" + day
     os.Mkdir(outdir, 0755)
     // data/2022-04-26/ispar-2022-04-26.json
     outfile := fmt.Sprintf("%s/%s/%s-%s.json", filetmpl, day, group,
     	     					time.Now().Format(datafile))
     fmt.Printf("WriteJsonData: outfile='%s', day='%s', group='%s'\n", outfile, day, group)
     err := ioutil.WriteFile(outfile, json, 0644)
     if err != nil {
     	log.Fatalf("Error writing json blob received from Hardenize to file %s: %v", outfile, err)
     }
     
     return err
}

func WriteJsonData(conf *Config, filetmpl, group string, jsondata []byte) error {
     day := time.Now().Format(datadir)
     outdir := filetmpl + "/" + day
     os.Mkdir(outdir, 0755)
     // data/2022-04-26/ispar-2022-04-26.json
     outfile := fmt.Sprintf("%s/%s/%s-%s.json", filetmpl, day, group,
     	     					time.Now().Format(datafile))
     fmt.Printf("WriteJsonData: outfile='%s', day='%s', group='%s'\n", outfile, day, group)
     err := ioutil.WriteFile(outfile, jsondata, 0644)
     if err != nil {
     	log.Fatalf("Error writing json blob received from Hardenize to file %s: %v", outfile, err)
     }

     err = conf.HardDB.AddGroupDay(group, day, string(jsondata))
     if err != nil {
     	log.Fatalf("Error writing json blob received from Hardenize to db: %v", err)
     }
     
     
     return err
}

func WriteCSVInput(filetmpl, group string, csv [][]string) error {
     day := time.Now().Format(datadir)
     outdir := filetmpl + "/" + day
     os.Mkdir(outdir, 0755)
     outfile := fmt.Sprintf("%s/%s/%s-%s.csv", filetmpl, day, group,
     	     					time.Now().Format(datafile))

     out := []string{}
     for _, row := range csv {
     	 line := strings.Join(row, ", ")
	 out = append(out, line)
     }
     result := strings.Join(out, "\n")
     
     err := ioutil.WriteFile(outfile, []byte(result), 0644)
     return err
}

// for key, _ := range group.Hosts {
//         if group.Hosts[key].Hostname[0:4] == "xn--" {
//             newr, err := idna.ToUnicode(group.Hosts[key].Hostname)
// 	    	if err != nil {
//                 panic(err)
//             }
//             group.Hosts[key].Hostname = newr
//         }
//         fmt.Println(group.Hosts[key].Hostname)
// }
// 
// for _, data := range group.Hosts {
//         if data.Hostname[0:4] == "xn--" {
//             newr, err := idna.ToUnicode(data.Hostname)
// 	    	if err != nil {
//                 panic(err)
//             }
//             data.Hostname = newr
//         }
//         fmt.Println(data.Hostname)
// }

func UpdateCounters(gs *[]GroupStat, ad *map[string]Group) {
     log.Printf("Enter UpdateCounters")
     
     for k, v := range *ad {
     	 stat := GroupStat{Name: k}
     	 for _, h := range v.Hosts {
     	     host := HostStat{Name: h.Hostname}
	     host.EMAILpoints += str2points(h.EmailTls)
	     host.EMAILpoints += str2points(h.EmailDane)
	     host.EMAILpoints += str2points(h.Spf)
	     host.EMAILpoints += str2points(h.Dmarc)
	     fmt.Printf("Examining Group %s, host %s. Email points: %d\n",
	     			   v.Name, h.Hostname, host.EMAILpoints)

	     stat.DNSpoints += host.DNSpoints
	     stat.EMAILpoints += host.EMAILpoints
	     stat.WEBpoints += host.WEBpoints
	     stat.TOTALpoints += host.TOTALpoints
	     stat.HostStats = append(stat.HostStats, host)
	 }
	 fmt.Printf("Stats for group %s: Email points: %d\n", v.Name, stat.EMAILpoints)
     }
}