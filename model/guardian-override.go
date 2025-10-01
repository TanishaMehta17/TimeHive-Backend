package model

import "time"

type GuardianOverride struct {
	Token        string    `json:"token"`
	UserID       string    `json:"user_id"`
	AppID        string    `json:"app_id"`
	ExtraMinutes int       `json:"extra_minutes"`
	ValidFor     time.Time `json:"valid_for"` 
}