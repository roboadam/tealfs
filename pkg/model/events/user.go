package events

type Event struct {
	EventType Type
	argument  []byte
}

func (e *Event) GetString() string {
	return string(e.argument)
}

func (e *Event) GetBytes() []byte {
	return e.argument
}

func NewString(typ Type, argument string) Event {
	return Event{EventType: typ, argument: []byte(argument)}
}

func NewBytes(typ Type, argument []byte) Event {
	return Event{EventType: typ, argument: argument}
}

type Type int

const (
	ConnectTo Type = iota
	AddStorage
	AddData
)
