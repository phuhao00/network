package network

import (
	"log"
	"sync"
)

type Hub struct {
	l        Listener
	Sessions sync.Map
	Owner    interface{}
}

func NewHub() *Hub {
	return &Hub{}
}

func (h *Hub) loop() {
	for {
		s, err := h.l.Accept()
		if err != nil {
			log.Fatalf(err.Error())
		}
		session := NewSession(s)
		h.Sessions.Store(session, struct{}{})
		go h.Handle(session)
	}
}

func (h *Hub) monitor() {

}

func (h *Hub) Handle(session *Session) {
	buf := make([]byte, 4096)
	for {
		n, err := session.Read(buf)
		if err != nil {
			log.Println(err)
			return
		}

		n, err = session.Write(buf[:n])
		if err != nil {
			log.Println(err)
			return
		}
	}
}
