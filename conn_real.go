package network

type ConnReal interface {
	Send()
	Receive()
	Read(p []byte) (n int, err error)
	Write(b []byte) (n int, err error)
}
