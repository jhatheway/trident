///usr/bin/env go run -ldflags "-X main.Version=shell.run" $0 "$@"; exit
/* Trident Pitchfork Server */
package main

import (
	"flag"
	"fmt"
	"log"
	"log/syslog"
	"net/http"
	"strings"

	/* Pitchfork Libraries */
	tr "lib"
	tu "ui"
)

var Version = "unconfigured"
func Serve(dname string, appname string, version string, copyright string, website string, app_schema_version int, newui tu.PfNewUI, starthook func()) {
	var err error
	var use_syslog bool
	var disablestamps bool
	var insecurecookies bool
	var disabletwofactor bool
	var loglocation bool
	var verbosedb bool
	var confroot string
	var daemonize bool
	var pidfile string
	var username string
	var debug bool
	var showversion bool

	ldname := strings.ToLower(dname)

	tr.SetAppDetails(appname, version, copyright, website)

	flag.StringVar(&confroot, "config", "", "Configuration File Directory")
	flag.BoolVar(&use_syslog, "syslog", false, "Log to syslog")
	flag.BoolVar(&disablestamps, "disablestamps", false, "Disable timestamps in logs")
	flag.BoolVar(&insecurecookies, "insecurecookies", false, "Insecure Cookies (for testing directly against the daemon instead of going through nginx/apache")
	flag.BoolVar(&disabletwofactor, "disabletwofactor", false, "Disable Two Factor Authentication Check (development only)")
	flag.BoolVar(&loglocation, "loglocation", false, "Log Code location in log messages")
	flag.BoolVar(&verbosedb, "verbosedb", false, "Verbose DB output (Query Logging)")
	flag.BoolVar(&daemonize, "daemonize", false, "Daemonize")
	flag.StringVar(&pidfile, "pidfile", "", "PID File (useful in combo with daemonize)")
	flag.StringVar(&username, "username", "", "Change to user")
	flag.BoolVar(&debug, "debug", false, "Enable Debug output")
	flag.BoolVar(&showversion, "version", false, "Show version")

	flag.Parse()

	if showversion {
		fmt.Print(tr.VersionText())
		return
	}

	if daemonize {
		/* Part of this won't return */
		tr.Daemon(0, 0)

		/* Mandatory */
		use_syslog = true
	}

	if use_syslog {
		logwriter, e := syslog.New(syslog.LOG_NOTICE, ldname)
		if e != nil {
			fmt.Printf("Could not open syslog: %s", err.Error())
			return
		}

		/* Output to syslog */
		log.SetOutput(logwriter)

		/* Disable the timestamp, syslog takes care of that */
		disablestamps = true
	}

	/* Disable timestamps in the log? */
	if disablestamps {
		flags := log.Flags()
		flags &^= log.Ldate
		flags &^= log.Ltime
		log.SetFlags(flags)
	}

	/* Store the PID */
	pid := tr.GetPID()
	if pidfile != "" {
		tr.StorePID(pidfile, pid)
	}

	/* Drop privileges */
	if username != "" {
		err = tr.SetUID(username)
	}

	tr.CheckTwoFactor = !disabletwofactor
	tr.LogLocation = loglocation
	tr.Debug = debug

	tr.Logf("%s Daemon %s (%s) starting up", appname, dname, tr.AppVersionStr())

	/* Setup lib */
	err = tr.Setup(ldname, confroot, verbosedb, app_schema_version)
	if err != nil {
		return
	}

	/* Setup UI */
	err = tu.Setup(ldname, !insecurecookies)
	if err != nil {
		return
	}

	/* Everything goes through the UI root */
	r := tu.NewPfRootUI(newui)
	http.HandleFunc("/", r.H_root)

	/* Notify that we are ready */
	tr.Logf("%s is running on node %s", ldname, tr.Config.Nodename)

	tr.Starts()
	defer tr.Stops()

	/* Call Starthook in background goroutine when provided */
	if starthook != nil {
		go starthook()
	}

	/* Tell what HTTP port we are serving on */
	tr.Logf("%s serving on %s port %s", ldname, tr.Config.Http_host, tr.Config.Http_port)

	/* Listen and Serve the HTTP interface */
	err = http.ListenAndServe(tr.Config.Http_host+":"+tr.Config.Http_port, nil)
	if err != nil {
		log.Fatal(err)
		return
	}

	tr.Log("done")
	return
}

func main() {
	Serve("trident", "Trident", Version, "", "", tr.AppSchemaVersion, tu.NewTriUI, nil)
}



