package models

import "time"

type ScannedResult struct {
	IP        string    `json:"ip"`
	Port      int       `json:"port"`
	Service   string    `json:"service"`
	Timestamp time.Time `json:"timestamp"`
	Response  string    `json:"response,omitempty"`
}
