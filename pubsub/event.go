package pubsub

type Event struct {
	ID          string
	Headers     map[string]interface{}
	ContentType string
	Message     []byte
}
