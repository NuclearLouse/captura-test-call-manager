// structs_netsense_db.go
//
// The file contains the structures that are needed to work with "Netsense" tables.
//
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

func (netSenseAPI) TableName() string {
	return schemaPG + "CallingSys_API_NetSense"
}

type netSenseAPI struct {
	SystemName        string `gorm:"size:50;foreignkey:CallingSys_Settings.SystemName"`
	URL               string `gorm:"size:100"`
	User              string `gorm:"size:100"`
	Pass              string `gorm:"size:100"`
	TestInit          int    `gorm:"type:int"`
	TestInitCli       int    `gorm:"type:int"`
	TestStatus        int    `gorm:"type:int"`
	TestStatusDetails int    `gorm:"type:int"`
	SystemID          int    `gorm:"type:int"`
}
