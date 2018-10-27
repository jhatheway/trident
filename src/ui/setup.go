package TriUI

/*
 * Trident Pitchfork UI Setup
  *
   * Split out so that we can call it for Tests cases too next to normal server behaviour
*/

import (
	tr "lib"
)

func Setup(toolname string, securecookies bool) (err error) {
	/* Initialize UI Settings */
	err = UIInit(securecookies, "_"+toolname)
	if err != nil {
		tr.Errf("UI Init failed: %s", err.Error())
		return
	}

	/* Load Templates */
	err = tr.Template_Load()
	if err != nil {
		tr.Errf("Template Loading failed: %s", err.Error())
		return
	}

	/* Start Access Logger */
	if tr.Config.LogFile != "" {
		err = LogAccess_start()
		if err != nil {
			tr.Errf("Could not open log file (%s): %s", tr.Config.LogFile, err.Error())
			return
		}
		defer LogAccess_stop()
	} else {
		tr.Logf("Note: Access LogFile disabled")
	}

	return
}
