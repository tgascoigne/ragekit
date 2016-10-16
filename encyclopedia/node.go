package encyclopedia

type Node interface {
	Label() string
	Properties() map[string]interface{}
}
