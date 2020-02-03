// structs_itest.go
//
// The file contains the structures that are needed
// to process XML responses from the "iTest" system
// and contains the structures that are needed to work with "iTest" tables.
package main

import (
	"encoding/xml"
)

func (itestAPI) TableName() string {
	return schemaPG + "CallingSys_API_iTest"
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

type testInitItest struct {
	XMLName xml.Name `xml:"Test_Initiation"`
	Test    struct {
		TestID   string `xml:"Test_ID"`
		ShareURL string `xml:"Share_URL"`
	} `xml:"Test"`
}

type testStatusItest struct {
	XMLName       xml.Name `xml:"Test_Status"`
	Name          string   `xml:"Name"`
	CallsTotal    int      `xml:"Calls_Total"`
	CallsComplete int      `xml:"Calls_Complete"`
	CallsSuccess  int      `xml:"Calls_Success"`
	CallsNoAnswer int      `xml:"Calls_No_Answer"`
	CallsFail     int      `xml:"Calls_Fail"`
	PDD           float64  `xml:"PDD"`
	ShareURL      string   `xml:"Share_URL"`
}

type testResultItest struct {
	XMLName      xml.Name `xml:"Test_Status"`
	TestOverview struct {
		Name     string `xml:"Name"`
		Supplier string `xml:"Supplier"`
		InitBy   string `xml:"InitBy"`
		Init     int64  `xml:"Init"`
		Type     string `xml:"Type"`
		TestID   string `xml:"Test_ID"`
	} `xml:"Test_Overview"`
	Call []struct {
		ID          string  `xml:"ID"`
		Source      string  `xml:"Source"`
		Destination string  `xml:"Destination"`
		Start       int64   `xml:"Start"`
		End         float64 `xml:"End"`
		PDD         float64 `xml:"PDD"`
		MOS         float64 `xml:"MOS"`
		Ring        string  `xml:"Ring"`
		Call        string  `xml:"Call"`
		LastCode    string  `xml:"Last_Code"`
		Result      string  `xml:"Result"`
		ResultCode  int     `xml:"Result_Code"`
		CLI         string  `xml:"CLI"`
		FAS         string  `xml:"FAS"`
		LDFAS       string  `xml:"LD_FAS"`
		DeadAir     string  `xml:"Dead_Air"`
		NoRBT       string  `xml:"No_RBT"`
		Viber       string  `xml:"Viber"`
		FDLR        string  `xml:"F-DLR"`
	} `xml:"Call"`
}
