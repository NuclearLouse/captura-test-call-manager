// structs_assure_db.go
//
// The file contains the structures that are needed to work with "Assure" tables.
//
package main

import "os"

func (assureAPI) TableName() string {
	return os.Getenv("SCHEMA_PG") + "CallingSys_API_Assure"
}

type assureAPI struct {
	SystemName        string `gorm:"size:50;foreignkey:CallingSys_Settings.SystemName"`
	SystemID          int    `gorm:"type:int"`
	URL               string `gorm:"size:100"`
	User              string `gorm:"size:100"`
	Pass              string `gorm:"size:100"`
	StatusTests       string `gorm:"size:100"`
	TestResults       string `gorm:"size:100"`
	QueryResults      string `gorm:"size:100"`
	Routes            string `gorm:"size:25"`
	Destinations      string `gorm:"size:25"`
	Nodes             string `gorm:"size:25"`
	NewTestGet        string `gorm:"size:25"`
	NodesCapabilities string `gorm:"size:25"`
}

func (assureRoutes) TableName() string {
	return os.Getenv("SCHEMA_PG") + "CallingSys_assure_routes"
}

type assureRoutes struct {
	Active                 bool   `gorm:"type:boolean"`
	CallParameter          string `gorm:"size:100"`
	Carrier                string `gorm:"size:100"`
	CarrierID              string `gorm:"size:100"`
	Channel                int    `gorm:"type:int"`
	ChannelPoolID          int    `gorm:"type:int"`
	ChannelPoolName        string `gorm:"size:100"`
	Created                string `gorm:"size:100"`
	CreatedBy              int    `gorm:"type:int"`
	Description            string `gorm:"size:100"`
	DialerName             string `gorm:"size:100"`
	ExtRouteID             string `gorm:"size:100"`
	IsMainProduct          string `gorm:"size:100"`
	Modified               string `gorm:"size:100"`
	ModifiedBy             int    `gorm:"type:int"`
	Name                   string `gorm:"size:100"`
	NumberManipulationRule string `gorm:"size:100"`
	PoPID                  int    `gorm:"type:int"`
	Prefix                 string `gorm:"size:100"`
	RouteClass             string `gorm:"size:100"`
	RouteID                int    `gorm:"type:int"`
	RouteImportanceID      int    `gorm:"type:int"`
	RouteImportanceName    string `gorm:"size:100"`
	RouteTypeID            string `gorm:"size:100"`
	RouteTypeName          string `gorm:"size:100"`
	ShortName              string `gorm:"size:100"`
	SwitchID               int    `gorm:"type:int"`
	SwitchName             string `gorm:"size:100"`
}

func (assureDestinations) TableName() string {
	return os.Getenv("SCHEMA_PG") + "CallingSys_assure_destinations"
}

type assureDestinations struct {
	CountryID                 int    `gorm:"type:int"`
	CountryName               string `gorm:"size:100"`
	Created                   string `gorm:"size:100"`
	CreatedBy                 int    `gorm:"type:int"`
	DestinationCategoryID     int    `gorm:"type:int"`
	DestinationCategoryName   string `gorm:"size:100"`
	DestinationExtID          string `gorm:"size:100"`
	DestinationID             int    `gorm:"type:int"`
	DestinationImportanceID   int    `gorm:"type:int"`
	DestinationImportanceName string `gorm:"size:100"`
	Modified                  string `gorm:"size:100"`
	ModifiedBy                int    `gorm:"type:int"`
	Name                      string `gorm:"size:100"`
	PoPID                     int    `gorm:"size:100"`
	ShortName                 string `gorm:"size:100"`
}

func (assureNodes) TableName() string {
	return os.Getenv("SCHEMA_PG") + "CallingSys_assure_nodes"
}

type assureNodes struct {
	APartyCallParameter     string `gorm:"size:150"`
	APartyCustomNumberMRule string `gorm:"size:100"`
	APartyNumberMRule       string `gorm:"size:150"`
	CallParameter           string `gorm:"size:100"`
	ChannelProperties       string `gorm:"size:100"`
	Created                 string `gorm:"size:100"`
	CreatedBy               int    `gorm:"type:int"`
	Description             string `gorm:"size:100"`
	DisplayName             string `gorm:"size:100"`
	EquipmentID             int    `gorm:"type:int"`
	EquipmentName           string `gorm:"size:100"`
	HomePLMN                string `gorm:"size:100"`
	IMEI                    string `gorm:"size:100"`
	InternalComment         string `gorm:"size:100"`
	IsSynchronized          bool   `gorm:"type:boolean"`
	ModeID                  int    `gorm:"type:int"`
	Modified                string `gorm:"size:100"`
	ModifiedBy              int    `gorm:"type:int"`
	ModifiedUTC             string `gorm:"size:100"`
	Name                    string `gorm:"size:100"`
	NetworkAndTestNodeName  string `gorm:"size:150"`
	NetworkName             string `gorm:"size:100"`
	NoNumberName            string `gorm:"size:100"`
	NoOfChannels            int    `gorm:"type:int"`
	PLMN                    string `gorm:"size:100"`
	PhoneNumber             string `gorm:"type:text"`
	PhoneNumberFirstPart    string `gorm:"size:100"`
	PhoneNumberIsEncrypted  bool   `gorm:"type:boolean"`
	ProbeContainerUID       string `gorm:"size:150"`
	RTDThreshold            string `gorm:"size:100"`
	TestNodeDirectoryID     int    `gorm:"type:int"`
	TestNodeDirectoryName   string `gorm:"size:100"`
	TestNodeID              int    `gorm:"type:int"`
	TestNodeUID             string `gorm:"size:150"`
}

func (assureNodesCapabilities) TableName() string {
	return os.Getenv("SCHEMA_PG") + "CallingSys_assure_nodes_capabilities"
}

type assureNodesCapabilities struct {
	AAPoPID          int    `gorm:"type:int"`
	TestTypeName     string `gorm:"size:100"`
	DestinationIntID int    `gorm:"type:int"`
	DestinationID    string `gorm:"size:100"`
	DestinationName  string `gorm:"size:100"`
	TestNodeIntID    int    `gorm:"type:int"`
	TestNodeIntUID   string `gorm:"size:100"`
	TestNodeName     string `gorm:"size:100"`
	NetworkName      string `gorm:"size:100"`
	IsAParty         int    `gorm:"type:int"`
}
