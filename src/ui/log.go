package TriUI

/* Access Logging
 *
 * This logs, in JSON format to the selected access.log
 *
 * When a SIGUSR1 is received the log file is closed and re-opened
 * to support logrotate which moves the file out the way.
 *
 * Note that we have a mutex protecting la_running, but the actual
 * logging happens in a separate go thread so that there is no delay
 * while writing entries to the log (disks are slow).
 */

import (
	"encoding/json"
	"os"
	"os/signal"
	"sync"
	"syscall"
	tr "lib"
)

var la_chan chan string
var la_running bool
var la_done chan bool
var la_file *os.File = nil
var la_mutex sync.Mutex

func la_open() (err error) {
	/* Close any old open ones */
	la_close()

	if tr.Config.LogFile == "" {
		tr.Logf("No log file configured, skipping access logging")
		return
	}

	tr.Dbgf("Opening log file %q", tr.Config.LogFile)
	la_file, err = os.OpenFile(tr.Config.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	return
}

func la_close() {
	if la_file != nil {
		tr.Dbgf("Closing log file")
		la_file.Close()
		la_file = nil
	}
}

func la_write(txt string) {
	/* Logging disabled */
	if la_file == nil {
		return
	}

	_, err := la_file.WriteString(txt + "\n")
	if err != nil {
		tr.Errf("LogAccess() writing to %s failed: %s", tr.Config.LogFile, err.Error())

		/* Try to re-open access log file */
		la_open()
	}
}

func la_rtn() {
	la_mutex.Lock()
	la_running = true
	la_mutex.Unlock()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGUSR1)

	running := true

	for running {
		select {
		case txt, ok := <-la_chan:
			if !ok {
				running = false
				break
			}

			/* Write the log entry */
			la_write(txt)
			break

		case s := <-sigChan:
			if s == syscall.SIGUSR1 {
				tr.Dbgf("Received SIGUSR1, acting upon: rotating log file")
				la_close()
				la_open()
			}
			break
		}
	}

	/* Close the log file */
	la_close()

	/* Tell them we are done */
	la_done <- true

	la_mutex.Lock()
	la_running = false
	la_mutex.Unlock()
}

/* Extends PfUIS */
func (cui *PfUIS) logaccess() {
	/* No LogFile -> Nothing to do */
	if tr.Config.LogFile == "" {
		return
	}

	/* Log the access */
	username := ""
	theuser := cui.TheUser()
	if theuser != nil {
		username = theuser.GetUserName()
	}

	type la_item struct {
		Username    string `json:"username"`
		Nodename    string `json:"nodename"`
		IP          string `json:"ip"`
		XFF         string `json:"xff"`
		HTTP_Method string `json:"method"`
		HTTP_Host   string `json:"host"`
		HTTP_Path   string `json:"path"`
		HTTP_Args   string `json:"args"`
		Template    string `json:"template"`
		StaticFile  string `json:"staticfile"`
	}

	la := la_item{
		Username:    username,
		Nodename:    tr.Config.Nodename,
		IP:          cui.GetClientIP().String(),
		XFF:         cui.GetRemote(),
		HTTP_Method: cui.GetMethod(),
		HTTP_Host:   cui.GetHTTPHost(),
		HTTP_Path:   cui.GetFullPath(),
		HTTP_Args:   cui.r.URL.RawQuery,
		Template:    cui.show_name,
		StaticFile:  cui.staticfile,
	}

	txt, err := json.Marshal(la)
	if err != nil {
		cui.Errf("Could not format access log message: %s", err.Error())
		return
	}

	direct := false

	la_mutex.Lock()
	if la_running {
		direct = true
	}
	la_mutex.Unlock()

	if !direct {
		la_chan <- string(txt)
	} else {
		la_write(string(txt))
	}
}

func LogAccess_start() (err error) {
	la_chan = make(chan string, 1000)
	la_done = make(chan bool)

	/* Open the file at start, so we can detect initial errors */
	err = la_open()
	if err != nil {
		return
	}

	/* Start background logging process */
	go la_rtn()

	/* All dandy */
	return
}

func LogAccess_stop() {
	la_mutex.Lock()
	defer la_mutex.Unlock()

	if !la_running {
		return
	}

	/* Close the channel */
	close(la_chan)

	/* Wait for it to finish */
	<-la_done
}
