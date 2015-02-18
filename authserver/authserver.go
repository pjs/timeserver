// Copyright 2015 Paul Simpson. All rights reserved.

package main

import (
	"flag"
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/pjs/cmap"
	"net/http"
	"os"
)

var m *cmap.ConcurrentMap

func getHandler(w http.ResponseWriter, r *http.Request) {
	uuid := r.FormValue("cookie")
	log.Info("Received /get with uuid: " + uuid)

	if uuid != "" {
		w.WriteHeader(200)
		name, found := m.Get(uuid)
		if found {
			fmt.Fprint(w, name)
			log.Info("Successfully found name: " + name)
		} else {
			fmt.Fprint(w, "")
			log.Info("No associated name, return empty string")
		}
	} else {
		log.Info("Bad request")
		w.WriteHeader(400)
	}
}

func setHandler(w http.ResponseWriter, r *http.Request) {
	cookie := r.FormValue("cookie")
	name := r.FormValue("name")
	log.Info("Received /set with cookie: " + cookie + ", name: " + name)

	if cookie != "" && name != "" {
		m.Set(cookie, name)
		w.WriteHeader(200)
		log.Info("Cookie has been set... cookie: " + cookie + ", name: " + name)
	} else {
		log.Info("Cookie not set!")
		w.WriteHeader(400)
	}

}

func main() {
	port := flag.String("authport", "8888", "server port number")
	flag.Parse()

	m = cmap.New()

	http.HandleFunc("/get", getHandler)
	http.HandleFunc("/set", setHandler)
	err := http.ListenAndServe(":"+*port, nil)
	log.Info("Listening on port " + *port + "...")

	// return error and exit code 1 if port in use
	if err != nil {
		log.Error("Server port in use")
		os.Exit(1)
	}
}
