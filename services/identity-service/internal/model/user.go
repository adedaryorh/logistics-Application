package model

import "time"

type User struct {
	ID            string     `json:"id"`
	Email         string     `json:"email"`
	Phone         *string    `json:"phone,omitempty"`
	PasswordHash  string     `json:"-"`
	EmailVerified bool       `json:"email_verified"`
	PhoneVerified bool       `json:"phone_verified"`
	Role          string     `json:"role"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
}

type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string
	FamilyID  string
	DeviceID  *string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

type MagicLink struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

type PasswordReset struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

type OAuthSession struct {
	ID                  string
	Provider            string
	State               string
	Nonce               string
	RedirectURI         string
	CodeChallenge       *string
	CodeChallengeMethod *string
	ExpiresAt           time.Time
	UsedAt              *time.Time
	CreatedAt           time.Time
}
