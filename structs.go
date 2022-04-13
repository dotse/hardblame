package main

import (
//	"encoding/json"
	"net/http"
)

const SECTION_SMTP = "smtp"
const SECTION_HTTP = "http"
const SECTION_NONE = "none"

type HostStat struct {
	Name        string `json:"name"`
	DNSpoints   int    `json:"dns"`
	WEBpoints   int    `json:"web"`
	EMAILpoints int    `json:"email"`
	EMAILpoints2 int   `json:"email"`
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

type Config struct {
     Hardenize	   HardConf
     Log	   LogConf	   
}

type HardConf struct {
     APIUrl   string `validate:"required,url"`
     User     string `validate:"required"`
     Passwd   string
     WebUrl   string
     WebUser  string
     WebPasswd	string
     Organisation	string
}

type LogConf struct {
     Verbose bool
     Level   int
}

type Configuration struct {
	Verbose            int
	Organization       string
	Domain             string
	HardenizeRoot      string
	HardenizeUser      string
	HardenizePasswd    string
	HardenizeWebUser   string
	HardenizeWebPasswd string
	HardenizeWebRoot   string
}

type Group struct {
     Name  string
     Hosts []Host `json:"reports"`
}

type Host struct {
    Hostname                    string          `json:hostname`
    Status                      string          `json:status`
    ReportTime                  string          `json:reportTime`
    NameServers                 string          `json:nameServers`
    Dnssec                      string          `json:dnssec`
    EmailTls                    string          `json:emailTls`
    EmailDane                   string          `json:emailDane`
    Spf                         string          `json:spf`
    Dmarc                       string          `json:dmarc`
    WwwTls                      string          `json:wwwTls`
    Hsts                        string          `json:hsts`
    Hpkp                        string          `json:hpkp`
    WwwDane                     string          `json:wwwDane`
    Csp                         string          `json:csp`
    SecurityHeaders             string          `json:securityHeaders`
    Cookies                     string          `json:cookies`
    MixedContent                string          `json:mixedContent`
    HasDnssec                   bool            `json:hasDnssec`
    HasSmtp                     bool            `json:hasSmtp`
    HasSmtpTls                  bool            `json:hasSmtpTls`
    HasSmtpDane                 bool            `json:hasSmtpDane`
    HasSpf                      bool            `json:hasSpf`
    HasDmarc                    bool            `json:hasDmarc`
    HasHttp                     bool            `json:hasHttp`
    HasHttps                    bool            `json:hasHttps`
    HasHttpsValid               bool            `json:hasHttpsValid`
    HasHttpsDane                bool            `json:hasHttpsDane`
    HasHttpsRedirection         bool            `json:hasHttpsRedirection`
    HasHttpsTls12OrBetter       bool            `json:hasHttpsTls12OrBetter`
    HasSsl2                     bool            `json:hasSsl2`
    HasSsl3                     bool            `json:hasSsl3`
    HasTls10                    bool            `json:hasTls10`
    HasTls11                    bool            `json:hasTls11`
    HasTls12                    bool            `json:hasTls12`
    HasTls13                    bool            `json:hasTls13`
    SupportsSsl2                bool            `json:supportsSsl2`
    SupportsSsl3                bool            `json:supportsSsl3`
    SupportsTls10               bool            `json:supportsTls10`
    SupportsTls11               bool            `json:supportsTls11`
    SupportsTls12               bool            `json:supportsTls12`
    SupportsTls13               bool            `json:supportsTls13`
    HasHsts                     bool            `json:hasHsts`
    HasHstsPreloaded            bool            `json:hasHstsPreloaded`
}

type hgroup struct {
	Id   string
	Name string
}

type hgroups struct {
	Groups []hgroup
}

type hardenizeclient struct {
        baseurl	     	string
	organisation	string
	apiuser   	string
	apipasswd 	string
	webclient 	http.Client
}


