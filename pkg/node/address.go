package node

type Address struct {
	Value string
}

func NewAddress(rawValue string) Address {
	return Address{Value: rawValue}
}
