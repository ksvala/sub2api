package service

import "time"

type AdminActionLog struct {
	ID           int64
	AdminID      *int64
	Action       string
	ResourceType string
	ResourceID   *int64
	Payload      string
	IPAddress    string
	UserAgent    string
	CreatedAt    time.Time
}

type AdminActionLogInput struct {
	AdminID      *int64
	Action       string
	ResourceType string
	ResourceID   *int64
	Payload      string
	IPAddress    string
	UserAgent    string
}
