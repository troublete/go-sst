# go-sst
> Solid-state transitions for entities, with recursive component gates

## Introduction

This library provides a (solid-)state machine for linear processes in a hierarchical entity structure i.e. it gates the
linear progression of an entity by property checks and recursive checks on sub-entities. It is quite useful for the
validation and processes where the progression of the 'parent' entity not solely depends on its own internal state
but also on child states, for instance: how many sub-entities of type x are already in stage y; 

It can be used for instance within implementations for e-commerce where the progression of an order, not only depends on
the order itself (e.g. payment received) but also on the articles of that order (e.g. are all articles stock reserved).

It communicates passes or blocks of state progression, with full reason explanation, on a go channel which is user
defined. Implementing an execution runner based on the responses should be trivial, as the communication is verbose
enough for automatic processing.

The library is build around a generic interface which is defined as seen below.

```go
type Entity interface {
	ID() string // should return some form of id; also sent back in responses as reference
	Kind() string // this is required to map the entity, to a sequence within the defined orchestration
	Stage() string // this must return the current stage of the entity on the sequence in the orchestration

	Components(kind string) []Entity // returns all "components" of the entity; it is required that kind="" returns all components regardless of kind

	HasProperty(name string) bool // returns true if a property is available in the state of the entity
	Property(name string) any // returns the value of the property in the state of the entity
}
```

It shall be noted that the library aims for linear progression correctness, the datapoints on which it relies must be
set before attempting a progression so the library can confirm that every datapoint necessary to move an entity is
available i.e. it is recommended to decouple data state from process state so that data is always collected but the
progression is decided upon if enough data is available to act.

Caution: As the library channel writes are blocking, it is recommended for production use to have a fan-in pattern setup
where each 'Try' response is then routed to a unified buffered channel in a non-blocking manner.

## Quickstart

```go
// orchestration defined once in code; every kind has it own linear flow (sequence) of stages
o := sst.Orchestration{}

// orders can go from in_progress -> done, if property 'paymentReceived' is set and all articles are delivered
o.For("order").Add(&sst.Stage{
    Name: "in_progress",
}).Add(&sst.Stage{
    Name: "done",
    PropertyGate: map[string]any{
        "paymentReceived": sst.Any(),
    },
    ComponentGate: []*sst.ComponentGate{
        {
            Kind:      "articles", // filter by kind, if empty all components are considered
            StageName: "delivered",
            N:         -1, // all
            Min:       1,  // requires at least one
        },
    },
})

// articles can go from shipped -> delivered, if 'deliveredAt' is set
o.For("article").Add(&sst.Stage{
    Name: "shipped",
}).Add(&sst.Stage{
    Name: "delivered",
    PropertyGate: map[string]any{
        "deliveredAt": sst.Any(),
    },
})

// define listener
response := make(chan sst.Gate)
go func() {
    for {
        select{
        case r := <-response
            // do something with the gate response, i.e. update data, trigger work, ...
        default:
        }
    }
}()

success := o.Try(entityToCheck, "to_stage", response) 
// success is bool, indicating if the move, considering the current data state is successful
```

see `/demo` directory for extended example and how to use.

## License

All rights reserved.