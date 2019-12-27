package main

type destinationList struct {
	destination string
	code        string
	alias       string
}

type routeList struct {
	externalID  string // External ID
	folder      string // Folder (OPTIONAL)
	carrier     string // Carrier (OPTIONAL)
	trunkGroup  string // Trunk Group (OPTIONAL)
	trunk       string // Trunk Name /trunk string
	description string // Description (OPTIONAL)
	dialerGroup string // Dialer Group (OPTIONAL)
	iac         string // IAC (OPTIONAL)
	nac         string // NAC (OPTIONAL)
}
