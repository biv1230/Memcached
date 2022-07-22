package server

var ConnMap map[string]Conner

type Conner interface {
	Name()
}
