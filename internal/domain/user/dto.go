package user

import "time"

type RegisterRequest struct {
	Username  string  `json:"username"`
	Password  string  `json:"password"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Phone     string  `json:"phone"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
}

type TokenResponse struct {
	AccessToken      string    `json:"access_token"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshToken     string    `json:"refresh_token"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
	TokenType        string    `json:"token_type"`
}

type AuthResponse struct {
	User   UserResponse  `json:"user"`
	Tokens TokenResponse `json:"tokens"`
}

func ToUserResponse(u *User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		AvatarURL: u.AvatarUrl,
		Phone:     u.Phone,
		CreatedAt: u.CreatedDate,
	}
}

func ToTokenResponse(p TokenPair) TokenResponse {
	return TokenResponse{
		AccessToken:      p.AccessToken,
		AccessExpiresAt:  p.AccessExpiresAt,
		RefreshToken:     p.RefreshToken,
		RefreshExpiresAt: p.RefreshExpiresAt,
		TokenType:        "Bearer",
	}
}

func ToAuthResponse(r *AuthResult) AuthResponse {
	return AuthResponse{
		User:   ToUserResponse(r.User),
		Tokens: ToTokenResponse(r.Tokens),
	}
}
