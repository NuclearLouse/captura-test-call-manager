// structs_assure.go
//
// The file contains the structures that are needed
// to process JSON responses from the "Assure" system
// and contains the structures that are needed to work with "Assure" tables.
//
package main

//----------------------------------------------------------------------------------
// Structs for prepare tests
//----------------------------------------------------------------------------------
type Destinations struct {
	QueryResult1 []destination
}

type destination struct {
	CountryID                 int    `json:"CountryID"`
	CountryName               string `json:"CountryName"`
	Created                   string `json:"Created"`
	CreatedBy                 int    `json:"CreatedBy"`
	DestinationCategoryID     int    `json:"DestinationCategoryID"`
	DestinationCategoryName   string `json:"DestinationCategoryName"`
	DestinationExtID          string `json:"DestinationExtID"`
	DestinationID             int    `json:"DestinationID"`
	DestinationImportanceID   int    `json:"DestinationImportanceID"`
	DestinationImportanceName string `json:"DestinationImportanceName"`
	Modified                  string `json:"Modified"`
	ModifiedBy                int    `json:"ModifiedBy"`
	Name                      string `json:"Name"`
	PoPID                     int    `json:"PoPID"`
	ShortName                 string `json:"ShortName"`
}

type Routes struct {
	QueryResult1 []route
}

type route struct {
	Active                 bool        `json:"Active"`
	CallParameter          interface{} `json:"CallParameter"`
	Carrier                string      `json:"Carrier"`
	CarrierID              interface{} `json:"CarrierID"`
	Channel                int         `json:"Channel"`
	ChannelPoolID          int         `json:"ChannelPoolID"`
	ChannelPoolName        string      `json:"ChannelPoolName"`
	Created                string      `json:"Created"`
	CreatedBy              int         `json:"CreatedBy"`
	Description            interface{} `json:"Description"`
	DialerName             string      `json:"DialerName"`
	ExtRouteID             interface{} `json:"ExtRouteID"`
	IsMainProduct          interface{} `json:"IsMainProduct"`
	Modified               string      `json:"Modified"`
	ModifiedBy             int         `json:"ModifiedBy"`
	Name                   string      `json:"Name"`
	NumberManipulationRule interface{} `json:"NumberManipulationRule"`
	PoPID                  int         `json:"PoPID"`
	Prefix                 string      `json:"Prefix"`
	RouteClass             string      `json:"RouteClass"`
	RouteID                int         `json:"RouteID"`
	RouteImportanceID      int         `json:"RouteImportanceID"`
	RouteImportanceName    string      `json:"RouteImportanceName"`
	RouteTypeID            interface{} `json:"RouteTypeID"`
	RouteTypeName          interface{} `json:"RouteTypeName"`
	ShortName              string      `json:"ShortName"`
	SwitchID               int         `json:"SwitchID"`
	SwitchName             string      `json:"SwitchName"`
}

type Nodes struct {
	QueryResult1 []node
}

type node struct {
	APartyCallParameter     interface{} `json:"APartyCallParameter"`
	APartyCustomNumberMRule interface{} `json:"APartyCustomNumberMRule"`
	APartyNumberMRule       string      `json:"APartyNumberMRule"`
	CallParameter           interface{} `json:"CallParameter"`
	ChannelProperties       string      `json:"ChannelProperties"`
	Created                 string      `json:"Created"`
	CreatedBy               int         `json:"CreatedBy"`
	Description             string      `json:"Description"`
	DisplayName             string      `json:"DisplayName"`
	EquipmentID             int         `json:"EquipmentID"`
	EquipmentName           string      `json:"EquipmentName"`
	HomePLMN                string      `json:"HomePLMN"`
	IMEI                    interface{} `json:"IMEI"`
	InternalComment         interface{} `json:"InternalComment"`
	IsSynchronized          bool        `json:"IsSynchronized"`
	ModeID                  int         `json:"ModeID"`
	Modified                string      `json:"Modified"`
	ModifiedBy              int         `json:"ModifiedBy"`
	ModifiedUTC             string      `json:"ModifiedUTC"`
	Name                    string      `json:"Name"`
	NetworkAndTestNodeName  string      `json:"NetworkAndTestNodeName"`
	NetworkName             string      `json:"NetworkName"`
	NoNumberName            string      `json:"NoNumberName"`
	NoOfChannels            int         `json:"NoOfChannels"`
	PLMN                    string      `json:"PLMN"`
	PhoneNumber             string      `json:"PhoneNumber"`
	PhoneNumberFirstPart    string      `json:"PhoneNumberFirstPart"`
	PhoneNumberIsEncrypted  bool        `json:"PhoneNumberIsEncrypted"`
	ProbeContainerUID       string      `json:"ProbeContainerUID"`
	RTDThreshold            interface{} `json:"RTDThreshold"`
	TestNodeDirectoryID     int         `json:"TestNodeDirectoryID"`
	TestNodeDirectoryName   string      `json:"TestNodeDirectoryName"`
	TestNodeID              int         `json:"TestNodeID"`
	TestNodeUID             string      `json:"TestNodeUID"`
}

