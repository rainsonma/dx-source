package realtime

type Op string

const (
	OpSubscribe   Op = "subscribe"
	OpUnsubscribe Op = "unsubscribe"
	OpEvent       Op = "event"
	OpAck         Op = "ack"
	OpError       Op = "error"
)

type Envelope struct {
	Op      Op     `json:"op"`
	Topic   string `json:"topic,omitempty"`
	Type    string `json:"type,omitempty"`
	Data    any    `json:"data,omitempty"`
	ID      string `json:"id,omitempty"`
	OK      *bool  `json:"ok,omitempty"`
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type Event struct {
	Type string
	Data any
}
