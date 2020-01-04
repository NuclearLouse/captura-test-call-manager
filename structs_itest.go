// structs_itest.go
//
// The file contains the structures that are needed
// to process XML responses from the "iTest" system
// and contains the structures that are needed to work with "iTest" tables.
package main

import (
	"encoding/xml"
	"os"
)

type TestInitiation struct {
	XMLName xml.Name `xml:"Test_Initiation"`
	Test    test     `xml:"Test"`
}

type test struct {
	TestID   string `xml:"Test_ID"`
	ShareURL string `xml:"Share_URL"`
}

type ListNDB struct {
	XMLName   xml.Name   `xml:"NDB_List"`
	Breakouts []breakout `xml:"Breakout"`
}

type breakout struct {
	XMLName      xml.Name `xml:"Breakout"`
	CountryName  string   `xml:"Country_Name"`
	CountryID    string   `xml:"Country_ID"`
	BreakoutName string   `xml:"Breakout_Name"`
	BreakoutID   string   `xml:"Breakout_ID"`
}

type ProfilesList struct {
	XMLName  xml.Name  `xml:"Profiles_List"`
	Profiles []profile `xml:"Profile"`
}

type profile struct {
	XMLName          xml.Name `xml:"Profile"`
	ProfileID        string   `xml:"Profile_ID"`
	ProfileName      string   `xml:"Profile_Name"`
	ProfileIP        string   `xml:"Profile_IP"`
	ProfilePort      string   `xml:"Profile_Port"`
	ProfileSrcNumber string   `xml:"Profile_Src_Number"`
}

type SuppliersList struct {
	XMLName   xml.Name   `xml:"Vendors_List"`
	Suppliers []supplier `xml:"Supplier"`
}

type supplier struct {
	XMLName      xml.Name `xml:"Supplier"`
	SupplierID   string   `xml:"Supplier_ID"`
	SupplierName string   `xml:"Supplier_Name"`
	Prefix       string   `xml:"Prefix"`
	Codec        string   `xml:"Codec"`
}

type CallsInfo struct {
	XMLName      xml.Name     `xml:"Test_Status"`
	TestOverview testOverview `xml:"Test_Overview"`
	Calls        []call       `xml:"Call"`
}

type testOverview struct {
	XMLName  xml.Name `xml:"Test_Overview"`
	Name     string   `xml:"Name"`
	Supplier string   `xml:"Supplier"`
	InitBy   string   `xml:"InitBy"`
	Init     int64    `xml:"Init"`
	Type     string   `xml:"Type"`
	TestID   string   `xml:"Test_ID"`
}

type call struct {
	XMLName     xml.Name `xml:"Call"`
	CallID      string   `xml:"ID"`
	Source      string   `xml:"Source"`
	Destination string   `xml:"Destination"`
	Start       int64    `xml:"Start"`
	End         float64  `xml:"End"`
	PDD         float64  `xml:"PDD"`
	MOS         float64  `xml:"MOS"`
	Ring        string   `xml:"Ring"`
	Call        string   `xml:"Call"`
	LastCode    string   `xml:"Last_Code"`
	Result      string   `xml:"Result"`
	ResultCode  int      `xml:"Result_Code"`
	CLI         string   `xml:"CLI"`
	FAS         string   `xml:"FAS"`
	LdFAS       string   `xml:"LD_FAS"`
	DeadAir     string   `xml:"Dead_Air"`
	NoRBT       string   `xml:"No_RBT"`
	Viber       string   `xml:"Viber"`
	FDLR        string   `xml:"F-DLR"`
}

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
