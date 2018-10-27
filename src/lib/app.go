package trident

/* The version of the Database */
var AppSchemaVersion = 0

/* Version of AppSetup */
var AppSetupVersion = 0

var AppName = "Trident"
var AppVersion = "unconfigured"
var AppWebsite = "https:///tridentli"
var AppCopyright = ""

func SetAppDetails(name string, ver string, copyright string, website string) {
	AppName = name
	AppVersion = ver
	AppWebsite = website
	AppCopyright = copyright
}

func AppVersionStr() string {
	return AppVersion
}

func VersionText() string {
	t := AppName + "\n" +
		"Version: " + AppVersion + "\n"

	if AppWebsite != "" {
		t += "Website: " + AppWebsite + "\n"
	}

	if AppCopyright != "" {
		t += "Copyright: " + AppCopyright + "\n"
	}

	t += "\n" +
		"Using Trident Pitchfork\n" +
		"Copyright: (C) 2015-2017 The Trident Project\n" +
		"           Portions (C) 2015 National Cyber Forensics Training Alliance\n" +
		"Website: https://trident.li\n"

	return t
}
