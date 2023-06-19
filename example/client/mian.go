package main

import (
	"fmt"
	"github.com/phuhao00/greatestworks-proto/gen/player"
	"github.com/phuhao00/network"
	"github.com/phuhao00/network/example/logger"
	"time"
)

func main() {
	client := network.NewClient(":8023", 200, logger.Logger)
	client.OnMessageCb = OnClientMessage
	client.Run()
	go Tick(client)
	select {}
}

func OnClientMessage(packet *network.Packet) {
	if packet.Msg.ID == 1 {
		fmt.Println("hello world")
	}
	time.Sleep(time.Second)
	packet.Conn.AsyncSend(3, &player.ChatMessage{
		Content: "abd",
		Extra:   nil,
	})
}

func Tick(c *network.TcpClient) {

	c.TcpSession.AsyncSend(3, &player.ChatMessage{
		Content: "abd",
		Extra:   nil,
	})

}
