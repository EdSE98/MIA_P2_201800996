package commands

const CommentCommand = "#"

type Command struct {
	Name   string
	Params map[string]string
	Flags  map[string]bool
	Raw    string
	Line   int
}

func (c Command) IsComment() bool {
	return c.Name == CommentCommand
}
