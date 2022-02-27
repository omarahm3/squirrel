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
	PeerId string `json:"peerId"`
	Local  bool   `json:"local"`
}
