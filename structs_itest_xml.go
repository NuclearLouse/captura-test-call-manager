// structs_itest_xml.go
//
// The file contains the structures that are needed
// to process XML responses from the "iTest" system.
//
package main

import (
	"encoding/xml"
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
