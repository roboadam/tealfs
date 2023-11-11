package proto

const (
	Hello    = uint8(1)
	NodeSync = uint8(2)
)
 
type NetCmd struct {
	Value uint8
}