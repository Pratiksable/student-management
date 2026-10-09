package models

import (
	"database/sql"
	"encoding/json"
)

type Exec struct {
	ID                    int            `json:"id,omitempty" db:"id,omitempty"`
	FirstName             string         `json:"first_name,omitempty" db:"first_name,omitempty"`
	LastName              string         `json:"last_name,omitempty" db:"last_name,omitempty"`
	Email                 string         `json:"email,omitempty" db:"email,omitempty"`
	Username              string         `json:"username,omitempty" db:"username,omitempty"`
	Password              string         `json:"password,omitempty" db:"password,omitempty"`
	PasswordChangedAt     sql.NullString `json:"password_changed_at,omitempty" db:"password_changed_at,omitempty"`
	UserCreatedAt         sql.NullString `json:"user_created_at,omitempty" db:"user_created_at,omitempty"`
	PasswordResetCode     sql.NullString `json:"password_reset_code,omitempty" db:"password_reset_code,omitempty"`
	PasswordCodeExpiresAt sql.NullString `json:"password_code_expires_at,omitempty" db:"password_code_expires_at,omitempty"`
	Inactive              bool           `json:"inactive,omitempty" db:"inactive,omitempty"`
	Role                  string         `json:"role,omitempty" db:"role,omitempty"`
}

// MarshalJSON defines the public API representation of an exec. Authentication
// secrets must never be sent back to clients, and SQL nullable values should be
// represented as normal JSON strings or null rather than database internals.
func (e Exec) MarshalJSON() ([]byte, error) {
	type execResponse struct {
		ID                    int     `json:"id"`
		FirstName             string  `json:"first_name"`
		LastName              string  `json:"last_name"`
		Email                 string  `json:"email"`
		Username              string  `json:"username"`
		PasswordChangedAt     *string `json:"password_changed_at"`
		UserCreatedAt         *string `json:"user_created_at"`
		PasswordCodeExpiresAt *string `json:"password_code_expires_at"`
		Inactive              bool    `json:"inactive"`
		Role                  string  `json:"role"`
	}

	return json.Marshal(execResponse{
		ID:                    e.ID,
		FirstName:             e.FirstName,
		LastName:              e.LastName,
		Email:                 e.Email,
		Username:              e.Username,
		PasswordChangedAt:     nullableString(e.PasswordChangedAt),
		UserCreatedAt:         nullableString(e.UserCreatedAt),
		PasswordCodeExpiresAt: nullableString(e.PasswordCodeExpiresAt),
		Inactive:              e.Inactive,
		Role:                  e.Role,
	})
}

func nullableString(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}
