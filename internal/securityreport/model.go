package securityreport

import "html/template"

type Option struct {
	ID        uint64
	Name      string
	Code      string
	StartTime string
	EndTime   string
}

type ListItem struct {
	ID, RepresentativeID, MemberCount, LogCount uint64
	ReportDate, StoreName, ShiftName            string
	RepresentativeName, Status, HighestLevel    string
	WorkStartedAt, WorkEndedAt                  string
	CanEdit                                     bool
}

type Member struct {
	ID       uint64
	UserID   uint64
	Name     string
	Role     string
	JoinedAt string
}

type HourlyLog struct {
	ID, ReportID, CreatedByID                         uint64
	ReportedAt, Level, AreaName, LocationNotes        string
	Description, ActionTaken, FollowUpStatus, Creator string
	AreaID                                            uint64
	CanEdit                                           bool
}

type Detail struct {
	ID, ShiftID, RepresentativeID, SupervisorID              uint64
	ReportDate, StoreName, ShiftName, RepresentativeName     string
	RawReportDate                                            string
	SupervisorName, WorkStartedAt, WorkEndedAt, OpeningNotes string
	ClosingSummary, Status, HighestLevel                     string
	NormalCount, WaspadaCount, UrgentCount                   int
	Members                                                  []Member
	Logs                                                     []HourlyLog
	Handover                                                 *Handover
}

type Handover struct {
	ID, ReceiverUserID, SupervisorUserID                     uint64
	GiverName, ReceiverName, SupervisorName                  string
	HandoverNotes, PendingItems, InventoryNotes, Status      string
	HandedOverAt, ReceivedAt, SupervisorSignedAt             string
	HasGiverSignature, HasReceiverSignature, HasSPVSignature bool
	GiverSignature, ReceiverSignature, SPVSignature          template.URL
}

type HistoryFilter struct {
	StartDate, EndDate, Status, Level string
	ShiftID, UserID                   uint64
}
type HistoryData struct {
	Filter        HistoryFilter
	Reports       []ListItem
	Shifts, Users []Option
}
type MonitoringData struct {
	ActiveShifts, MissingReports, LateReports, UrgentOpen, WaspadaOpen int
	WaitingReceiver, WaitingSPV, RevisionReports                       int
	Active, Late, Waiting, Revisions                                   []ListItem
	Missing                                                            []MissingShift
	Urgent, Waspada                                                    []HourlyLog
}

type MissingShift struct {
	StoreName, ShiftName, StartTime string
}

type PageData struct {
	Reports     []ListItem
	Shifts      []Option
	Users       []Option
	Supervisors []Option
	Areas       []Option
}

type CreateInput struct {
	StoreID, ShiftID, RepresentativeID, CreatorID uint64
	SupervisorID                                  *uint64
	ReportDate, OpeningNotes                      string
	MemberIDs                                     []uint64
}

type UpdateInput struct {
	ReportID, ActorID, StoreID, ShiftID, RepresentativeID uint64
	SupervisorID                                          *uint64
	ReportDate, OpeningNotes                              string
	MemberIDs                                             []uint64
}

type HourlyLogInput struct {
	ReportID, LogID, CreatorID uint64
	AreaID                     *uint64
	Level                      string
	LocationNotes              string
	Description                string
	ActionTaken                string
	FollowUpStatus             string
}

type HandoverInput struct {
	ReportID, ActorID, ReceiverID, SupervisorID uint64
	ClosingSummary, HandoverNotes               string
	PendingItems, InventoryNotes                string
	Signature                                   string
}

type ReceiveInput struct {
	ReportID, ActorID uint64
	Notes, Signature  string
	Accept            bool
}

type ReviewInput struct {
	ReportID, ActorID uint64
	Notes, Signature  string
	Approve, AllowAny bool
}
