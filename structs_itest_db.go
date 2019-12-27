// structs_itest_db.go
//
// The file contains the structures that are needed to work with "iTest" tables.
//
package main

import (
	"os"
)

func (itestAPI) TableName() string {
	return os.Getenv("SCHEMA_PG") + "CallingSys_API_iTest"
}

type itestAPI struct {
	SystemName        string `gorm:"size:50;foreignkey:CallingSys_Settings.SystemName"`
	URL               string `gorm:"size:100"`
	RepoURL           string `gorm:"size:100"`
	User              string `gorm:"size:100"`
	Pass              string `gorm:"size:100"`
	Profiles          int    `gorm:"type:int"`
	Suppliers         int    `gorm:"type:int"`
	NdbStd            int    `gorm:"type:int"`
	NdbCli            int    `gorm:"type:int"`
	TestInit          int    `gorm:"type:int"`
	TestInitCli       int    `gorm:"type:int"`
	TestStatus        int    `gorm:"type:int"`
	TestStatusDetails int    `gorm:"type:int"`
	SystemID          int    `gorm:"type:int"`
}

func (itestProfiles) TableName() string {
	return os.Getenv("SCHEMA_PG") + "CallingSys_iTest_profiles"
}

type itestProfiles struct {
	ProfileID        string `gorm:"column:profile_id;size:100"`
	ProfileName      string `gorm:"column:profile_name;size:100"`
	ProfileIP        string `gorm:"column:profile_ip;size:100"`
	ProfilePort      string `gorm:"column:profile_port;size:100"`
	ProfileSrcNumber string `gorm:"column:profile_src_number;size:100"`
}

func (itestSuppliers) TableName() string {
	return os.Getenv("SCHEMA_PG") + "CallingSys_iTest_suppliers"
}

type itestSuppliers struct {
	SupplierID   string `gorm:"column:supplier_id;size:100"`
	SupplierName string `gorm:"column:supplier_name;size:100"`
	Prefix       string `gorm:"column:prefix;size:100"`
	Codec        string `gorm:"column:codec;size:100"`
}

func (b itestBreakouts) TableName() string {
	var name string
	switch b.BreakType {
	case "cli":
		name = os.Getenv("SCHEMA_PG") + "CallingSys_iTest_breakouts_cli"
	case "std":
		name = os.Getenv("SCHEMA_PG") + "CallingSys_iTest_breakouts_std"
	}
	return name
}

type itestBreakouts struct {
	CountryName  string `gorm:"column:country_name;size:100"`
	CountryID    string `gorm:"column:country_id;size:100"`
	BreakoutName string `gorm:"column:breakout_name;size:100"`
	BreakoutID   string `gorm:"column:breakout_id;size:100"`
	BreakType    string `gorm:"-"`
}
