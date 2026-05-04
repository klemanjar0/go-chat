package user

import (
	"time"

	"go-chat/internal/domain/user"
)

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type UpdateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type SetAvatarRequest struct {
	AvatarURL *string `json:"avatar_url"`
}

type Response struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
}

func toResponse(u *user.User) Response {
	return Response{
		ID:        u.ID,
		Username:  u.Username,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		AvatarURL: u.AvatarURL,
		Phone:     u.Phone,
		CreatedAt: u.CreatedAt,
	}
}
