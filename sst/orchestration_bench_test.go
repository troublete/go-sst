package sst

import (
	"testing"
	"time"
)

func BenchmarkOrchestration(b *testing.B) {
	b.Run("property and component validation", func(b *testing.B) {
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

		te := &testEntity{
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
		}

		for b.Loop() {
			_, _ = o.ValidateTransition(te, "booked")
		}
	})
}
