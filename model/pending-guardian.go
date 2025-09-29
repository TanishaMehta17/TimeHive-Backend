package model

import "time"

type PendingGuardian struct {
    ID            string    `json:"id" db:"id"`
    UserID        string    `json:"user_id" db:"user_id"`
    GuardianEmail string    `json:"guardian_email" db:"guardian_email"`
    Token         string    `json:"token" db:"token"`
    CreatedAt     time.Time `json:"created_at" db:"created_at"`
    ExpiresAt     time.Time `json:"expires_at" db:"expires_at"`
}
