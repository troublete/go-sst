package sst

type Message struct {
	Content  string
	ID, Kind string
	Context  map[string]string
}
