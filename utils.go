package main

import (
        "fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

const datadir = "2006-01-02"
const datafile = "15:04:05"

func WriteJsonInput(filetmpl, group string, json []byte) error {
     day := time.Now().Format(datadir)
     outdir := filetmpl + "/" + day
     os.Mkdir(outdir, 0755)
     outfile := fmt.Sprintf("%s/%s/%s-%s.json", filetmpl, day, group,
     	     					time.Now().Format(datafile))
     err := ioutil.WriteFile(outfile, json, 0644)
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