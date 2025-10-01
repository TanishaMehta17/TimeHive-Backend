package model

import "time"

type AppLimit struct {
	LimitID           int       `json:"limit_id"` 
	UserID            string    `json:"user_id"`
	AppID             string    `json:"app_id"`
	DailyLimitMinutes int       `json:"daily_limit_minutes"`
	CreatedAt         time.Time `json:"created_at"`
}