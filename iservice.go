package network

type ISession interface {
	OnConnect()
	OnClose()
	OnMessage(*Message, *TcpSession)
}
