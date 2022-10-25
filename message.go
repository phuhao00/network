package network

import "sync"

type Message struct {
	ID   uint64
	Data []byte
}

var msgPool = sync.Pool{
	New: func() interface{} {
		return &Message{}
	},
}

// GetPooledMessage gets a pooled Message.
func GetPooledMessage() *Message {
	return msgPool.Get().(*Message)
}

// FreeMessage puts a Message into the pool.
func FreeMessage(msg *Message) {
	if msg != nil && len(msg.Data) > 0 {
		ResetMessage(msg)
		msgPool.Put(msg)
	}
}

//ResetMessage reset a Message
func ResetMessage(m *Message) {
	m.Data = m.Data[:0]
}
