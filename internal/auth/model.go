package auth

type User struct {
	ID           uint64
	Name         string
	Username     string
	PasswordHash string
	UserStoreID  uint64
	StoreID      uint64
	StoreName    string
	RoleName     string
}
