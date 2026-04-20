# go-sst
> State flow analysis of hierarchical entity structures

[![test](https://github.com/troublete/go-sst/actions/workflows/test.yml/badge.svg)](https://github.com/troublete/go-sst/actions/workflows/test.yml)

## Introduction

In systems where the correct aggregated state is dependent on correct states of all elements within itself and the
mutation of element state depend on the valid state of the element in question and hierarchical related elements, it
becomes necessary to validate or affirm changes on an almost holistic level to assure keeping the system correct.

For systems whose process can be defined linearly, this library is designed to do exactly that while providing the basis
upon which data changes may be orchestrated.

Systems where this applies are for instance e-commerce systems, where the status progression from an order from "
processing" to "complete" is not merely dependent on its own datapoints (e.g. payment_received=true) but also on the
states of articles within that order which all ought to be delivered to facilitate an order being complete. Based on
that example it becomes clear that other systems where this might become handy are (industrial) manufacturing or even
generic system state analysis, anything where state is split into multiple domain elements which are correlated.

The library is build around a generic interface, to capture the whole system structure, which is defined as seen below.

```go
type Entity interface {
	ID() string // should return some form of id; sent back in responses as reference
	Kind() string // this is required to map the entity, to a sequence within the defined orchestration; sent back in response
	Stage() string // this must return the current stage of the entity on the sequence in the orchestration

	Properties() map[string]string // the properties surface of the entities; we limit here to string to make the validation predictable
	Components() []Entity // related components which are also of type Entity
}
```

Any validation/evaluation on the defined orchestration can be on bootup or runtime, but the definition of an
orchestration with all its sequences should be defined on bootup and never on runtime.

Any validation returns two variables, the first being if the overall process (the root elements stage transition + all
required component stage transitions) do work (in theory) and secondly all events that happened during the evaluation,
those can be for instance successful stage passes, problems like failing property or component requirements and
definition errors that make it impossible to determine state. The returned data can be used as a baseline for further
processing or data mutation on the elements in question as the order is deterministic.

It shall be noted that the library can confirm that datapoints necessary for a transition are
correct, but does not fix issues. It acts on the idea that progression happens idempotently, based on existing
element data.

## License

All rights reserved.