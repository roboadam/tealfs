package events

type Event struct {
	EventType Type
	argument  []byte
	result    chan []byte
}

func (e *Event) GetString() string {
	return string(e.argument)
}

func (e *Event) GetBytes() []byte {
	return e.argument
}

func (e *Event) GetResult() chan []byte {
	return e.result
}

func NewString(typ Type, argument string) Event {
	return Event{EventType: typ, argument: []byte(argument)}
}

func NewBytes(typ Type, argument []byte) Event {
	return Event{EventType: typ, argument: argument}
}

func NewBytesWithResult(typ Type, argument []byte, result chan []byte) Event {
	e := NewBytes(typ, argument)
	e.result = result
	return e
}

type Type int

const (
	ConnectTo Type = iota
	AddData
	ReadData
)
