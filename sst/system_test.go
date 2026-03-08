package sst

import (
	"sync"
	"testing"
	"time"
)

type testEntity map[string]any

func (t testEntity) ID() string {
	return t["_id"].(string)
}

func (t testEntity) Kind() string {
	return t["_kind"].(string)
}

func (t testEntity) Stage() string {
	return t["_stage"].(string)
}

func (t testEntity) Components(kind string) []Entity {
	if kind == "" {
		return t["_components"].([]Entity)
	}
	var cg []Entity
	for _, c := range t["_components"].([]Entity) {
		if c.Kind() == kind {
			cg = append(cg, c)
		}
	}
	return cg
}

func (t testEntity) HasProperty(name string) bool {
	_, ok := t[name]
	return ok
}

func (t testEntity) Property(name string) any {
	v, ok := t[name]
	if ok {
		return v
	}
	return nil
}

var _ Entity = testEntity{}

func NewTestEntity(id, kind, stage string, components ...Entity) testEntity {
	e := testEntity{}
	e["_id"] = id
	e["_kind"] = kind
	e["_stage"] = stage
	e["_components"] = components
	return e
}

func testOrchestration() *Orchestration {
	o := Orchestration{}

	o.For("order").Add(
		&Stage{Name: "Created"},
	).Add(&Stage{
		Name: "Booked",
		PropertyGate: map[string]any{
			"locationAssigned": Any(),
		},
		ComponentGate: []*ComponentGate{
			{
				StageName: "Allocated",
				N:         -1,
			},
		},
	})

	o.For("fulfilment").Add(
		&Stage{Name: "Created"},
	).Add(&Stage{
		Name: "Allocated",
		PropertyGate: map[string]any{
			"reservedQuantity": Any(),
		},
	})

	return &o
}

func testListener(response chan Gate, waitFor int) (chan bool, *[]Gate) {
	gates := []Gate{}
	done := make(chan bool)
	go func() {
		var wg sync.WaitGroup
		wg.Add(waitFor)
		go func() {
			for {
				select {
				case g := <-response:
					gates = append(gates, g)
					wg.Done()
				default:
				}
			}
		}()
		wg.Wait()
		time.Sleep(time.Millisecond * 50) // wait to see if there are more messages; which implies an errornous behaviour
		done <- true
	}()
	return done, &gates
}

