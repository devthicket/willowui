package reactive

import (
	"cmp"
	"slices"
	"strings"
)

// ArrayMap transforms each element of a using fn and returns a plain []U.
// The result is not reactive.
func ArrayMap[T, U any](a *Array[T], fn func(T) U) []U {
	out := make([]U, len(a.items))
	for i, v := range a.items {
		out[i] = fn(v)
	}
	return out
}

// ArrayReduce folds a into a single value of type U using fn, starting from init.
func ArrayReduce[T, U any](a *Array[T], fn func(U, T) U, init U) U {
	acc := init
	for _, v := range a.items {
		acc = fn(acc, v)
	}
	return acc
}

// ArraySort sorts a in ascending order. T must satisfy cmp.Ordered.
func ArraySort[T cmp.Ordered](a *Array[T]) {
	a.Sort(func(x, y T) int { return cmp.Compare(x, y) })
}

// ArraySortDesc sorts a in descending order. T must satisfy cmp.Ordered.
func ArraySortDesc[T cmp.Ordered](a *Array[T]) {
	a.Sort(func(x, y T) int { return cmp.Compare(y, x) })
}

// ArraySortFold sorts a case-insensitively using key to extract a string from
// each element. Comparison uses Unicode simple case-folding so "Abc" == "abc" == "ABC".
func ArraySortFold[T any](a *Array[T], key func(T) string) {
	a.Sort(func(x, y T) int { return strings.Compare(strings.ToLower(key(x)), strings.ToLower(key(y))) })
}

// IndexOf returns the index of item in a, or -1 if absent. T must be comparable.
func IndexOf[T comparable](a *Array[T], item T) int {
	return slices.Index(a.items, item)
}

// Includes reports whether item is present in a. T must be comparable.
func Includes[T comparable](a *Array[T], item T) bool {
	return slices.Contains(a.items, item)
}
