// structs_netsense_db.go
//
// The file contains the structures that are needed to work with "Netsense" tables.
//
package main

import "os"

func (netSenseAPI) TableName() string {
	return os.Getenv("SCHEMA_PG") + "CallingSys_API_NetSense"
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
