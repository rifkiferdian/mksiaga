package dailyreport

import "mksiaga/internal/storescope"

type Option struct {
	ID   uint64
	Name string
	Code string
}

type ChecklistOption struct {
	ID                  uint64
	Name                string
	Phase               string
	RequiresCameraCount bool
}

type Report struct {
	ID               uint64
	StoreName        string
	ReportDate       string
	ShiftName        string
	OfficerName      string
	Status           string
	CreatedAt        string
	HandoverToUserID *uint64
	CanReceive       bool
	CanReview        bool
}

type WorkflowReport struct {
	ID                  uint64
	ReportNumber        string
	StoreName           string
	OfficerID           uint64
	ReportDate          string
	ShiftName           string
	OfficerName         string
	SupervisorName      string
	WorkStart           string
	WorkEnd             string
	Location            string
	StartNotes          string
	EndNotes            string
	HandoverToName      string
	HandoverToUserID    uint64
	HandoverAt          string
	HandoverNotes       string
	Status              string
	ReviewNotes         string
	Checklists          []WorkflowChecklist
	Monitoring          []WorkflowMonitoring
	Incidents           []WorkflowIncident
	Events              []WorkflowEvent
	ChecklistIssueCount int
	FindingCount        int
	IncidentCount       int
}

type WorkflowChecklist struct{ Phase, Name, Answer, CameraCount, Notes string }
type WorkflowMonitoring struct{ Time, AreaName, Condition, Finding, Action string }
type WorkflowIncident struct{ Time, AreaName, CameraName, TypeName, Description, PersonDetails, Response, Status, Actions string }
type WorkflowEvent struct{ ActorName, FromStatus, ToStatus, Notes, CreatedAt string }

type ChecklistAnswer struct {
	ChecklistID uint64
	Answer      string
	Active      *uint16
	Total       *uint16
	Notes       string
}

type MonitoringLog struct {
	Time      string
	AreaID    *uint64
	Condition string
	Finding   string
	Action    string
}

type Incident struct {
	Time             string
	AreaID           *uint64
	CameraID         *uint64
	IncidentTypeID   *uint64
	OtherType        string
	Description      string
	Gender           string
	Age              string
	Clothing         string
	Features         string
	ResponseMinutes  *uint16
	Status           string
	ActionIDs        []uint64
	OtherActionNotes string
}

type Input struct {
	StoreID          uint64
	OfficerID        uint64
	SupervisorID     *uint64
	ShiftID          uint64
	ReportDate       string
	WorkStart        string
	WorkEnd          string
	Location         string
	StartNotes       string
	EndNotes         string
	HandoverToUserID *uint64
	HandoverAt       string
	HandoverNotes    string
	HandoverFromSign string
	HandoverToSign   string
	Status           string
	Checklists       []ChecklistAnswer
	Monitoring       []MonitoringLog
	Incidents        []Incident
}

type PageData struct {
	Reports        []Report
	Shifts         []Option
	Areas          []Option
	PriorityAreas  []Option
	Cameras        []Option
	Users          []Option
	Checklists     []ChecklistOption
	IncidentTypes  []Option
	Actions        []Option
	DefaultShiftID uint64
}

type RecapFilter struct {
	StartDate string
	EndDate   string
	StoreID   uint64
	Status    string
}

type RecapSummary struct {
	Total, Approved, WaitingHandover, WaitingReview, Revision, Draft int
	Findings, Incidents                                              int
}

type RecapStore struct {
	StoreName                          string
	Total, Approved, Pending, Revision int
	Findings, Incidents                int
}

type RecapRow struct {
	ID                                  uint64
	ReportNumber, ReportDate, StoreName string
	ShiftName, OfficerName, Status      string
	Findings, Incidents                 int
}

type RecapData struct {
	Filter  RecapFilter
	Summary RecapSummary
	Stores  []storescope.Store
	ByStore []RecapStore
	Rows    []RecapRow
}
