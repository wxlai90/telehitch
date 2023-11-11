package models

type MessageType int

const (
	TextMessage MessageType = iota
	CallbackMessage
	CommandMessage
	UnknownMessageType
)
