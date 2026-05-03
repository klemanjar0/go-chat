package user

import "time"

type User struct {
	ID           string
	Username     string
	PasswordHash string
	FirstName    string
	LastName     string
	AvatarUrl    string
	Phone        string
	CreatedDate  time.Time
	UpdatedDate  time.Time
}

func FromPg() *User {
	return &User{}
}

func (user *User) UpdatePassword(newPasswordHash string) {
	user.PasswordHash = newPasswordHash
	user.UpdatedDate = time.Now()
}

func (user *User) UpdateProfileData(firstName, lastName string) {
	user.FirstName = firstName
	user.LastName = lastName
	user.UpdatedDate = time.Now()
}

func (user *User) SetAvatar(url string) {
	user.AvatarUrl = url
	user.UpdatedDate = time.Now()
}
