package server

type Message struct {
	Id      string      `json:"id"`
	Payload interface{} `json:"payload"`
	Event   string      `json:"event"`
}

type LogMessage struct {
	Line string `json:"line"`
}

type IdentityMessage struct {
	PeerId      string `json:"peerId"`
	Broadcaster bool   `json:"broadcaster"`
	Subscriber  bool   `json:"subscriber"`
}
