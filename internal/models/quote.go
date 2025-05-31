package models

import "time"

type Quote struct {
	ID        int       `json:"id"`
	Author    string    `json:"author"`
	Quote     string    `json:"quote"`
	CreatedAt time.Time `json:"created_at"`
}
