package blmkc

type Summary struct {
	Total          int
	Wanted         int
	Identified     int
	Apprehended    int
	Resolved       int
	PoliceReported int
	PendingReview  int
	LossAmount     float64
}

type StatusCount struct {
	Status string
	Count  int
}

type Case struct {
	ID                 uint64
	CaseNumber         string
	StoreName          string
	Title              string
	OccurredAt         string
	CaseStatus         string
	VerificationStatus string
	IncidentType       string
	PrimarySubject     string
	ItemCount          int
	EvidenceCount      int
}

type DashboardData struct {
	Summary      Summary
	RecentCases  []Case
	StatusCounts []StatusCount
}

type Filter struct {
	Search             string
	StoreID            uint64
	CaseStatus         string
	VerificationStatus string
	StartDate          string
	EndDate            string
	Page               int
	PageSize           int
}

type Option struct {
	ID   uint64
	Code string
	Name string
}

type CaseListData struct {
	Cases      []Case
	Stores     []Option
	Filter     Filter
	Total      int
	TotalPages int
	PrevPage   int
	NextPage   int
}

type WantedData struct {
	Cases           []Case
	Search          string
	Status          string
	WantedCount     int
	IdentifiedCount int
}

type Identity struct {
	ID             uint64
	SubjectNumber  string
	FullName       string
	Aliases        string
	IdentityStatus string
	CaseCount      int
	StoreCount     int
	LastCaseDate   string
	LastCaseStatus string
}

type IdentityData struct {
	Subjects []Identity
	Search   string
	Total    int
}

type VerificationData struct {
	Cases     []Case
	Status    string
	Search    string
	Submitted int
	Revision  int
	Verified  int
	Rejected  int
}
