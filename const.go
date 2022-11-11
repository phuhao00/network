package network

type ProtocolCategory string

const (
	ProtocolCategoryTCP       ProtocolCategory = "TCP"
	ProtocolCategoryUDP                        = "UDP"
	ProtocolCategoryKPC                        = "KCP"
	ProtocolCategoryQUIC                       = "QUIC"
	ProtocolCategoryWebSocket                  = "WebSocket"
	ProtocolCategoryHttp                       = "Http"
)
