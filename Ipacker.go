package network

type IPacker interface {
	Pack(msgID uint16, msg interface{}) ([]byte, error)
	Read(*TcpSession) ([]byte, error)
	Unpack([]byte) (*Message, error)
}