type NodeCapabilities struct {
	QueryResult1 []nodeCap
}

type nodeCap struct {
	AAPoPID          int         `json:"AAPoPID"`
	DestinationID    interface{} `json:"DestinationID"`
	DestinationIntID int         `json:"DestinationIntID"`
	DestinationName  string      `json:"DestinationName"`
	IsAParty         int         `json:"IsAParty"`
	NetworkName      string      `json:"NetworkName"`
	TestNodeIntID    int         `json:"TestNodeIntID"`
	TestNodeIntUID   string      `json:"TestNodeIntUID"`
	TestNodeName     string      `json:"TestNodeName"`
	TestTypeName     string      `json:"TestTypeName"`
}

type NodeStatuses struct {
	QueryResult1 []nodeStatus
}

type nodeStatus struct {
	CountryName           string `json:"CountryName"`
	Description           string `json:"Description"`
	DestinationName       string `json:"DestinationName"`
	EquipmentID           int    `json:"EquipmentID"`
	EquipmentName         string `json:"EquipmentName"`
	GSMStatusChanged      string `json:"GSMStatusChanged"`
	GSMStatusID           int    `json:"GSMStatusID"`
	GSMStatusText         string `json:"GSMStatusText"`
	IPStatusChanged       string `json:"IPStatusChanged"`
	IPStatusID            int    `json:"IPStatusID"`
	IPStatusText          string `json:"IPStatusText"`
	Name                  string `json:"Name"`
	NetworkName           string `json:"NetworkName"`
	PoPID                 int    `json:"PoPID"`
	PoPName               string `json:"PoPName"`
	RegionName            string `json:"RegionName"`
	TestNodeDirectoryID   int    `json:"TestNodeDirectoryID"`
	TestNodeDirectoryName string `json:"TestNodeDirectoryName"`
	TestNodeID            int    `json:"TestNodeID"`
	TestNodeStatusID      int    `json:"TestNodeStatusID"`
	TestNodeUID           string `json:"TestNodeUID"`
}

type SMSRoutes struct {
	QueryResult1 []smsRoute
}

type smsRoute struct {
	Active                 bool        `json:"Active"`
	Carrier                string      `json:"Carrier"`
	Created                string      `json:"Created"`
	CreatedBy              int         `json:"CreatedBy"`
	Description            interface{} `json:"Description"`
	IsMainProduct          interface{} `json:"IsMainProduct"`
	Modified               string      `json:"Modified"`
	ModifiedBy             int         `json:"ModifiedBy"`
	Name                   string      `json:"Name"`
	NumberManipulationRule interface{} `json:"NumberManipulationRule"`
	PoPID                  int         `json:"PoPID"`
	RouteClass             string      `json:"RouteClass"`
	RouteTypeID            interface{} `json:"RouteTypeID"`
	RouteTypeName          interface{} `json:"RouteTypeName"`
	SMSAdapterInstanceID   int         `json:"SMSAdapterInstanceID"`
	SMSAdapterInstanceName string      `json:"SMSAdapterInstanceName"`
	SMSRouteExtID          interface{} `json:"SMSRouteExtID"`
	SMSRouteID             int         `json:"SMSRouteID"`
	SMSRouteImportanceID   int         `json:"SMSRouteImportanceID"`
	SMSRouteImportanceName string      `json:"SMSRouteImportanceName"`
	SMSRouteProperties     string      `json:"SMSRouteProperties"`
	ShortName              string      `json:"ShortName"`
}

