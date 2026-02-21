package reactive

import "testing"

// ---------------------------------------------------------------------------
// ArraySortFold
// ---------------------------------------------------------------------------

func TestArraySortFold(t *testing.T) {
	resetScheduler()

	type item struct {
		name string
	}
	a := NewArrayFrom([]item{
		{name: "cherry"},
		{name: "Apple"},
		{name: "banana"},
		{name: "APRICOT"},
	})

	ArraySortFold(a, func(v item) string { return v.name })

	want := []string{"Apple", "APRICOT", "banana", "cherry"}
	for i, w := range want {
		if a.At(i).name != w {
			t.Fatalf("at %d: expected %q, got %q", i, w, a.At(i).name)
		}
	}
}

func TestArraySortFoldEmpty(t *testing.T) {
	resetScheduler()
	a := NewArray[string]()
	ArraySortFold(a, func(v string) string { return v })
	if a.Len() != 0 {
		t.Fatalf("expected empty, got len %d", a.Len())
	}
}

func TestArraySortFoldStrings(t *testing.T) {
	resetScheduler()
	a := NewArrayFrom([]string{"Zebra", "apple", "Mango", "BANANA"})
	ArraySortFold(a, func(v string) string { return v })
	want := []string{"apple", "BANANA", "Mango", "Zebra"}
	for i, w := range want {
		if a.At(i) != w {
			t.Fatalf("at %d: expected %q, got %q", i, w, a.At(i))
		}
	}
}
