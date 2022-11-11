package network

type Session struct {
	SessionReal
}

func NewSession(real SessionReal) *Session {
	s := &Session{
		SessionReal: real,
	}
	return s
}