type SMSTemplates struct {
	QueryResult1 []smsTemplate
}

type smsTemplate struct {
	Created                 string      `json:"Created"`
	CreatedBy               int         `json:"CreatedBy"`
	Description             interface{} `json:"Description"`
	Modified                string      `json:"Modified"`
	ModifiedBy              int         `json:"ModifiedBy"`
	Name                    string      `json:"Name"`
	PoPID                   int         `json:"PoPID"`
	SMSAdapaterInstanceName string      `json:"SMSAdapaterInstanceName"`
	SMSAdapterInstanceID    int         `json:"SMSAdapterInstanceID"`
	SMSTemplateID           int         `json:"SMSTemplateID"`
	TestTypeID              int         `json:"TestTypeID"`
	TestTypeName            string      `json:"TestTypeName"`
}

//----------------------------------------------------------------------------------
// Structs for initiated new tests
//----------------------------------------------------------------------------------
type testSetDestination struct {
	NoOfExecutions int
	TestSetItems   []batchDestination
}

type batchDestination struct {
	TestTypeName  string
	RouteID       int
	DestinationID int
}

type testSetBnumbers struct {
	TestSetItems []batchBnumbers
}

type batchBnumbers struct {
	TestTypeName string
	RouteID      int
	PhoneNumber  int64
}

//----------------------------------------------------------------------------------
// Structs for obtain tests results
//----------------------------------------------------------------------------------
type TestBatches struct {
	TestBatchID    int    `json:"TestBatchID"`
	StatusID       int    `json:"StatusID"`
	Status         string `json:"Status"`
	IsDone         bool   `json:"IsDone"`
	Created        string `json:"Created"` //time.Time
	TestBatchItems []struct {
		IsDone   bool   `json:"IsDone"`
		Status   string `json:"Status"`
		StatusID int    `json:"StatusID"`
	} `json:"TestBatchItems"`
}

type TestBatchResults struct {
	QueryResult1 []batchResult
}

type batchResult struct {
	CallResultID            int
	ANumber                 string `json:"A number"`
	ANumberIsEncrypted      bool   `json:"A_NumberIsEncrypted"`
	BNumberIsEncrypted      bool   `json:"B_NumberIsEncrypted"`
	Result                  string
	TestType                string `json:"Test Type"`
	BNetwork                string `json:"B Network"`
	BTestNode               string `json:"B Test Node"`
	Bnumber                 string `json:"B number (first 6 digits)"`
	Route                   string
	TestStartTime           string `json:"Test Start Time"` //time.Time
	Alert                   interface{}
	Connect                 interface{}
	CLI                     interface{} //int
	VQA                     interface{} `json:"VQ A"`
	VQB                     interface{} `json:"VQ B"`
	FAS                     interface{} //int
	CLIDelivered            string      `json:"CLI Delivered"`
	MOSA                    float64     `json:"MOS A"`
	MOSB                    float64     `json:"MOS B"`
	PGRD                    float64
	PGAD                    float64
	CallDuration            float64     `json:"Call Duration"`
	DisconnectTime          float64     `json:"Disconnect Time"`
	BAlertTime              float64     `json:"B Alert Time"`
	BCallDuration           float64     `json:"B Call Duration"`
	BConnectTime            float64     `json:"B Connect Time"`
	BDisconnectTime         float64     `json:"B Disconnect Time"`
	ReleaseCause            string      `json:"Release Cause"`
	ReleaseLocation         string      `json:"Release Location"`
	SpeechFirstDetectedTime interface{} `json:"Speech First Detected Time"` //time.Time
	SpeechLastDetectedTime  interface{} `json:"Speech Last Detected Time"`  //time.Time
	APartyAudio             string      `json:"A Party Audio"`
}

