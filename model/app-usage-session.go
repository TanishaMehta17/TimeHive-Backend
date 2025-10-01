package model

import "time"

type AppUsageSession struct {
	SessionID       int        `json:"session_id"` 
	UserID          string     `json:"user_id"`
	AppID           string     `json:"app_id"`
	DeviceID        string     `json:"device_id"`
	SessionStart    time.Time  `json:"session_start"`
	SessionEnd      *time.Time `json:"session_end,omitempty"`      
	DurationMinutes *int       `json:"duration_minutes,omitempty"` 
}