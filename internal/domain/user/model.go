package user

import "time"

type User struct {
	ID           string
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
	CreatedDate  time.Time
	UpdatedDate  time.Time
}

func FromPg() *User {
	return &User{}
}
