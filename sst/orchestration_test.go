package sst

import (
	"strings"
	"testing"
	"time"
)

func TestOrchestration(t *testing.T) {
	t.Run("property gate pass expected if entity matches criteria", func(t *testing.T) {
		o := NewOrchestration()
		o.SequenceFor("order").AddStage(&Stage{
			Name: "created",
		}).AddStage(&Stage{
			Name: "booked",
			PropertyGates: []*PropertyGateDefinition{
				NewPropertyGate(
					map[string]string{
						"transmitted_at": "",
					},
				),
			},
		})

		success, messages := o.ValidateTransition(&testEntity{
			id:    "ORD-1",
			kind:  "order",
			stage: "created",
			properties: map[string]string{
				"transmitted_at": time.Now().Format(time.RFC3339),
			},
		}, "booked")

		if success != true || len(messages) != 1 {
			t.Error("failed to reach successful next stage")
		}
	})

	t.Run("property gate fail expected if entity doesn't match criteria", func(t *testing.T) {
		o := NewOrchestration()
		o.SequenceFor("order").AddStage(&Stage{
			Name: "created",
		}).AddStage(&Stage{
			Name: "booked",
			PropertyGates: []*PropertyGateDefinition{
				NewPropertyGate(
					map[string]string{
						"transmitted_at": "",
					},
				),
			},
		})

		success, messages := o.ValidateTransition(&testEntity{
			id:         "ORD-1",
			kind:       "order",
			stage:      "created",
			properties: map[string]string{},
		}, "booked")

		if success == true || len(messages) != 2 {
			t.Error("expected to fail on missing property")
		}
	})

	t.Run("component gate pass expected if entity matches criteria", func(t *testing.T) {
		o := NewOrchestration()
		o.SequenceFor("order").AddStage(&Stage{
			Name: "created",
		}).AddStage(&Stage{
			Name: "booked",
			PropertyGates: []*PropertyGateDefinition{
				NewPropertyGate(
					map[string]string{
						"transmitted_at": "",
					},
				),
			},
			ComponentGates: []*ComponentGateDefinition{
				NewComponentGate("article", []string{"complete", "transmitted"}, "", 1, 1),
			},
		})

		success, messages := o.ValidateTransition(&testEntity{
			id:    "ORD-1",
			kind:  "order",
			stage: "created",
			properties: map[string]string{
				"transmitted_at": time.Now().Format(time.RFC3339),
			},
			components: []Entity{
				&testEntity{
					id:    "ART-1",
					kind:  "article",
					stage: "transmitted",
				},
			},
		}, "booked")

		if success != true || len(messages) != 1 {
			t.Error("failed to reach successful next stage")
		}
	})

	t.Run("component gate pass expected if entity matches criteria & component can be made matching", func(t *testing.T) {
		o := NewOrchestration()
		o.SequenceFor("order").AddStage(&Stage{
			Name: "created",
		}).AddStage(&Stage{
			Name: "booked",
			PropertyGates: []*PropertyGateDefinition{
				NewPropertyGate(
					map[string]string{
						"transmitted_at": "",
					},
				),
			},
			ComponentGates: []*ComponentGateDefinition{
				NewComponentGate("article", []string{"complete", "transmitted"}, "transmitted", 1, 1),
			},
		})
		o.SequenceFor("article").AddStage(&Stage{
			Name: "created",
		}).AddStage(&Stage{
			Name: "transmitted",
			PropertyGates: []*PropertyGateDefinition{
				NewPropertyGate(map[string]string{"transmitted_at": ""}),
			},
		})

		success, messages := o.ValidateTransition(&testEntity{
			id:    "ORD-1",
			kind:  "order",
			stage: "created",
			properties: map[string]string{
				"transmitted_at": time.Now().Format(time.RFC3339),
			},
			components: []Entity{
				&testEntity{
					id:    "ART-1",
					kind:  "article",
					stage: "created",
					properties: map[string]string{
						"transmitted_at": time.Now().Format(time.RFC3339),
					},
				},
			},
		}, "booked")

		if success != true || len(messages) != 2 {
			t.Error("failed to reach successful next stage for order and article")
		}
	})

	t.Run("component gate fail expected if entity does not match criteria", func(t *testing.T) {
		o := NewOrchestration()
		o.SequenceFor("order").AddStage(&Stage{
			Name: "created",
		}).AddStage(&Stage{
			Name: "booked",
			PropertyGates: []*PropertyGateDefinition{
				NewPropertyGate(
					map[string]string{
						"transmitted_at": "",
					},
				),
			},
			ComponentGates: []*ComponentGateDefinition{
				NewComponentGate("article", []string{"complete", "transmitted"}, "", 1, 1),
			},
		})

		success, messages := o.ValidateTransition(&testEntity{
			id:    "ORD-1",
			kind:  "order",
			stage: "created",
			properties: map[string]string{
				"transmitted_at": time.Now().Format(time.RFC3339),
			},
			components: []Entity{
				&testEntity{
					id:    "ART-1",
					kind:  "article",
					stage: "booked",
				},
			},
		}, "booked")

		if success == true || len(messages) != 2 {
			t.Error("expected to fail on failing component gate")
		}
	})

	t.Run("component gate and property gate fail expected if entity does not match criteria", func(t *testing.T) {
		o := NewOrchestration()
		o.SequenceFor("order").AddStage(&Stage{
			Name: "created",
		}).AddStage(&Stage{
			Name: "booked",
			PropertyGates: []*PropertyGateDefinition{
				NewPropertyGate(
					map[string]string{
						"transmitted_at": "",
					},
				),
			},
			ComponentGates: []*ComponentGateDefinition{
				NewComponentGate("article", []string{"complete", "transmitted"}, "", 1, 1),
			},
		})

		success, messages := o.ValidateTransition(&testEntity{
			id:         "ORD-1",
			kind:       "order",
			stage:      "created",
			properties: map[string]string{},
			components: []Entity{
				&testEntity{
					id:    "ART-1",
					kind:  "article",
					stage: "booked",
				},
			},
		}, "booked")

		if success == true || len(messages) != 3 {
			t.Error("expected to fail on failing component gate and property gate")
		}
	})

	t.Run("logic AND gate partial fail should add logic gate specific error and still fail", func(t *testing.T) {
		o := NewOrchestration()
		o.SequenceFor("order").AddStage(&Stage{
			Name: "created",
		}).AddStage(&Stage{
			Name: "booked",
			PropertyGates: []*PropertyGateDefinition{
				NewPropertyGate(
					map[string]string{
						"transmitted_at": "",
					},
					"g1",
				),
			},
			ComponentGates: []*ComponentGateDefinition{
				NewComponentGate("article", []string{"transmitted", "complete"}, "complete", 1, 1, "g1"),
			},
			LogicGates: []*LogicGateDefinition{
				And("g1"),
			},
		})
		o.SequenceFor("article").AddStage(&Stage{
			Name: "created",
		}).AddStage(&Stage{
			Name: "transmitted",
			PropertyGates: []*PropertyGateDefinition{
				NewPropertyGate(map[string]string{"transmitted_at": ""}),
			},
		}).AddStage(&Stage{
			Name: "complete",
		})

		success, messages := o.ValidateTransition(&testEntity{
			id:    "ORD-1",
			kind:  "order",
			stage: "created",
			properties: map[string]string{
				"transmitted_at": time.Now().Format(time.RFC3339),
			},
			components: []Entity{
				&testEntity{
					id:    "ART-1",
					kind:  "article",
					stage: "created",
				},
			},
		}, "booked")

		if success == true || len(messages) != 6 {
			t.Error("expected fail to reach milestone")
		}
	})

	t.Run("logic OR gate partial fail should add logic gate specific error and pass", func(t *testing.T) {
		o := NewOrchestration()
		o.SequenceFor("order").AddStage(&Stage{
			Name: "created",
		}).AddStage(&Stage{
			Name: "booked",
			PropertyGates: []*PropertyGateDefinition{
				NewPropertyGate(
					map[string]string{
						"transmitted_at": "",
					},
					"g1",
				),
			},
			ComponentGates: []*ComponentGateDefinition{
				NewComponentGate("article", []string{"transmitted", "complete"}, "complete", 1, 1, "g1"),
			},
			LogicGates: []*LogicGateDefinition{
				Or("g1"),
			},
		})
		o.SequenceFor("article").AddStage(&Stage{
			Name: "created",
		}).AddStage(&Stage{
			Name: "transmitted",
			PropertyGates: []*PropertyGateDefinition{
				NewPropertyGate(map[string]string{"transmitted_at": ""}),
			},
		}).AddStage(&Stage{
			Name: "complete",
		})

		success, messages := o.ValidateTransition(&testEntity{
			id:    "ORD-1",
			kind:  "order",
			stage: "created",
			properties: map[string]string{
				"transmitted_at": time.Now().Format(time.RFC3339),
			},
			components: []Entity{
				&testEntity{
					id:    "ART-1",
					kind:  "article",
					stage: "created",
				},
			},
		}, "booked")

		if success != true || len(messages) != 5 {
			t.Error("expected to reach milestone")
		}
	})

	t.Run("fails early if sequence does not exist", func(t *testing.T) {
		o := NewOrchestration()
		success, messages := o.ValidateTransition(&testEntity{
			kind: "not-registered",
		}, "to-state")

		if success != false || len(messages) != 1 || !strings.Contains(messages[0], SequenceNoMatch) {
			t.Error("expected to fail")
		}
	})

	t.Run("fails early if sequence stage does not exist", func(t *testing.T) {
		o := NewOrchestration()
		o.SequenceFor("kind")
		success, messages := o.ValidateTransition(&testEntity{
			kind:  "kind",
			stage: "from-state",
		}, "to-state")

		if success != false || len(messages) != 1 || !strings.Contains(messages[0], SequenceStageNoMatch) || !strings.Contains(messages[0], "stage") {
			t.Error("expected to fail")
		}
	})

	t.Run("fails early if sequence milestone does not exist", func(t *testing.T) {
		o := NewOrchestration()
		o.SequenceFor("kind").AddStage(&Stage{
			Name: "from-state",
		})
		success, messages := o.ValidateTransition(&testEntity{
			kind:  "kind",
			stage: "from-state",
		}, "to-state")

		if success != false || len(messages) != 1 || !strings.Contains(messages[0], SequenceStageNoMatch) || !strings.Contains(messages[0], "milestone") {
			t.Error("expected to fail")
		}
	})

	t.Run("fails for faulty sequence", func(t *testing.T) {
		a := &Stage{
			Name: "a",
		}
		b := &Stage{
			Name: "b",
		}
		s := &Sequence{
			stages: map[string]*Stage{
				"a": a,
				"b": b,
			},
			stageSlice: []*Stage{
				a, b,
			},
		}
		o := &Orchestration{
			sequences: map[string]*Sequence{
				"faulty": s,
			},
		}

		success, messages := o.ValidateTransition(&testEntity{
			kind:  "faulty",
			stage: "a",
		}, "b")

		if success || len(messages) != 1 || !strings.Contains(messages[0], SequenceMilestoneNotReached) {
			t.Error("expected to flag invalid sequence structure")
		}
	})
}
