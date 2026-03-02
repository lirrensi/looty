package clipboard

import (
	"encoding/json"
)

type MessageType string

const (
	TypeClipboard  MessageType = "clipboard"
	TypeRefresh    MessageType = "refresh"
	TypeScratchpad MessageType = "scratchpad"
)

type Message struct {
	Type MessageType `json:"type"`
	Data string      `json:"data"`
}

func ParseMessage(data []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func NewClipboardMessage(text string) []byte {
	msg := Message{
		Type: TypeClipboard,
		Data: text,
	}
	data, _ := json.Marshal(msg)
	return data
}

func NewRefreshMessage() []byte {
	msg := Message{
		Type: TypeRefresh,
		Data: "",
	}
	data, _ := json.Marshal(msg)
	return data
}

func NewScratchpadMessage(content string) []byte {
	msg := Message{
		Type: TypeScratchpad,
		Data: content,
	}
	data, _ := json.Marshal(msg)
	return data
}
