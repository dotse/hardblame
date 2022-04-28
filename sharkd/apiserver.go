/*
 * apiserver.go
 *
 * Johan Stenstam, johan.stenstam@internetstiftelsen.se
 */
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"github.com/dotse/hardblame/lib"
)

func homeLink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome home!")
}

// func API_NYI(conf *Config) func(w http.ResponseWriter, r *http.Request) {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		fmt.Fprintf(w, "NYI")
// 
// 		status := 101
// 		resp := "NYI"
// 
// 		apistatus := music.APIstatus{Status: status, Message: resp}
// 		w.Header().Set("Content-Type", "application/json")
// 		err := json.NewEncoder(w).Encode(apistatus)
// 		if err != nil {
// 			log.Printf("Error from Encoder: %v\n", err)
// 		}
// 	}
// }
 
func APIGoAway(conf *Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		status := 404
		resp := "These are not the droids you're looking for"

		apistatus := APIstatus{Status: status, Message: resp}
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(apistatus)
		if err != nil {
			log.Printf("Error from Encoder: %v\n", err)
		}
	}
}

var pongs int = 0

func APIping(conf *Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		log.Printf("APIping: received /ping request from %s.\n", r.RemoteAddr)

		tls := ""
		if r.TLS != nil {
			tls = "TLS "
		}

		decoder := json.NewDecoder(r.Body)
		var pp sharklib.PingPost
		err := decoder.Decode(&pp)
		if err != nil {
			log.Println("APIping: error decoding ping post:", err)
		}

		pongs += 1

		host, _ := os.Hostname()
		response := sharklib.PingResponse{
			Time:    time.Now(),
			Client:  r.RemoteAddr,
			Message: fmt.Sprintf("%spong from sharkd @ %s", tls, host),
			Pings:   pp.Pings + 1,
			Pongs:   pongs}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			log.Printf("Error from Encoder: %v\n", err)
		}
	}
}

func APIhardenize(conf *Config) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		decoder := json.NewDecoder(r.Body)
		var hp sharklib.HardenizePost
		err := decoder.Decode(&hp)
		if err != nil {
			log.Println("APIhardenize: error decoding hardenize post:", err)
		}

		log.Printf("APIhardenize: received /hardenize %s command from %s.\n",
					  hp.Command, r.RemoteAddr)
		var msg string
		
		switch hp.Command {
		case "fetch":
		     err, msg = FetchAndProcess(conf)

		default:
		     err = fmt.Errorf("No such hardenize command: %s", hp.Command)
		     msg = "Bummer"
		}

		response := sharklib.HardenizeResponse{
			Time:    time.Now(),
			Error:	 err,
			Message: msg,
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			log.Printf("Error from Encoder: %v\n", err)
		}
	}
}

func APIgroup(conf *Config) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		decoder := json.NewDecoder(r.Body)
		var gp sharklib.GroupPost
		err := decoder.Decode(&gp)
		if err != nil {
			log.Println("APIgroup: error decoding group post:", err)
		}

		log.Printf("APIgroup: received /group %s command from %s.\n",
					  gp.Command, r.RemoteAddr)
		var msg string
		var gc map[string]int
		
		switch gp.Command {
		case "count":
		     err, gc = conf.HardDB.GroupCount()
		     msg = "Tjolahopp"

		default:
		     err = fmt.Errorf("No such hardenize command: %s", gp.Command)
		     msg = "Bummer"
		}

		response := sharklib.GroupResponse{
			Time:    time.Now(),
			Counts:	 gc,
			Error:	 err,
			Message: msg,
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			log.Printf("Error from Encoder: %v\n", err)
		}
	}
}


func SetupRouter(conf *Config) *mux.Router {
	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/", homeLink)

	sr := r.PathPrefix("/api/v1").Headers("X-API-Key",
		viper.GetString("apiserver.apikey")).Subrouter()
	sr.HandleFunc("/ping", APIping(conf)).Methods("POST")
	sr.HandleFunc("/hardenize", APIhardenize(conf)).Methods("POST")
	sr.HandleFunc("/group", APIgroup(conf)).Methods("POST")

	return r
}

func walkRoutes(router *mux.Router, address string) {
	log.Printf("Defined API endpoints for router on: %s\n", address)

	walker := func(route *mux.Route, router *mux.Router,
		ancestors []*mux.Route) error {
		path, _ := route.GetPathTemplate()
		methods, _ := route.GetMethods()
		for m := range methods {
			log.Printf("%-6s %s\n", methods[m], path)
		}
		return nil
	}
	if err := router.Walk(walker); err != nil {
		log.Panicf("Logging err: %s\n", err.Error())
	}
	//    return nil
}

// In practice APIdispatcher doesn't need a termination signal, as it will
// just sit inside http.ListenAndServe, but we keep it  for  symmetry.
//
// func APIdispatcher(conf *Config, done <-chan struct{}) {
func APIdispatcher(conf *Config) error {
	router := SetupRouter(conf)
	address := viper.GetString("apiserver.address")
	certFile := viper.GetString("apiserver.certFile")
	keyFile := viper.GetString("apiserver.keyFile")

	if address != "" {
		log.Println("Starting API dispatcher. Listening on", address)
		log.Fatal(http.ListenAndServeTLS(address, certFile, keyFile, router))
	}

	log.Println("API dispatcher: unclear how to stop the http server nicely.")
	return nil
}
