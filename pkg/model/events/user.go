package events

type Ui struct {
	EventType Type
	Argument  string
}

type Type int

const (
	ConnectTo Type = iota
	AddStorage
)
