package main

import (
	"fmt"
	"github.com/phuhao00/network"
	"time"
)

func main() {

	kcp := network.NewKcp()
	go kcp.Connect()
	Data(kcp)
	c := network.NewKcp()
	go c.Dial()
	Data(c)
	select {}
}

func Data(k *network.KCP) {
	go func() {
		tk := time.NewTicker(time.Second * 10)
		for {
			select {
			case <-tk.C:
				k.ChOut <- []byte("im send data  mmmmmm")
			case data := <-k.ChIn:
				fmt.Println(string(data))

			}
		}
	}()
}
