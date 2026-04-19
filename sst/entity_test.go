package sst

type testEntity struct {
	id, kind, stage string
	properties      map[string]string
	components      []Entity
}

func (t *testEntity) ID() string {
	return t.id
}

func (t testEntity) Kind() string {
	return t.kind
}

func (t testEntity) Stage() string {
	return t.stage
}

func (t testEntity) Properties() map[string]string {
	return t.properties
}

func (t testEntity) Components() []Entity {
	return t.components
}

var _ Entity = &testEntity{}
