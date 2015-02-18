// Copyright 2015 Paul Simpson. All rights reserved.

package main

import (
	"flag"
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/pjs/command/timeserver/auth"
	"html"
	"html/template"
	"net/http"
	"os"
	"time"
)

type Page struct {
	Body template.HTML
}

var templateDir string
var logins *auth.AuthMap

// Handler to display the current time
func HandleTime(w http.ResponseWriter, req *http.Request) {
	t := time.Now()
	ut := time.Now().UTC()
	const layout = "3:04:05 PM"
	const layout2 = "3:04:05 UTC"
	timeNow := t.Format(layout)
	timeNowUTC := ut.Format(layout2)

	message := "The time is now <span class=\"time\">" + timeNow + "</span> (" + timeNowUTC + ")"

	// if they are logged in, personalize the time
	name, logged := logins.IsLoggedIn(req)

	if logged {
		message = message + ", " + name + "."
		log.Info("Time was requested at /login for user " + name)
	} else {
		log.Info("Time was requested at /login, no user logged in")
	}

	c := template.HTML(message)
	p := &Page{Body: c}
	renderTemplate(w, p)
}

// Handler to set status code to 404 and display custom message
func Handle404(w http.ResponseWriter, req *http.Request) {
	w.Header()["content-type"] = []string{"text/html"}
	http.Error(w, "", http.StatusNotFound)

	message := "These are not the URLs you're looking for."
	c := template.HTML(message)
	p := &Page{Body: c}
	renderTemplate(w, p)

	log.Info("404 page was request at path: " + html.EscapeString(req.URL.Path))
}

// Handler for people to login with their name
func HandleLogin(w http.ResponseWriter, req *http.Request) {
	name := req.URL.Query().Get("name")

	if len(name) != 0 {
		log.Info(name + " logged in to the timeserver")

		logins.Login(w, name)

		http.Redirect(w, req, "/", 302)

	} else {
		log.Warn("Someone attempted to login without a name!")
		c := template.HTML("C'mon, I need a name.")
		p := &Page{Body: c}
		renderTemplate(w, p)
	}
}

// displays login page or greets the user
func HandleIndex(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		if req.URL.Path != "/index.html" {
			Handle404(w, req)
			return
		}
	}

	name, logged := logins.IsLoggedIn(req)

	if logged {
		log.Info("Logged in user viewed the index: " + name)
		c := template.HTML("Greetings, " + name + ".")
		p := &Page{Body: c}
		renderTemplate(w, p)
	} else {
		log.Info("Anonymous user visited the index")
		DisplayLogin(w, req)
	}
}

// login page
func DisplayLogin(w http.ResponseWriter, req *http.Request) {
	p := &Page{Body: "<form action=\"login\">" +
		"What is your name, Earthling?" +
		"<input type=\"text\" name=\"name\" size=\"50\">" +
		"<input type=\"submit\">" +
		"</form>"}

	renderTemplate(w, p)
}

// logs out the user (writes garbage to cookie)
func HandleLogout(w http.ResponseWriter, req *http.Request) {
	logins.Logout(w)

	p := &Page{Body: "<META http-equiv=\"refresh\" content=\"10;URL=/\">Good-bye."}
	renderTemplate(w, p)

	log.Info("User logged out")
}

// Page struct gets rendered into the templated and sent out
func renderTemplate(w http.ResponseWriter, p *Page) {
	p.Body = template.HTML(p.Body)
	t, _ := template.ParseFiles(templateDir+"base.html", templateDir+"../templates/nav.html")
	t.Execute(w, p)
}

// loads in log configuration from the file specified
func configureLogger(name string) {
	logger, err := log.LoggerFromConfigAsFile(name)

	if err != nil {
		log.Error("Could not configure logger")
	} else {
		log.ReplaceLogger(logger)

		log.Info("Logger has been configured")
	}
}

// Launches a web server thats main purpose is to display the time
// e.g. http://localhost:8080/time will display the messsage:
//
//        "The time is now 2:34:03 PM"
//
func main() {
	port := flag.String("port", "8080", "server port number")
	version := flag.Bool("V", false, "server version")
	template := flag.String("templates", "../templates/", "templates directory")
	//logFile := flag.String("log", "../etc/log.cfg", "log configuration file")
	authHost := flag.String("authhost", "127.0.0.1", "hostname of authserver")
	authPort := flag.String("authport", "8888", "port of authserver")
	timeout := flag.Int("authtimeout-ms", 5, "auth server timeout")
	flag.Parse()
	templateDir = *template

	// if flag set, display version then exit
	if *version {
		fmt.Println("timeserver version 4.0 css490/assignment4")
		os.Exit(0)
	}

	durationTimeout := time.Duration(*timeout)*time.Millisecond

	logins = auth.New(*authHost+":"+*authPort, durationTimeout)

	//configureLogger(*logFile)

	// start up the server
	http.HandleFunc("/", HandleIndex)
	http.HandleFunc("/index.html", HandleIndex)
	http.HandleFunc("/time", HandleTime)
	http.HandleFunc("/login", HandleLogin)
	http.HandleFunc("/logout", HandleLogout)
	err := http.ListenAndServe(":"+*port, nil)
	log.Info("Listening on port " + *port + "...")

	// return error and exit code 1 if port in use
	if err != nil {
		log.Error("Server port in use")
		os.Exit(1)
	}
}
