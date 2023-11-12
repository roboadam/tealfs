package proto

func Hello() NetCmd {
	return NetCmd{Value: 1}
}

func NodeSync() NetCmd {
	return NetCmd{Value: 2}
}

type NetCmd struct {
	Value uint8
}
