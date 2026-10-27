package dashboard

import "html/template"

type Summary struct {
	Total, Wanted, Apprehended, Pending, Other int
}

type Composition struct {
	Draft, Wanted, Identified, Apprehended, Resolved, Police int
}

type Record struct {
	ID                                         uint64
	Name, Alias, Status, StatusClass, Initials string
	Verification, VerificationClass            string
	Location, Date, Note, Search               string
}

type Activity struct {
	Label, Location, Time, Color string
}

type Data struct {
	Summary                              Summary
	Composition                          Composition
	Records                              []Record
	Activities                           []Activity
	MonthLabels                          []string
	TotalPoints, HandledPoints, AreaPath string
	Gradient                             template.CSS
}