func (assureAPI) TableName() string {
	return schemaPG + "CallingSys_API_Assure"
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
	return schemaPG + "CallingSys_assure_routes"
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
	return schemaPG + "CallingSys_assure_destinations"
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
	return schemaPG + "CallingSys_assure_nodes"
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
	return schemaPG + "CallingSys_assure_nodes_capabilities"
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

// possible fields in the response - batchResult
// AdapterInstanceIDS        int
// AdapterInstanceNameS      string
// AlphaOK                   bool   `json:"Alpha OK"`
// AlphaS                    string `json:"Alpha (S)"`
// AlphaR                    string `json:"Alpha (R)"`
// APIResponse               string `json:"API Response"`
// AdapterInstanceIDR        int
// AdapterInstanceNameR      string
// BRegion                   string  `json:"B Region"`
// BDestination              string  `json:"B Destination"`
// CallBatchItemID           int
// CallBatchStatusTypeID     int
// CallBatchExceptionMessage string
// ContentOK                 bool `json:"Content OK"`
// CustomerID                string
// DelTimeOK                 bool    `json:"Del. Time OK"`
// DelRepOK                  bool    `json:"Del. Rep. OK"`
// DelTime                   int     `json:"Del. Time (sec)"`
// DeliveryDetails           string  `json:"Delivery Details"`
// DelTimeLimit              string  `json:"Del. Time Limit"` //time.Time
// DeliveryUpdateDelay       float64
// DelRepTime                string `json:"Del. Rep. Time"` //time.Time
// DCSCharacterSetR          string
// Duplicates                int
// ErrorMsg                  string `json:"Error Msg."`
// ExceptionMsg              string `json:"Exception Msg."`
// HasAudio                  int
// HasSecondAudio            int
// IP                        float64
// Modified                  string //time.Time
// MessageID                 int    `json:"Message Id"`
// NoSegS                    int    `json:"No of Seg (S)"`
// NoSegR                    int    `json:"No of Seg (R)"`
// Network                   string
// OAOK                      bool    `json:"OA OK"`
// OAS                       string  `json:"OA (S)"`
// OAR                       string  `json:"OA (R)"`
// PDUR                      string  `json:"PDU (R)"`
// PoPIDR                    int
// PoPIDS                    int
// ResultTrace               string      `json:"Result Trace"`
// ResultRecTime             string      `json:"Result Rec. Time"` //time.Time
// RecTime                   string      `json:"Rec. Time"`        //time.Time
// SNR                       float64     `json:"SNR (dB)"`
// SpeechLevel               float64     `json:"Speech Level (dB)"`
// SMSResultID               int
// Supplier                  string
// SentTime                  string `json:"Sent Time"`
// STonOK                    bool   `json:"S. TON OK"`
// Smsc                      string
// SmscOwner                 string `json:"SMSC Owner"`
// STonS                     string `json:"S. TON (S)"`
// STonR                     string `json:"S. TON (R)"`
// SNumPlanS                 string `json:"S. Num. Plan (S)"`
// SNumPlanR                 string `json:"S. Num. Plan (R)"`
// SMSID                     string
// SMSIDExpireTime           string //time.Time
// SMSTemplateID             int
// SMSTemplateModified       string
// SMSCRecTime               string `json:"SMSC Rec. Time"` //time.Time
// SubmitSMSResponseDelay    float64
// SMS                       string
// SMSReceiveDelay           float64
// SMSResultReceiveID        int
// SMSResultIDR              int
// SMR                       string
// SegmentDuplicate          int
// SMSResultSendID           int
// SMSResultIDS              int
// SubmitSMSResponseTime     string //time.Time
// StatusUpdateCode          string
// StatusUpdateTime          string //time.Time
// TestScenarioRuntimeUID    int
// TestNode                  string `json:"Test Node"`
// TextS                     string `json:"Text (S)"`
// TextR                     string `json:"Text (R)"`
// Template                  string
// UITestStatusID            int
// UITestStatusDisplay       string
// UDHS                      string      `json:"UDH (S)"`
// UDHR                      string      `json:"UDH (R)"`
// VoiceQualityProblem       string      `json:"Voice Quality Problem"`
