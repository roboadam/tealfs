package node

type Address struct {
	value string
}

func NewAddress(rawValue string) Address {
	return Address{value: rawValue}
}
