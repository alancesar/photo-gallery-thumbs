package listener

type Message struct {
	ID         string
	Attributes map[string]string
	Data       []byte
}
