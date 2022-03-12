package photo

type (
	Tag struct {
		ID       uint16
		TypeName string
		Count    uint32
		Value    string
	}

	Path map[string]Tag
	Exif map[string]Path
)

func (e Exif) SetTag(path, name string, tag Tag) {
	if e[path] == nil {
		e[path] = Path{}
	}

	e[path][name] = tag
}
