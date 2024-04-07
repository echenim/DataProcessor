package models

type MessageData struct {
	IP          string `json:"ip"`
	Port        int    `json:"port"`
	Service     string `json:"service"`
	DataVersion int    `json:"data_version"`
	Data        Data   `json:"data"`
}

type Data struct {
	ResponseBytesUTF8 string `json:"response_bytes_utf8,omitempty"`
	ResponseStr       string `json:"response_str,omitempty"`
}