func TestSST(t *testing.T) {
	t.Run("setting to initial state from nothing should pass gates", func(t *testing.T) {
		orch := testOrchestration()
		ord := NewTestEntity("ORD-1", "order", "")

		response := make(chan Gate)
		done, gp := testListener(response, 1)
		success := orch.Try(ord, "Created", response)
		<-done
		gates := *gp

		if success != true ||
			len(gates) != 1 ||
			gates[0].Passes != true ||
			gates[0].FromStage != "" ||
			gates[0].ToStage != "Created" ||
			gates[0].EntityKind != "order" ||
			gates[0].EntityID != "ORD-1" ||
			gates[0].Reason != "" ||
			gates[0].ReasonReference != "" {
			t.Error("expected pass")
		}
	})

	t.Run("setting to sequential stage without required properties should stop at gate", func(t *testing.T) {
		orch := testOrchestration()
		ord := NewTestEntity("ORD-1", "order", "Created")

		response := make(chan Gate)
		done, gp := testListener(response, 1)
		success := orch.Try(ord, "Booked", response)
		<-done
		gates := *gp

		if success == true ||
			len(gates) != 1 ||
			gates[0].Passes != false ||
			gates[0].FromStage != "Created" ||
			gates[0].ToStage != "Booked" ||
			gates[0].EntityKind != "order" ||
			gates[0].EntityID != "ORD-1" ||
			gates[0].Reason == "" ||
			gates[0].ReasonReference != "property_gate_not_passed,locationAssigned" {
			t.Error("expected fail")
		}
	})

	t.Run("setting to sequential stage with required properties should pass all gates", func(t *testing.T) {
		orch := testOrchestration()
		ord := NewTestEntity("ORD-1", "order", "Created")
		ord["locationAssigned"] = ""

		response := make(chan Gate)
		done, gp := testListener(response, 1)
		success := orch.Try(ord, "Booked", response)
		<-done
		gates := *gp

		if success == false ||
			len(gates) != 1 ||
			gates[0].Passes != true ||
			gates[0].FromStage != "Created" ||
			gates[0].ToStage != "Booked" ||
			gates[0].EntityKind != "order" ||
			gates[0].EntityID != "ORD-1" ||
			gates[0].Reason != "" ||
			gates[0].ReasonReference != "" {
			t.Error("expected pass")
		}
	})

	t.Run("component gates should also be referenced on entity level", func(t *testing.T) {
		orch := testOrchestration()
		ord := NewTestEntity("ORD-1", "order", "Created", NewTestEntity("FUL-1", "fulfilment", "Created"))
		ord["locationAssigned"] = ""

		response := make(chan Gate)
		done, gp := testListener(response, 2)
		success := orch.Try(ord, "Booked", response)
		<-done
		gates := *gp

		if success == true ||
			len(gates) != 2 {
			t.Error("expected fail")
		}

		if gates[0].Passes != false ||
			gates[0].FromStage != "Created" ||
			gates[0].ToStage != "Allocated" ||
			gates[0].EntityKind != "fulfilment" ||
			gates[0].EntityID != "FUL-1" ||
			gates[0].Reason == "" ||
			gates[0].ReasonReference != "property_gate_not_passed,reservedQuantity" {
			t.Error("expected fail")
		}

		if gates[1].Passes != false ||
			gates[1].FromStage != "Created" ||
			gates[1].ToStage != "Booked" ||
			gates[1].EntityKind != "order" ||
			gates[1].EntityID != "ORD-1" ||
			gates[1].Reason == "" ||
			gates[1].ReasonReference != "component_gate_not_passed,Allocated,-1" {
			t.Error("expected fail")
		}
	})

	t.Run("setting all requirements should pass", func(t *testing.T) {
		orch := testOrchestration()
		ff := NewTestEntity("FUL-1", "fulfilment", "Created")
		ff["reservedQuantity"] = 0
		ord := NewTestEntity("ORD-1", "order", "Created", ff)
		ord["locationAssigned"] = ""

		response := make(chan Gate)
		done, gp := testListener(response, 2)
		success := orch.Try(ord, "Booked", response)
		<-done
		gates := *gp

		if success == false ||
			len(gates) != 2 {
			t.Error("expected fail")
		}

		if gates[0].Passes != true ||
			gates[0].FromStage != "Created" ||
			gates[0].ToStage != "Allocated" ||
			gates[0].EntityKind != "fulfilment" ||
			gates[0].EntityID != "FUL-1" ||
			gates[0].Reason != "" ||
			gates[0].ReasonReference != "" {
			t.Error("expected pass")
		}

		if gates[1].Passes != true ||
			gates[1].FromStage != "Created" ||
			gates[1].ToStage != "Booked" ||
			gates[1].EntityKind != "order" ||
			gates[1].EntityID != "ORD-1" ||
			gates[1].Reason != "" ||
			gates[1].ReasonReference != "" {
			t.Error("expected pass")
		}
	})

	t.Run("checking for explicit value should work", func(t *testing.T) {
		o := Orchestration{}
		o.For("order").Add(
			&Stage{Name: "Created"},
		).Add(&Stage{
			Name: "Booked",
			PropertyGate: map[string]any{
				"locationAssigned": "yes",
			},
		})

		ord := NewTestEntity("ORD-1", "order", "Created")
		ord["locationAssigned"] = "yes"

		response := make(chan Gate)
		done, gp := testListener(response, 1)
		success := o.Try(ord, "Booked", response)
		<-done
		gates := *gp

		if success != true || len(gates) != 1 {
			t.Error("expected pass")
		}

		if gates[0].Passes != true ||
			gates[0].FromStage != "Created" ||
			gates[0].ToStage != "Booked" ||
			gates[0].EntityKind != "order" ||
			gates[0].EntityID != "ORD-1" ||
			gates[0].Reason != "" ||
			gates[0].ReasonReference != "" {
			t.Error("expected pass")
		}

		ordF := NewTestEntity("ORD-1", "order", "Created")
		ordF["locationAssigned"] = "wrong"

		response = make(chan Gate)
		done, gp = testListener(response, 1)
		success = o.Try(ordF, "Booked", response)
		<-done
		gates = *gp

		if success != false || len(gates) != 1 {
			t.Error("expected fail")
		}

		if gates[0].Passes != false ||
			gates[0].FromStage != "Created" ||
			gates[0].ToStage != "Booked" ||
			gates[0].EntityKind != "order" ||
			gates[0].EntityID != "ORD-1" ||
			gates[0].Reason == "" ||
			gates[0].ReasonReference != "property_gate_not_passed,locationAssigned" {
			t.Error("expected fail")
		}
	})

	t.Run("checking for required component amount should work", func(t *testing.T) {
		o := Orchestration{}
		o.For("order").Add(
			&Stage{Name: "Created"},
		).Add(&Stage{
			Name: "Booked",
			PropertyGate: map[string]any{
				"locationAssigned": Any(),
			},
			ComponentGate: []*ComponentGate{
				{
					StageName: "Allocated",
					N:         -1,
					Min:       1,
				},
			},
		})

		ord := NewTestEntity("ORD-1", "order", "Created")
		ord["locationAssigned"] = ""

		response := make(chan Gate)
		done, gp := testListener(response, 1)
		success := o.Try(ord, "Booked", response)
		<-done
		gates := *gp

		if success == true || len(gates) != 1 {
			t.Error("expected fail")
		}

		if gates[0].Passes != false ||
			gates[0].FromStage != "Created" ||
			gates[0].ToStage != "Booked" ||
			gates[0].EntityKind != "order" ||
			gates[0].EntityID != "ORD-1" ||
			gates[0].Reason == "" ||
			gates[0].ReasonReference != "component_gate_not_passed,Allocated,-1" {
			t.Error("expected fail")
		}
	})

	t.Run("checking for not available kind should fail", func(t *testing.T) {
		o := Orchestration{}

		response := make(chan Gate)
		done, gp := testListener(response, 1)
		success := o.Try(NewTestEntity("ord-1", "order", "created"), "Booked", response)
		<-done
		gates := *gp

		if success == true || len(gates) != 1 {
			t.Error("expected fail")
		}

		if gates[0].Passes != false ||
			gates[0].FromStage != "created" ||
			gates[0].ToStage != "Booked" ||
			gates[0].EntityKind != "order" ||
			gates[0].EntityID != "ord-1" ||
			gates[0].Reason == "" ||
			gates[0].ReasonReference != "sequence_not_available" {
			t.Error("expected fail")
		}
	})

	t.Run("checking for not available stage should fail", func(t *testing.T) {
		o := Orchestration{}
		o.For("order")
		response := make(chan Gate)
		done, gp := testListener(response, 1)
		success := o.Try(NewTestEntity("ord-1", "order", "created"), "Booked", response)
		<-done
		gates := *gp

		if success == true || len(gates) != 1 {
			t.Error("expected fail")
		}

		if gates[0].Passes != false ||
			gates[0].FromStage != "created" ||
			gates[0].ToStage != "Booked" ||
			gates[0].EntityKind != "order" ||
			gates[0].EntityID != "ord-1" ||
			gates[0].Reason == "" ||
			gates[0].ReasonReference != "stage_not_defined" {
			t.Error("expected fail")
		}
	})

	t.Run("checking for not available milestone should fail", func(t *testing.T) {
		o := Orchestration{}
		o.For("order").Add(&Stage{
			Name: "created",
		})
		response := make(chan Gate)
		done, gp := testListener(response, 1)
		success := o.Try(NewTestEntity("ord-1", "order", "created"), "Booked", response)
		<-done
		gates := *gp

		if success == true || len(gates) != 1 {
			t.Error("expected fail")
		}

		if gates[0].Passes != false ||
			gates[0].FromStage != "created" ||
			gates[0].ToStage != "Booked" ||
			gates[0].EntityKind != "order" ||
			gates[0].EntityID != "ord-1" ||
			gates[0].Reason == "" ||
			gates[0].ReasonReference != "milestone_stage_not_defined" {
			t.Error("expected fail")
		}
	})

	t.Run("checking for wrong order should fail", func(t *testing.T) {
		o := Orchestration{}
		o.For("order").Add(&Stage{
			Name: "created",
		}).Add(&Stage{
			Name: "booked",
		})

		response := make(chan Gate)
		done, gp := testListener(response, 1)
		success := o.Try(NewTestEntity("ord-1", "order", "booked"), "created", response)
		<-done
		gates := *gp

		if success == true || len(gates) != 1 {
			t.Error("expected fail")
		}

		if gates[0].Passes != false ||
			gates[0].FromStage != "booked" ||
			gates[0].ToStage != "created" ||
			gates[0].EntityKind != "order" ||
			gates[0].EntityID != "ord-1" ||
			gates[0].Reason == "" ||
			gates[0].ReasonReference != "milestone_stage_in_the_past" {
			t.Error("expected fail")
		}
	})

	t.Run("complex passing test; ord + 2 ful (one pass, one fail)", func(t *testing.T) {
		o := Orchestration{}
		o.For("order").Add(&Stage{
			Name: "created",
		}).Add(&Stage{
			Name: "booked",
			ComponentGate: []*ComponentGate{
				{
					Kind:      "fulfilment",
					StageName: "allocated",
					N:         -1,
					Min:       1,
				},
			},
		})

		o.For("fulfilment").Add(&Stage{Name: "not-allocated"}).Add(&Stage{
			Name: "allocated",
			PropertyGate: map[string]any{
				// this is important; if there are no gates defined, the transition is always possible,
				// therefore can a component though, initialized in the wrong state be moved forward
				// making the following top level gate valid
				"n": Any(),
			},
		})

		f := NewTestEntity("ful-1", "fulfilment", "not-allocated")
		f["n"] = 0

		response := make(chan Gate)
		done, gp := testListener(response, 3)
		success := o.Try(NewTestEntity("ord-1", "order", "created",
			f,
			NewTestEntity("ful-2", "fulfilment", "not-allocated"),
		), "booked",
			response,
		)
		<-done
		gates := *gp

		if success == true || len(gates) != 3 {
			t.Error("expected fail")
		}

		if gates[0].Passes != true ||
			gates[0].FromStage != "not-allocated" ||
			gates[0].ToStage != "allocated" ||
			gates[0].EntityKind != "fulfilment" ||
			gates[0].EntityID != "ful-1" ||
			gates[0].Reason != "" ||
			gates[0].ReasonReference != "" {
			t.Error("expected pass")
		}

		if gates[1].Passes != false ||
			gates[1].FromStage != "not-allocated" ||
			gates[1].ToStage != "allocated" ||
			gates[1].EntityKind != "fulfilment" ||
			gates[1].EntityID != "ful-2" ||
			gates[1].Reason == "" ||
			gates[1].ReasonReference != "property_gate_not_passed,n" {
			t.Error("expected fail")
		}

		if gates[2].Passes != false ||
			gates[2].FromStage != "created" ||
			gates[2].ToStage != "booked" ||
			gates[2].EntityKind != "order" ||
			gates[2].EntityID != "ord-1" ||
			gates[2].Reason == "" ||
			gates[2].ReasonReference != "component_gate_not_passed,allocated,-1" {
			t.Error("expected fail")
		}
	})

	t.Run("filtered component gate should be possible", func(t *testing.T) {
		o := Orchestration{}
		o.For("order").Add(&Stage{
			Name: "created",
		}).Add(&Stage{
			Name: "booked",
			ComponentGate: []*ComponentGate{
				{
					Kind:      "fulfilment",
					StageName: "allocated",
					N:         -1,
					Min:       1,
				},
			},
		})

		o.For("fulfilment").Add(&Stage{Name: "not-allocated"}).Add(&Stage{
			Name: "allocated",
		})

		f := NewTestEntity("ful-1", "fulfilment", "not-allocated")

		response := make(chan Gate)
		done, gp := testListener(response, 2)
		success := o.Try(NewTestEntity("ord-1", "order", "created",
			f,
			NewTestEntity("comp-1", "component", "not-allocated"),
		), "booked",
			response,
		)
		<-done
		gates := *gp

		if success == false || len(gates) != 2 {
			t.Error("expected pass")
		}

		if gates[0].Passes != true ||
			gates[0].FromStage != "not-allocated" ||
			gates[0].ToStage != "allocated" ||
			gates[0].EntityKind != "fulfilment" ||
			gates[0].EntityID != "ful-1" ||
			gates[0].Reason != "" ||
			gates[0].ReasonReference != "" {
			t.Error("expected pass")
		}

		if gates[1].Passes != true ||
			gates[1].FromStage != "created" ||
			gates[1].ToStage != "booked" ||
			gates[1].EntityKind != "order" ||
			gates[1].EntityID != "ord-1" ||
			gates[1].Reason != "" ||
			gates[1].ReasonReference != "" {
			t.Error("expected pass")
		}
	})
}
