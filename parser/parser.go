package parser

// Parser is ...
type Parser interface {
	// Parse
	Parse(src string) (attrs []string, err error)
}
