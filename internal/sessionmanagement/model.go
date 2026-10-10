package sessionmanagement

import "time"

type Session struct {
	ID         string
	IPAddress  string
	UserAgent  string
	Browser    string
	Device     string
	Icon       string
	CreatedAt  time.Time
	LastSeenAt time.Time
	ExpiresAt  time.Time
	Created    string
	LastSeen   string
	Expires    string
	Remaining  string
	Current    bool
}
