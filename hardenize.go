package main

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"golang.org/x/net/html"
)

type hgroup struct {
	Id   string
	Name string
}

type hgroups struct {
	Groups []hgroup
}

type hardenizeclient struct {
	apiuser   string
	apipasswd string
	webclient http.Client
}

func GetHardenizeClient(apiuser, apipasswd, webuser, webpasswd string) *hardenizeclient {
	options := cookiejar.Options{}
	jar, err := cookiejar.New(&options)
	if err != nil {
		log.Fatal(err)
	}
	client := http.Client{Jar: jar}

	resp, err := client.Get("https://www.hardenize.com/account/signIn")
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != 200 {
		log.Fatalf("FAILED - Could not get login page. StatusCode: %d", resp.StatusCode)
	}
	formdata := ParseLogin(html.NewTokenizer(resp.Body))
	data := url.Values{}
	for k, v := range formdata {
		data.Set(k, v)
	}
	data.Set("email", webuser)
	data.Set("password", webpasswd)
	data.Set("submitButton", "Submit")
	resp, err = client.PostForm("https://www.hardenize.com/account/signIn", data)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != 200 {
		log.Fatalf("FAILED - Logging in. StatusCode: %d", resp.StatusCode)
	}

	// Done
	hc := hardenizeclient{
		apiuser:	apiuser,
		apipasswd:	apipasswd,
		webclient:	client,
	}
	return &hc
}

func (hc *hardenizeclient) GetAPIData(url string) []byte {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	req.SetBasicAuth(hc.apiuser, hc.apipasswd)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode != 200 {
		fmt.Printf("HTTP Failure %d %s\n", resp.StatusCode, resp.Status)
		panic(fmt.Errorf("FAILED - Getting %s", url))
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	return body
}

func (hc *hardenizeclient) GetWebPage(url string) *html.Tokenizer {
	resp, err := hc.webclient.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != 200 {
		log.Fatalf("FAILED - Getting %s", url)
	}

	return html.NewTokenizer(resp.Body)
}

func (hc *hardenizeclient) GetCSV(url string) [][]string {
	resp, err := hc.webclient.Get(url)
	if err != nil {
		log.Fatalf("Web client failed: %s", err)
	}

	if resp.StatusCode != 200 {
		log.Fatalf("FAILED - Getting %s", url)
	}

	r := csv.NewReader(resp.Body)
	records, err := r.ReadAll()
	if err != nil {
		log.Fatal("CSV error: ", err)
	}
	return records
}

func ParseLogin(htmltokens *html.Tokenizer) map[string]string {
	result := make(map[string]string)

P:
	for {
		tt := htmltokens.Next()
		t := htmltokens.Token()
		isInput := false
		switch tt {
		case html.ErrorToken:
			// End of the document, we're done
			break P
		case html.StartTagToken:
			isInput = t.Data == "input"
		case html.SelfClosingTagToken:
			isInput = t.Data == "input"
		}
		if isInput {
			var inputtype string
			var name string
			var value string
			for _, attr := range t.Attr {
				if attr.Key == "type" && attr.Val == "hidden" {
					inputtype = attr.Val
				}
				if attr.Key == "name" {
					name = attr.Val
				}
				if attr.Key == "value" {
					value = attr.Val
				}
			}
			if inputtype == "hidden" {
				result[name] = value
			}
		}
	}
	return result
}
