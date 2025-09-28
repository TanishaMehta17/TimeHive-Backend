package model

import "time"

type User struct {
	UserID    string     `json:"user_id"`              
	Name      string     `json:"name"`                  
	Email     string     `json:"email"`                 
	GuardianID *string    `json:"guardian_id,omitempty"` 
	CreatedAt time.Time  `json:"created_at"`           
}
