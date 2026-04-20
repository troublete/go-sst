package sst

import (
	"fmt"
	"slices"
	"testing"
)

/*
	the following tests are just evaluations to determine the baseline for certain operation decision in the library
	e.g. should a lookup for a small n (as n is usually small) be implemented as a map or should it be done on the slice
	directly; or should an iteration happen on a map or on the slice (which obviously is cheaper); or should the response
	be a concat string or proper struct

	this therefore provides just development notes that might be interesting to look at
*/

func BenchmarkLookup(b *testing.B) {
	b.Run("10", func(b *testing.B) {
		b.Run("map lookup", func(b *testing.B) {
			m := map[string]string{
				"a": "av", "b": "bv", "c": "cv", "d": "dv", "e": "ev",
				"f": "fv", "g": "gv", "h": "hv", "i": "iv", "j": "iv",
			}

			for b.Loop() {
				_ = m["j"]
			}
		})

		b.Run("double array lookup", func(b *testing.B) {
			keys := []string{
				"a", "b", "c", "d", "e",
				"f", "g", "h", "i", "j",
			}
			values := []string{
				"av", "av", "av", "av", "av",
				"av", "av", "av", "av", "av",
			}

			for b.Loop() {
				for idx, k := range keys {
					if k == "j" {
						_ = values[idx]
					}
				}
			}
		})

		b.Run("double array lookup binary search", func(b *testing.B) {
			keys := []string{
				"a", "b", "c", "d", "e",
				"f", "g", "h", "i", "j",
			}
			values := []string{
				"av", "av", "av", "av", "av",
				"av", "av", "av", "av", "av",
			}

			for b.Loop() {
				if idx, ok := slices.BinarySearch(keys, "j"); ok {
					_ = values[idx]
				}
			}
		})
	})
	b.Run("20", func(b *testing.B) {
		b.Run("map lookup", func(b *testing.B) {
			m := map[string]string{
				"a": "av", "b": "bv", "c": "cv", "d": "dv", "e": "ev",
				"f": "fv", "g": "gv", "h": "hv", "i": "iv", "j": "iv",
				"k": "iv", "l": "iv", "m": "iv", "n": "iv", "o": "iv",
				"p": "iv", "q": "iv", "r": "iv", "s": "iv", "t": "iv",
			}

			for b.Loop() {
				_ = m["t"]
			}
		})

		b.Run("double array lookup", func(b *testing.B) {
			keys := []string{
				"a", "b", "c", "d", "e",
				"f", "g", "h", "i", "j",
				"k", "l", "m", "n", "o",
				"p", "q", "r", "s", "t",
			}
			values := []string{
				"av", "av", "av", "av", "av",
				"av", "av", "av", "av", "av",
				"av", "av", "av", "av", "av",
				"av", "av", "av", "av", "av",
			}

			for b.Loop() {
				for idx, k := range keys {
					if k == "t" {
						_ = values[idx]
					}
				}
			}
		})

		b.Run("double array lookup binary search", func(b *testing.B) {
			keys := []string{
				"a", "b", "c", "d", "e",
				"f", "g", "h", "i", "j",
				"k", "l", "m", "n", "o",
				"p", "q", "r", "s", "t",
			}
			values := []string{
				"av", "av", "av", "av", "av",
				"av", "av", "av", "av", "av",
				"av", "av", "av", "av", "av",
				"av", "av", "av", "av", "av",
			}

			for b.Loop() {
				if idx, ok := slices.BinarySearch(keys, "t"); ok {
					_ = values[idx]
				}
			}
		})
	})
}

func BenchmarkIterate(b *testing.B) {
	for _, i := range []int{10, 20, 50, 100} {
		var sl []string
		m := map[int]string{}
		for n := 0; n < i; n++ {
			sl = append(sl, "strstrstr")
			m[n] = "strstrstr"
		}

		b.Run(fmt.Sprintf("slice-%d", i), func(b *testing.B) {
			for b.Loop() {
				for _, str := range sl {
					_ = str
				}
			}
		})

		b.Run(fmt.Sprintf("map-%d", i), func(b *testing.B) {
			for b.Loop() {
				for _, str := range m {
					_ = str
				}
			}
		})
	}
}

func BenchmarkResponseAllocation(b *testing.B) {
	b.Run("string concat", func(b *testing.B) {
		_ = "aaaaaaaa" + "bbbbbbbb" + "cccccccc" + "dddddddd"
	})

	type s struct {
		a, b, c, d string
	}
	b.Run("struct create", func(b *testing.B) {
		st := &s{}
		st.a = "aaaaaaaa"
		st.b = "bbbbbbbb"
		st.c = "cccccccc"
		st.d = "dddddddd"
	})
}
