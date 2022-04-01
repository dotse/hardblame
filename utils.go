package main

import (
        "fmt"
	"io/ioutil"
	"os"
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

