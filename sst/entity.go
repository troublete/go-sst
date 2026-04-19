package sst

type Entity interface {
	ID() string

	Kind() string
	Stage() string

	Properties() map[string]string
	Components() []Entity
}
