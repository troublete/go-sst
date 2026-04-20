package sst

import "testing"

func TestComponentGate(t *testing.T) {
	t.Run("should be able to tell if passes (softcheck)", func(t *testing.T) {
		cg := NewComponentGate("order", []string{"matching"}, "", 1, 1)
		bc := cg.BufferedComponent()

		pass := bc.Passes(&testEntity{
			id:    "e-1",
			kind:  "order",
			stage: "matching",
		})
		if !pass {
			t.Error("expected that kind would pass")
		}

		pass = bc.Passes(&testEntity{
			id:    "e-1",
			kind:  "order",
			stage: "non-matching",
		})
		if pass {
			t.Error("expected that kind would fail")
		}

		pass = bc.Passes(&testEntity{
			id:    "e-1",
			kind:  "article",
			stage: "non-matching",
		})
		if !pass {
			t.Error("expected that kind would pass")
		}
	})

	t.Run("should count up inputs correctly", func(t *testing.T) {
		cg := NewComponentGate("order", []string{"correct"}, "", 1, 1)
		bc := cg.BufferedComponent()
		bc.Input("order", "correct")
		if bc.matchN != 1 || bc.noMatchN != 0 {
			t.Error("failed to count up properly")
		}

		bc.Input("order", "wrong")
		if bc.matchN != 1 || bc.noMatchN != 1 {
			t.Error("failed to count up properly")
		}
	})

	t.Run("should evaluate min correctly", func(t *testing.T) {
		cg := NewComponentGate("order", []string{"correct"}, "", 1, 1)
		bc := cg.BufferedComponent()

		pass, issues := bc.Evaluate("", "")
		if pass || len(issues) < 1 || issues[0].Content != ComponentGateCountTooLow || issues[0].Context["component_kind"] != "order" {
			t.Error("expected to fail evaluate")
		}

		bc.Input("order", "correct")
		pass, issues = bc.Evaluate("", "")
		if !pass || len(issues) > 0 {
			t.Error("expected to pass evaluate")
		}
	})

	t.Run("should evaluate n correctly", func(t *testing.T) {
		cg := NewComponentGate("order", []string{"correct"}, "", 1, 1)
		bc := cg.BufferedComponent()

		bc.Input("fulfilment", "wrong") // skip not matching

		bc.Input("order", "wrong")
		pass, issues := bc.Evaluate("", "")
		if pass || len(issues) < 1 || issues[0].Content != ComponentGateCountNoMatch || issues[0].Context["component_kind"] != "order" {
			t.Error("expected to fail evaluate")
		}

		bc.Input("order", "correct")
		pass, issues = bc.Evaluate("", "")
		if !pass || len(issues) > 0 {
			t.Error("expected to pass evaluate")
		}
	})

	t.Run("should be able to tell preferred stage", func(t *testing.T) {
		cg := NewComponentGate("order", []string{"correct"}, "", 1, 1)
		bc := cg.BufferedComponent()
		if bc.PreferredStage() != "" {
			t.Error("expected empty preferred stage")
		}

		cg = NewComponentGate("order", []string{"correct"}, "preferred", 1, 1)
		bc = cg.BufferedComponent()
		if bc.PreferredStage() != "preferred" {
			t.Error("expected filled preferred stage")
		}
	})

	t.Run("should be able to reference logic gate", func(t *testing.T) {
		cg := NewComponentGate("order", []string{"correct"}, "", 1, 1)
		bc := cg.BufferedComponent()
		if bc.LogicGate() != "" {
			t.Error("expected empty logic gate")
		}

		cg = NewComponentGate("order", []string{"correct"}, "", 1, 1, "g1")
		bc = cg.BufferedComponent()
		if bc.LogicGate() != "g1" {
			t.Error("expected filled logic gate")
		}
	})
}
