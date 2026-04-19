package sst

import (
	"slices"
	"testing"
)

func BenchmarkLookup20(b *testing.B) {
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
}

func BenchmarkLookup10(b *testing.B) {
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
}
