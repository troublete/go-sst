package main

import (
	"fmt"
	"time"

	"github.com/troublete/go-sst/sst"
)

type Order struct {
	Ref        string
	Status     string
	Articles   []*Article
	attributes map[string]any
}

func (o *Order) ID() string    { return o.Ref }
func (o *Order) Kind() string  { return "order" }
func (o *Order) Stage() string { return o.Status }
func (o *Order) Components(_ string) []sst.Entity {
	var entities []sst.Entity
	for _, a := range o.Articles {
		entities = append(entities, a)
	}
	return entities
}
func (o *Order) HasProperty(name string) bool {
	if _, ok := o.attributes[name]; ok {
		return true
	}
	return false
}
func (o *Order) Property(name string) any {
	if v, ok := o.attributes[name]; ok {
		return v
	}
	return nil
}

var _ sst.Entity = &Order{}

type Article struct {
	Ref        string
	Status     string
	attributes map[string]any
}

func (a *Article) ID() string                       { return a.Ref }
func (a *Article) Kind() string                     { return "article" }
func (a *Article) Stage() string                    { return a.Status }
func (a *Article) Components(_ string) []sst.Entity { return nil }
func (a *Article) HasProperty(name string) bool {
	if _, ok := a.attributes[name]; ok {
		return true
	}
	return false
}
func (a *Article) Property(name string) any {
	if v, ok := a.attributes[name]; ok {
		return v
	}
	return nil
}

var _ sst.Entity = &Article{}

func main() {
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
				Kind:      "articles",
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

	order := &Order{
		Ref:    "order-1",
		Status: "", // if left empty will presume to initial state of sequence, in this case 'in_progress'
		Articles: []*Article{
			{
				Ref:    "article-1",
				Status: "shipped",
				attributes: map[string]any{
					"deliveredAt": time.Now(),
				},
			},
		},
		attributes: nil,
	}

	response := make(chan sst.Gate)
	go func() {
		for {
			select {
			case r := <-response:
				// receives 3 responses;
				//		1 that confirms that order (because started on "") can move into in_progress
				//		1 that confirms article can move into delivered,
				//		1 that denies the order can move to done, because of a property gate (paymentReceived missing)

				// ... do something with the event;
				fmt.Println(r)
			default:
			}
		}
	}()

	done := make(chan bool)
	success := o.Try(order, "done", response) // attempt to move order to 'done'
	fmt.Println(success)                      // should output false, because the article can move, but the order can't
	<-done                                    // just block for the demo
}
