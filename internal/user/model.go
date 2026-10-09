package user

type User struct {
	ID             uint64
	EmployeeNumber string
	Name           string
	Username       string
	Email          string
	Phone          string
	Status         string
	StoreID        uint64
	StoreName      string
	RoleID         uint64
	RoleNames      string
	StoreCount     int
	LastLogin      string
}

type StoreOption struct {
	ID   uint64
	Name string
}

type RoleOption struct {
	ID   uint64
	Name string
}

type Input struct {
	EmployeeNumber string
	Name           string
	Username       string
	Email          string
	Phone          string
	Password       string
	Status         string
	StoreID        uint64
	RoleID         uint64
}
