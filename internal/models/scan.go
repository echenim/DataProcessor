package models

type Scan struct {
	Ip          string
	Port        uint32
	Service     string
	Timestamp   int64
	DataVersion int
	Data        map[string]string
}
