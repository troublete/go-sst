# go-sst
> State flow analysis of hierarchical entity structures

## Introduction

This library provides a solid-state like linear flow analysis of processes in a hierarchical entity structure i.e. it
gates the linear progression of an entity by property checks and checks on sub-entities i.e. components. It is quite
useful for the validation and processes where the progression of the 'parent' entity not solely depends on its own
internal state but also on 'child' states, for instance: how many sub-entities of type x are already in stage y;

It can be used for any process where the state progression of entities can be linearly defined: for instance ecommerce,
manufacturing, system state verification and so on. A good example is the flow of an order, where the progression into '
done' is not only determined based on the internal state of the order (e.g. payment received) but also on the articles
of that order (e.g. are all articles stock reserved).

Any evaluation on the defined orchestration (should be defined on bootup and evaluations run on runtime) returns a
boolean value if the whole transition proposed would work and and a list containing all messages declaring what happened
along the way e.g. if a article of an order could progress it will be output even if the overall process would fail
according to the validation. In the following the output messages can be used as orchestration of state updates on the
actual entities.

The library is build around a generic interface which is defined as seen below.

```go
type Entity interface {
	ID() string // should return some form of id; also sent back in responses as reference
	Kind() string // this is required to map the entity, to a sequence within the defined orchestration
	Stage() string // this must return the current stage of the entity on the sequence in the orchestration

	Properties() map[string]string // all properties of the entities, we limit here to string to make the validation predictable
	Components() []Entity // all components entities of the 'parent'
}
```

It shall be noted that the library aims for linear progression correctness, the datapoints on which it relies must be
set before attempting a progression so the library can confirm that every datapoint necessary to move any entity is
available i.e. it is recommended to decouple data state from process state so that data is always collected but the
progression is decided upon if enough data is available to act.

## License

All rights reserved.