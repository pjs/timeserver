// Copyright 2015 Paul Simpson. All rights reserved.

package auth

import (
	log "github.com/cihub/seelog"
	"net/http"
	"os/exec"
	"strings"
	"time"
	"io/ioutil"
)

type AuthMap struct {
	server string
	timeout time.Duration
}

// constructor
func New(s string, t time.Duration) *AuthMap {
	return &AuthMap{
		server:  s,
		timeout: t,
	}
}

// for people to login with their name
func (a *AuthMap) Login(w http.ResponseWriter, name string) {

	id, _ := exec.Command("uuidgen").Output()
	sid := strings.Trim(string(id), "\n")

	c := &http.Cookie{Name: "login", Value: sid}

	client := http.Client{
		Timeout: a.timeout,
	}

	url := "http://" + a.server + "/set?cookie=" + sid + "&name=" + name
	resp, err := client.Get(url)
	if err != nil {
		log.Error("Error contacting authserver")
		return
	}

	code := resp.StatusCode
	log.Debug("Status code set in response: " + string(code))
	if code == 200 {
		log.Info("Login sucessful")
		http.SetCookie(w, c)
	} else {
		log.Info("Login failure")
		a.Logout(w)
	}

}

// returns name of user if they are logged in
func (a *AuthMap) IsLoggedIn(req *http.Request) (string, bool) {
	cookie, err := req.Cookie("login")
	if err != nil {
		return "", false
	}

	client := http.Client{
		Timeout: a.timeout,
	}
	url := "http://" + a.server + "/get?cookie=" + cookie.Value
	resp, err := client.Get(url)
	log.Info("Sent request to login with cookie: " + cookie.Value)
	if err != nil {
		log.Error("Error contacting authserver")
		return "", false
	}

	code := resp.StatusCode
	if code == 400 {
		log.Error("Authserver request was malformed")
		return "", false
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	name := string(body)

	if name == "" {
		log.Info("Empty string was returned, no user logged in")
		return "", false
	} else {
		log.Info("Logged in user found: " + name)
		return name, true
	}
}

// logs out the user (writes garbage to cookie)
func (a *AuthMap) Logout(w http.ResponseWriter) {
	c := &http.Cookie{Name: "login", Value: "0"}
	http.SetCookie(w, c)
}
