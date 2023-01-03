package network

type ConnReal interface {
	Read(p []byte) (n int, err error)
	Write(b []byte) (n int, err error)
}
