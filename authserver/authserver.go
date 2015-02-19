// Copyright 2015 Paul Simpson. All rights reserved.

package main

import (
	"flag"
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/pjs/cmap"
	"net/http"
	"os"
	"reflect"
	"time"
	"io/ioutil"
	"encoding/json"
	"os/exec"
	"bytes"
)

var m *cmap.ConcurrentMap
var port *string
var dump *string
var checkpoint *int
var dumpFile string

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

func dumpProcessing() {

	// check if we need to load up the dump for the first time
	if _, err := os.Stat(dumpFile); err == nil {
		log.Info("Dump file exists! Loading in...")
		dataMap := make(map[string]string)
		read(dataMap)
		m.LoadInNewMap(dataMap)
	} else {
		log.Info("Dump file does not exist.")
	}
	go dumpIntervalChecking()
}

func dumpIntervalChecking() {
	for {
		time.Sleep(time.Duration(*checkpoint) * time.Second)
		copyFile(dumpFile, dumpFile+".bak")

		// copy current map and write to file
		dataMap := make(map[string]string)
		m.CopyCurrentMap(dataMap)
		write(dataMap)

		// read back and verify
		success := verify(dataMap)

		if success {
			log.Info("Successful verification, deleting backup...")
			err := os.Remove(dumpFile + ".bak")
			if err != nil {
				log.Error("Problem deleting backup!")
			}
		} else {
			log.Error("Failure verifying backup file")
		}

	}
}

func verify(dataMap map[string]string) bool {
	verifyMap := make(map[string]string)
	read(verifyMap)

	match := reflect.DeepEqual(verifyMap, dataMap)
	if match {
		return true
	} else {
		return false
	}
}

func read(dataMap map[string]string) {
	dump, _ := ioutil.ReadFile(dumpFile)
	dec := json.NewDecoder(bytes.NewReader(dump))
	dec.Decode(&dataMap)
}

func write(dataMap map[string]string) {
	data, err := json.Marshal(dataMap)

	if err != nil {
		log.Info("Problem with json writing to file")
		return
	}

	err2 := ioutil.WriteFile(dumpFile, data, 0644)
	if err2 != nil {
		log.Error("Problem writing dump file")
	}
}

func copyFile(src string, dest string) {
	cmd := exec.Command("cp", src, dest)
	err := cmd.Run()
	if err != nil {
		log.Error("Problem copying .bak file")
	}
}

func main() {
	port = flag.String("authport", "8888", "server port number")
	dump = flag.String("dumpfile", "not specified", "login state dump")
	checkpoint = flag.Int("checkpoint-interval", 5, "rate which to backup state")
	flag.Parse()
	dumpFile = *dump

	m = cmap.New()

	// crude way of checking to see if the flag was present
	// using default value in flag, if not, completely ignore dumpfile stuff
	if dumpFile != "not specified" {
		dumpProcessing()
	}

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
