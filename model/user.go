package model

import "time"

type User struct {
	UserID    string     `json:"user_id"`              
	Name      string     `json:"name"`                  
	Email     string     `json:"email"`
	GuardianEmail *string    `json:"guardian_name,omitempty"`                  
	GuardianID *string    `json:"guardian_id,omitempty"` 
	CreatedAt time.Time  `json:"created_at"`           
}
