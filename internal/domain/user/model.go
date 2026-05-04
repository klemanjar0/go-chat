package user

import (
	"time"

	"go-chat/internal/db/sqlcgen"
	"go-chat/pkg/utilid"

	"github.com/jackc/pgx/v5/pgtype"
)

type User struct {
	ID           string
	Username     string
	PasswordHash string
	FirstName    string
	LastName     string
	AvatarURL    *string
	Phone        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CreateInput struct {
	Username     string
	PasswordHash string
	FirstName    string
	LastName     string
	AvatarURL    *string
	Phone        string
}

func fromPg(row *sqlcgen.User) *User {
	return &User{
		ID:           utilid.FromPg(row.ID).AsString(),
		Username:     row.Username,
		PasswordHash: row.PasswordHash,
		FirstName:    row.FirstName,
		LastName:     row.LastName,
		AvatarURL:    row.AvatarUrl,
		Phone:        row.Phone,
		CreatedAt:    row.CreatedDate.Time,
		UpdatedAt:    row.UpdatedDate.Time,
	}
}

func (u *User) ToPg() *sqlcgen.User {
	return &sqlcgen.User{
		ID:           utilid.FromString(u.ID).AsPgUUID(),
		Username:     u.Username,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		PasswordHash: u.PasswordHash,
		Phone:        u.Phone,
		AvatarUrl:    u.AvatarURL,
		CreatedDate:  pgtype.Timestamptz{Valid: true, Time: u.CreatedAt},
		UpdatedDate:  pgtype.Timestamptz{Valid: true, Time: u.UpdatedAt},
	}
}

func (u *User) UpdatePassword(hash string) {
	u.PasswordHash = hash
	u.UpdatedAt = time.Now()
}

func (u *User) UpdateProfile(firstName, lastName string) {
	u.FirstName = firstName
	u.LastName = lastName
	u.UpdatedAt = time.Now()
}

func (u *User) SetAvatar(url *string) {
	u.AvatarURL = url
	u.UpdatedAt = time.Now()
}
