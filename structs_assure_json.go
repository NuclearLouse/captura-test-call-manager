// structs_assure_json.go
//
// The file contains the structures that are needed
// to process JSON responses from the "Assure" system.
//
package main

import (
	"strings"
	"time"
)

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

func (r *TestBatches) ParseTime(pt string) time.Time {
	var t time.Time
	if pt == "" {
		return t
	}
	st := strings.Split(pt, ".")
	t, _ = time.Parse("2006-01-02T15:04:05", st[0])
	return t
}

type TestBatches struct {
	TestBatchID    int    `json:"TestBatchID"`
	StatusID       int    `json:"StatusID"`
	Status         string `json:"Status"`
	IsDone         bool   `json:"IsDone"`
	Created        string `json:"Created"` //time.Time
	TestBatchItems []testBatchItem
}

type testBatchItem struct {
	StatusID int    `json:"StatusID"`
	Status   string `json:"Status"`
	IsDone   bool   `json:"IsDone"`
}

func (r *TestBatchResults) ParseTime(pt string) time.Time {
	var t time.Time
	if pt == "" {
		return t
	}
	st := strings.Split(pt, ".")
	t, _ = time.Parse("2006-01-02T15:04:05", st[0])
	return t
}

type TestBatchResults struct {
	TestBatchResult1 []batchResult
}

type queryResults struct {
	QueryResult1 []batchQuery
}

type batchQuery struct {
	CallResultID int
	APartyAudio  string `json:"A Party Audio"`
	PGAD         float64
	TestType     string `json:"Test Type"`
}

// Потом надо убрать не нужные поля, чтобы не засорять код
type batchResult struct {
	AdapterInstanceIDS        int
	AdapterInstanceNameS      string
	ANumberIsEncrypted        bool   `json:"A_NumberIsEncrypted"`
	AlphaOK                   bool   `json:"Alpha OK"`
	AlphaS                    string `json:"Alpha (S)"`
	AlphaR                    string `json:"Alpha (R)"`
	APartyAudio               []byte `json:"A Party Audio"`
	APIResponse               string `json:"API Response"`
	AdapterInstanceIDR        int
	AdapterInstanceNameR      string
	BNumberIsEncrypted        bool    `json:"B_NumberIsEncrypted"`
	BRegion                   string  `json:"B Region"`
	BDestination              string  `json:"B Destination"`
	BNetwork                  string  `json:"B Network"`
	BTestNode                 string  `json:"B TestNode"`
	BAlertTime                float64 `json:"B Alert Time"`
	BConnectTime              float64 `json:"B Connect Time"`
	CallBatchItemID           int
	CallBatchStatusTypeID     int
	CallBatchExceptionMessage string
	CallDuration              float64 `json:"Call Duration"`
	CallResultID              int
	CLI                       string
	CLIDelivered              string `json:"CLI Delivered"`
	ContentOK                 bool   `json:"Content OK"`
	CustomerID                string
	DisconnectTime            float64 `json:"Disconnect Time"`
	DelTimeOK                 bool    `json:"Del. Time OK"`
	DelRepOK                  bool    `json:"Del. Rep. OK"`
	DelTime                   int     `json:"Del. Time (sec)"`
	DeliveryDetails           string  `json:"Delivery Details"`
	DelTimeLimit              string  `json:"Del. Time Limit"` //time.Time
	DeliveryUpdateDelay       float64
	DCSCharacterSetR          string
	Duplicates                int
	ErrorMsg                  string `json:"Error Msg."`
	ExceptionMsg              string `json:"Exception Msg."`
	UITestStatusID            int
	UITestStatusDisplay       string
	Modified                  string //time.Time
	HasAudio                  int
	HasSecondAudio            int
	TestScenarioRuntimeUID    int
	Result                    string
	Route                     string
	TestStartTime             string  `json:"Test Start Time"` //time.Time
	MOSA                      float64 `json:"MOS A"`
	PGRD                      float64
	PGAD                      float64
	SNR                       float64 `json:"SNR (dB)"`
	SpeechLevel               float64 `json:"Speech Level (dB)"`
	VoiceQualityProblem       string  `json:"Voice Quality Problem"`
	ReleaseCause              string  `json:"Release Cause"`
	ReleaseLocation           string  `json:"Release Location"`
	SpeechFirstDetectedTime   string  `json:"Speech First Detected Time"` //time.Time
	SpeechLastDetectedTime    string  `json:"Speech Last Detected Time"`  //time.Time
	IP                        float64
	SMSResultID               int
	Network                   string
	TestNode                  string `json:"Test Node"`
	BNumber                   string `json:"B Number"`
	Supplier                  string
	SentTime                  string `json:"Sent Time"`
	OAOK                      bool   `json:"OA OK"`
	STonOK                    bool   `json:"S. TON OK"`
	Smsc                      string
	SmscOwner                 string `json:"SMSC Owner"`
	OAS                       string `json:"OA (S)"`
	OAR                       string `json:"OA (R)"`
	STonS                     string `json:"S. TON (S)"`
	STonR                     string `json:"S. TON (R)"`
	SNumPlanS                 string `json:"S. Num. Plan (S)"`
	SNumPlanR                 string `json:"S. Num. Plan (R)"`
	TextS                     string `json:"Text (S)"`
	TextR                     string `json:"Text (R)"`
	NoSegS                    int    `json:"No of Seg (S)"`
	NoSegR                    int    `json:"No of Seg (R)"`
	RecTime                   string `json:"Rec. Time"`      //time.Time
	SMSCRecTime               string `json:"SMSC Rec. Time"` //time.Time
	DelRepTime                string `json:"Del. Rep. Time"` //time.Time
	MessageID                 int    `json:"Message Id"`
	Template                  string
	SMS                       string
	ResultRecTime             string `json:"Result Rec. Time"` //time.Time
	UDHS                      string `json:"UDH (S)"`
	UDHR                      string `json:"UDH (R)"`
	PDUR                      string `json:"PDU (R)"`
	PoPIDR                    int
	ResultTrace               string `json:"Result Trace"`
	SMSID                     string
	SMSIDExpireTime           string //time.Time
	SMSTemplateID             int
	SMSTemplateModified       string
	SubmitSMSResponseDelay    float64
	SMSReceiveDelay           float64
	SMSResultReceiveID        int
	SMSResultIDR              int
	SMR                       string
	SegmentDuplicate          int
	SMSResultSendID           int
	SMSResultIDS              int
	PoPIDS                    int
	SubmitSMSResponseTime     string //time.Time
	StatusUpdateCode          string
	StatusUpdateTime          string //time.Time
	TestType                  string `json:"Test Type"`
}

// type TestSet struct {
// 	"PoPName": "<StringValue>",
// "NoOfExecutions": <IntegerValue>,
// "Priority": <IntegerValue>,
// "TestSetItems": [
// {
// "TestTypeName": "<StringValue>",
// "RouteID": <IntegerValue>,
// "RouteExtID": <IntegerValue>,
// "SMSRouteID": <IntegerValue>,
// "SMSRouteExtID": <IntegerValue>,
// "SMSTemplateName": "<StringValue>",
// "SMSTemplateModifications": <JSON Object>,
// "ATestNodeUID": <StringValue>,
// "APLMN": <StringValue>,
// "BTestNodeUID": <StringValue>,
// "DestinationID": <IntegerValue>,
// "DestinationExtID": <IntegerValue>,
// "BPLMN": <StringValue>,
// "AllTestNodes": <BooleanValue>,
// "PhoneNumber": "<StringValue>"
// }
