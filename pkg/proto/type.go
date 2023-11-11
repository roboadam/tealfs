package proto

const (
	NodeSync = uint8(2)
)

func Hello() NetCmd {
	return NetCmd{Value: 1}
}

type NetCmd struct {
	Value uint8
}
