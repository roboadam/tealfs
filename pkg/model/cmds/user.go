package cmds

type User struct {
	CmdType Type
	Argument string
}

type Type int

const (
	ConnectTo Type = iota
	AddStorage
)
