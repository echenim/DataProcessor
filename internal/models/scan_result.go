package models

import "time"

type ScannedResult struct {
	IP        string    `json:"ip"`
	Port      uint32    `json:"port"`
	Service   string    `json:"service"`
	Timestamp time.Time `json:"timestamp"`
	Response  string    `json:"response,omitempty"`
}
