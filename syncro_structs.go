package main

import "time"

func (syncAutomation) TableName() string {
	return schemaPG + "CallingSys_sync_automation"
}

type syncAutomation struct {
	Systemid            int
	Synctype            string
	DoSynch             bool
	UpdateCaptRelations bool
	Syncstate           int
	SyncStart           time.Time
	SyncEnd             time.Time
	Comment             string
}

//-----------------------------------------------------------------------------
//*******************Block of Assure syncro structures*******************
//-----------------------------------------------------------------------------
type destinations struct {
	QueryResult1 []assureDestination
}

func (assureDestination) TableName() string {
	return schemaPG + "CallingSys_assure_destinations"
}

type assureDestination struct {
	CountryID                 int    `json:"CountryID" gorm:"type:int"`
	CountryName               string `json:"CountryName" gorm:"size:100"`
	Created                   string `json:"Created" gorm:"size:100"`
	CreatedBy                 int    `json:"CreatedBy" gorm:"type:int"`
	DestinationCategoryID     int    `json:"DestinationCategoryID" gorm:"type:int"`
	DestinationCategoryName   string `json:"DestinationCategoryName" gorm:"size:100"`
	DestinationExtID          string `json:"DestinationExtID" gorm:"size:100"`
	DestinationID             int    `json:"DestinationID" gorm:"type:int"`
	DestinationImportanceID   int    `json:"DestinationImportanceID"`
	DestinationImportanceName string `json:"DestinationImportanceName" gorm:"size:100"`
	Modified                  string `json:"Modified" gorm:"size:100"`
	ModifiedBy                int    `json:"ModifiedBy" gorm:"type:int"`
	Name                      string `json:"Name" gorm:"size:100"`
	PoPID                     int    `json:"PoPID" gorm:"type:int"`
	ShortName                 string `json:"ShortName" gorm:"size:100"`
}

type routes struct {
	QueryResult1 []assureRoute
}

func (assureRoute) TableName() string {
	return schemaPG + "CallingSys_assure_routes"
}

type assureRoute struct {
	Active                 bool        `json:"Active" gorm:"type:boolean"`
	CallParameter          interface{} `json:"CallParameter" gorm:"size:100"`
	Carrier                string      `json:"Carrier" gorm:"size:100"`
	CarrierID              interface{} `json:"CarrierID" gorm:"size:100"`
	Channel                int         `json:"Channel" gorm:"type:int"`
	ChannelPoolID          int         `json:"ChannelPoolID" gorm:"type:int"`
	ChannelPoolName        string      `json:"ChannelPoolName" gorm:"size:100"`
	Created                string      `json:"Created" gorm:"size:100"`
	CreatedBy              int         `json:"CreatedBy" gorm:"type:int"`
	Description            interface{} `json:"Description" gorm:"size:100"`
	DialerName             string      `json:"DialerName" gorm:"size:100"`
	ExtRouteID             interface{} `json:"ExtRouteID" gorm:"size:100"`
	IsMainProduct          interface{} `json:"IsMainProduct" gorm:"size:100"`
	Modified               string      `json:"Modified" gorm:"size:100"`
	ModifiedBy             int         `json:"ModifiedBy" gorm:"type:int"`
	Name                   string      `json:"Name" gorm:"size:100"`
	NumberManipulationRule interface{} `json:"NumberManipulationRule" gorm:"size:100"`
	PoPID                  int         `json:"PoPID" gorm:"type:int"`
	Prefix                 string      `json:"Prefix" gorm:"size:100"`
	RouteClass             string      `json:"RouteClass" gorm:"size:100"`
	RouteID                int         `json:"RouteID" gorm:"type:int"`
	RouteImportanceID      int         `json:"RouteImportanceID" gorm:"type:int"`
	RouteImportanceName    string      `json:"RouteImportanceName" gorm:"size:100"`
	RouteTypeID            interface{} `json:"RouteTypeID" gorm:"size:100"`
	RouteTypeName          interface{} `json:"RouteTypeName" gorm:"size:100"`
	ShortName              string      `json:"ShortName" gorm:"size:100"`
	SwitchID               int         `json:"SwitchID" gorm:"type:int"`
	SwitchName             string      `json:"SwitchName" gorm:"size:100"`
}
