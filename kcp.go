package network

import (
	"crypto/sha1"
	"github.com/xtaci/kcp-go/v5"
	"golang.org/x/crypto/pbkdf2"
	"log"
)

type KCP struct {
	key          []byte
	address      string `json:"address"`
	dataShards   int
	parityShards int
	KCPReal
	listener  *kcp.Listener
	ChOut     chan []byte
	ChIn      chan []byte
	msgParser *BufferPacker
}

func NewKcp() *KCP {
	return &KCP{
		key:          pbkdf2.Key([]byte("demo pass"), []byte("demo salt"), 1024, 32, sha1.New),
		address:      "127.0.0.1:12345",
		dataShards:   10,
		parityShards: 3,
		KCPReal:      nil,
		listener:     nil,
		ChIn:         make(chan []byte),
		ChOut:        make(chan []byte),
		msgParser:    newInActionPacker(),
	}
}

func (k *KCP) Connect() {
	block, _ := kcp.NewAESBlockCrypt(k.key)
	if listener, err := kcp.ListenWithOptions(k.address, block, k.dataShards, k.parityShards); err == nil {
		for {
			s, err := listener.AcceptKCP()
			if err != nil {
				log.Fatal(err)
			}
			k.handleEcho(s)
		}
	}

}

func (k *KCP) Dial() {
	block, _ := kcp.NewAESBlockCrypt(k.key)
	if sess, err := kcp.DialWithOptions(k.address, block, k.dataShards, k.parityShards); err == nil {
		k.handleEcho(sess)
	} else {
		log.Fatal(err)
	}
}

func (k *KCP) handleEcho(conn *kcp.UDPSession) {
	go k.handleRead(conn)
	go k.handleWrite(conn)
}

func (k *KCP) handleRead(conn *kcp.UDPSession) {
	for {
		data, err := k.msgParser.Read(conn)
		if err != nil {
			log.Println(err)
			return
		}
		k.ChIn <- data
	}
}

func (k *KCP) handleWrite(conn *kcp.UDPSession) {
	for {
		select {
		case data := <-k.ChOut:
			err := k.msgParser.Write(conn, data...)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}
