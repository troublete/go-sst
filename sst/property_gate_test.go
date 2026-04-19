package sst

import (
	"fmt"
	"reflect"
	"testing"
)

func TestPropertyGates(t *testing.T) {
	t.Run("input test", func(t *testing.T) {

		for _, tc := range []struct {
			pgd      *PropertyGateDefinition
			in       map[string]string
			out      bool
			messages []string
		}{
			{
				pgd: NewPropertyGate(map[string]string{
					"a": "b",
				}),
				in: map[string]string{
					"a": "b",
				},
				out: true,
			},
			{
				pgd: NewPropertyGate(map[string]string{
					"a": "b",
				}),
				in: map[string]string{
					"a": "c",
				},
				out: false,
				messages: []string{
					RequiredPropertyNoMatch + "," + "key=a" + ",",
				},
			},
			{
				pgd: NewPropertyGate(map[string]string{
					"a": "b",
					"c": "d",
				}),
				in: map[string]string{
					"a": "b",
				},
				out: false,
				messages: []string{
					RequiredPropertyMissing + "," + "key=c" + ",",
				},
			},
			{
				pgd: NewPropertyGate(map[string]string{
					"a": "b",
					"c": "d",
				}),
				in: map[string]string{
					"a": "b",
					"c": "d",
					"e": "f",
				},
				out: true,
			},
			{
				pgd: NewPropertyGate(map[string]string{
					"a": "",
				}),
				in: map[string]string{
					"a": "any-value",
				},
				out: true,
			},
		} {
			t.Run(fmt.Sprintf("%#v", tc), func(t *testing.T) {
				got, messages := tc.pgd.Evaluate(tc.in, "")
				if got != tc.out || !reflect.DeepEqual(messages, tc.messages) {
					t.Errorf("failed, expected %v, got %v", got, tc.out)
				}
			})
		}
	})

	t.Run("should be able to receive logic gate", func(t *testing.T) {
		pg := NewPropertyGate(map[string]string{}, "g1")
		if pg.logicGate != "g1" {
			t.Error("expected logic gate to be set")
		}
	})
}
