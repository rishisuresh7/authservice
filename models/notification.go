package models

import "encoding/json"

type ChannelMessage struct {
	Medium 		 string `json:"medium"`
	Type		 string  `json:"type"`
	Notification []byte `json:"notification"`
}

type SMS struct {
	To 		[]string `json:"to"`
	From 	string 	 `json:"from"`
	Message string 	 `json:"message"`
}

func (s *SMS) GetBytes() []byte {
	bytes, _ := json.Marshal(s)
	return bytes
}

type Email struct {
	To      []string `json:"to,omitempty"`
	From    string   `json:"from,omitempty"`
	Message []byte   `json:"message,omitempty"`
}

func (e *Email) GetBytes() []byte {
	bytes, _ := json.Marshal(e)
	return bytes
}
