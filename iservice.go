package network

type IService interface {
	OnConnect()
	OnClose()
	OnMessage(*Message, *TcpSession)
}
