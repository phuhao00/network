package network

type Listener interface {
	Accept() (SessionReal, error)
}
