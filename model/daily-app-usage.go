package model

import "time"

type DailyAppUsage struct {
	UsageID      int       `json:"usage_id"`  
	UserID       string    `json:"user_id"`
	AppID        string    `json:"app_id"`
	UsageDate    time.Time `json:"usage_date"`
	TotalMinutes int       `json:"total_minutes"`
	UpdatedAt    time.Time `json:"updated_at"`
}
