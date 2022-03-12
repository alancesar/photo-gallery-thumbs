package photo

type Photo struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	Metadata
}
