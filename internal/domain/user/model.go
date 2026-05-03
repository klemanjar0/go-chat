package user

import (
	"go-chat/internal/db/sqlcgen"
	"go-chat/pkg/utilid"
	"time"
)

type User struct {
	ID           string
	Username     string
	PasswordHash string
	FirstName    string
	LastName     string
	AvatarUrl    *string
	Phone        string
	CreatedDate  time.Time
	UpdatedDate  time.Time
}

func FromPg(pgUser *sqlcgen.User) *User {
	return &User{
		ID:           utilid.FromPg(pgUser.ID).AsString(),
		Username:     pgUser.Username,
		PasswordHash: pgUser.PasswordHash,
		FirstName:    pgUser.FirstName,
		LastName:     pgUser.LastName,
		AvatarUrl:    pgUser.AvatarUrl,
		Phone:        pgUser.Phone,
		CreatedDate:  pgUser.CreatedDate.Time,
		UpdatedDate:  pgUser.UpdatedDate.Time,
	}
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

func (user *User) SetAvatar(url *string) {
	user.AvatarUrl = url
	user.UpdatedDate = time.Now()
}
