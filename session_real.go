package network

type SessionReal interface {
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
}
