package main

import (
	"fmt"
	"github.com/phuhao00/greatestworks-proto/gen/player"
	"github.com/phuhao00/network"
	"github.com/phuhao00/network/example/logger"
	"time"
)

func main() {
	server := network.NewServer(":8023", 100, 200, logger.Logger)
	server.MessageHandler = OnMessage
	go server.Run()
	select {}
}

func OnMessage(packet *network.Packet) {
	time.Sleep(time.Second)

	if packet.Msg.ID == 3 {
		fmt.Println("hello hao")
	}
	packet.Conn.AsyncSend(
		1,
		&player.SCSendChatMsg{},
	)
}
