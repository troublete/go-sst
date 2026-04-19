package sst

import (
	"fmt"
	"testing"
)

func TestLogicGate(t *testing.T) {
	t.Run("input test", func(t *testing.T) {

		for _, tc := range []struct {
			lgd LogicGateDefinition
			ins []bool
			out bool
		}{
			{
				lgd: LogicGateDefinition{},
				ins: []bool{true, true, true, true},
				out: true,
			},
			{
				lgd: LogicGateDefinition{},
				ins: []bool{true, true, true, false},
				out: false,
			},
			{
				lgd: LogicGateDefinition{},
				ins: []bool{true, false, true, false},
				out: false,
			},
			{
				lgd: LogicGateDefinition{
					isOr: true,
				},
				ins: []bool{false, false, false},
				out: false,
			},
			{
				lgd: LogicGateDefinition{
					isOr: true,
				},
				ins: []bool{false, true, false},
				out: true,
			},
			{
				lgd: LogicGateDefinition{
					isOr: true,
				},
				ins: []bool{true, true, true},
				out: true,
			},
			{
				lgd: LogicGateDefinition{
					isOr: true,
				},
				ins: []bool{},
				out: false,
			},
			{
				lgd: LogicGateDefinition{},
				ins: []bool{},
				out: false,
			},
		} {
			t.Run(fmt.Sprintf("or=%v %#v %#v", tc.lgd.isOr, tc.ins, tc.out), func(t *testing.T) {
				c := tc.lgd.Gate()
				for _, v := range tc.ins {
					c.Input(v)
				}

				if c.Evaluate() != tc.out {
					t.Errorf("failed check, want %v, got %v", tc.out, c.Evaluate())
				}
			})
		}

	})

	t.Run("should provide a simple AND creation", func(t *testing.T) {
		g := And("g1")
		if g.name != "g1" || g.isOr == true {
			t.Error("expected g1 AND gate")
		}
	})

	t.Run("should provide a simple OR creation", func(t *testing.T) {
		g := Or("g2")
		if g.name != "g2" || g.isOr == false {
			t.Error("expected g2 OR gate")
		}
	})

	t.Run("a gate component should be able to tell its name", func(t *testing.T) {
		gand := And("g1").Gate().Name()
		gor := Or("g2").Gate().Name()
		if gand != "g1" || gor != "g2" {
			t.Error("expected proper names")
		}
	})
}
