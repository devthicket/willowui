package integration

import (
	"strings"
	"testing"
	"time"

	ui "github.com/devthicket/willowui"
)

// ---------------------------------------------------------------------------
// Basic value API
// ---------------------------------------------------------------------------

func TestSearchBoxValue(t *testing.T) {
	resetScheduler()
	sb := ui.NewSearchBox("sb", newTestFont(), 14)
	defer sb.Dispose()

	sb.SetValue("hello")
	if sb.Value() != "hello" {
		t.Errorf("Value() = %q, want %q", sb.Value(), "hello")
	}
}

func TestSearchBoxPlaceholder(t *testing.T) {
	resetScheduler()
	sb := ui.NewSearchBox("sb", newTestFont(), 14)
	defer sb.Dispose()

	sb.SetPlaceholder("Search...")
	if sb.GetPlaceholder() != "Search..." {
		t.Errorf("GetPlaceholder() = %q, want %q", sb.GetPlaceholder(), "Search...")
	}
}

// ---------------------------------------------------------------------------
// Clear button visibility
// ---------------------------------------------------------------------------

func TestSearchBoxClearButtonVisibleWhenHasText(t *testing.T) {
	resetScheduler()
	sb := ui.NewSearchBox("sb", newTestFont(), 14)
	defer sb.Dispose()

	if sb.ClearVisible() {
		t.Error("clear button should be hidden when value is empty")
	}
	sb.SetValue("query")
	if !sb.ClearVisible() {
		t.Error("clear button should be visible when value is non-empty")
	}
}

func TestSearchBoxClearButtonHiddenWhenDisabled(t *testing.T) {
	resetScheduler()
	sb := ui.NewSearchBox("sb", newTestFont(), 14)
	defer sb.Dispose()

	sb.SetShowClearButton(false)
	sb.SetValue("query")
	if sb.ClearVisible() {
		t.Error("clear button should be hidden when showClearButton=false")
	}
}

// ---------------------------------------------------------------------------
// Clear functionality
// ---------------------------------------------------------------------------

func TestSearchBoxClearEmptiesValue(t *testing.T) {
	resetScheduler()
	sb := ui.NewSearchBox("sb", newTestFont(), 14)
	defer sb.Dispose()

	sb.SetValue("query")
	sb.Clear()
	if sb.Value() != "" {
		t.Errorf("after Clear(), Value() = %q, want empty", sb.Value())
	}
	if sb.ClearVisible() {
		t.Error("clear button should be hidden after Clear()")
	}
}

func TestSearchBoxClearFiresCallback(t *testing.T) {
	resetScheduler()
	sb := ui.NewSearchBox("sb", newTestFont(), 14)
	defer sb.Dispose()

	cleared := false
	sb.SetOnClear(func() { cleared = true })
	sb.SetValue("query")
	sb.Clear()
	if !cleared {
		t.Error("OnClear callback should fire after Clear()")
	}
}

func TestSearchBoxClearClearsResults(t *testing.T) {
	resetScheduler()
	sb := ui.NewSearchBox("sb", newTestFont(), 14)
	defer sb.Dispose()

	results := ui.NewArray[string]()
	ui.SetSearchBoxFunc(sb, results, func(q string) []string {
		return []string{"a", "b", "c"}
	})

	sb.SetValue("abc")
	sb.TriggerSearchNow()
	if results.Len() == 0 {
		t.Error("results should be populated after TriggerSearchNow()")
	}

	sb.Clear()
	if results.Len() != 0 {
		t.Errorf("results.Len() = %d after Clear(), want 0", results.Len())
	}
}

// ---------------------------------------------------------------------------
// Debounce
// ---------------------------------------------------------------------------

func TestSearchBoxDebounceDelaysSearch(t *testing.T) {
	resetScheduler()
	sb := ui.NewSearchBox("sb", newTestFont(), 14)
	defer sb.Dispose()

	sb.SetDebounce(200 * time.Millisecond)
	if sb.Debounce() != 200*time.Millisecond {
		t.Errorf("Debounce() = %v, want 200ms", sb.Debounce())
	}
}

func TestSearchBoxTriggerSearchNow(t *testing.T) {
	resetScheduler()
	sb := ui.NewSearchBox("sb", newTestFont(), 14)
	defer sb.Dispose()

	sb.SetDebounce(10 * time.Second) // very long debounce

	searchCalled := false
	results := ui.NewArray[string]()
	ui.SetSearchBoxFunc(sb, results, func(q string) []string {
		searchCalled = true
		return []string{"hit"}
	})

	sb.SetValue("abc")
	// Without calling TriggerSearchNow, search would not run yet.
	if searchCalled {
		t.Error("search should not run before debounce or TriggerSearchNow()")
	}

	sb.TriggerSearchNow()
	if !searchCalled {
		t.Error("TriggerSearchNow() should bypass debounce and run search immediately")
	}
}

// ---------------------------------------------------------------------------
// MinQueryLength
// ---------------------------------------------------------------------------

func TestSearchBoxMinQueryLength(t *testing.T) {
	resetScheduler()
	sb := ui.NewSearchBox("sb", newTestFont(), 14)
	defer sb.Dispose()

	sb.SetMinQueryLength(3)
	if sb.MinQueryLength() != 3 {
		t.Errorf("MinQueryLength() = %d, want 3", sb.MinQueryLength())
	}

	results := ui.NewArray[string]()
	ui.SetSearchBoxFunc(sb, results, func(q string) []string {
		return []string{"result"}
	})

	// Query shorter than min — results should be cleared.
	sb.SetValue("ab")
	sb.TriggerSearchNow()
	if results.Len() != 0 {
		t.Errorf("results.Len() = %d for short query, want 0", results.Len())
	}

	// Query at min — results should be populated.
	sb.SetValue("abc")
	sb.TriggerSearchNow()
	if results.Len() == 0 {
		t.Error("results should be populated when query >= MinQueryLength")
	}
}

// ---------------------------------------------------------------------------
// SetSearchFunc
// ---------------------------------------------------------------------------

func TestSearchBoxSetSearchFunc(t *testing.T) {
	resetScheduler()
	sb := ui.NewSearchBox("sb", newTestFont(), 14)
	defer sb.Dispose()

	items := []string{"Potion", "Elixir", "Ether", "Phoenix Down"}
	results := ui.NewArray[string]()
	ui.SetSearchBoxFunc(sb, results, func(q string) []string {
		q = strings.ToLower(q)
		var out []string
		for _, item := range items {
			if strings.Contains(strings.ToLower(item), q) {
				out = append(out, item)
			}
		}
		return out
	})

	sb.SetValue("th")
	sb.TriggerSearchNow()

	if results.Len() == 0 {
		t.Error("results should contain matches after TriggerSearchNow()")
	}
	// "Ether" and "Phoenix Down" (phoe-nix) don't contain "th"? Let's check:
	// "Potion" -> no, "Elixir" -> no, "Ether" -> yes (eth contains th), "Phoenix Down" -> no
	if results.Len() != 1 {
		t.Errorf("results.Len() = %d for query 'th', want 1 (Ether)", results.Len())
	}
}

// ---------------------------------------------------------------------------
// SetSearchIntoFunc
// ---------------------------------------------------------------------------

func TestSearchBoxSetSearchIntoFunc(t *testing.T) {
	resetScheduler()
	sb := ui.NewSearchBox("sb", newTestFont(), 14)
	defer sb.Dispose()

	results := ui.NewArray[int]()
	ui.SetSearchBoxIntoFunc(sb, results, func(q string, arr *ui.Array[int]) {
		arr.Clear()
		for i := 0; i < 5; i++ {
			arr.Push(i)
		}
	})

	sb.SetValue("anything")
	sb.TriggerSearchNow()

	if results.Len() != 5 {
		t.Errorf("results.Len() = %d, want 5", results.Len())
	}
}

// ---------------------------------------------------------------------------
// Search lifecycle callbacks
// ---------------------------------------------------------------------------

func TestSearchBoxLifecycleCallbacks(t *testing.T) {
	resetScheduler()
	sb := ui.NewSearchBox("sb", newTestFont(), 14)
	defer sb.Dispose()

	var startQuery, finishQuery, emptyQuery string
	var finishCount int
	var emptyCalled bool

	sb.SetOnSearchStart(func(q string) { startQuery = q })
	sb.SetOnSearchFinish(func(q string, count int) {
		finishQuery = q
		finishCount = count
	})
	sb.SetOnSearchEmpty(func(q string) {
		emptyQuery = q
		emptyCalled = true
	})

	results := ui.NewArray[string]()
	ui.SetSearchBoxFunc(sb, results, func(q string) []string {
		if q == "nomatch" {
			return nil
		}
		return []string{"result"}
	})

	// Normal search.
	sb.SetValue("test")
	sb.TriggerSearchNow()
	if startQuery != "test" {
		t.Errorf("OnSearchStart got %q, want %q", startQuery, "test")
	}
	if finishQuery != "test" || finishCount != 1 {
		t.Errorf("OnSearchFinish: query=%q count=%d, want test/1", finishQuery, finishCount)
	}
	if emptyCalled {
		t.Error("OnSearchEmpty should not fire when results are non-empty")
	}

	// Empty result search.
	sb.SetValue("nomatch")
	sb.TriggerSearchNow()
	if !emptyCalled {
		t.Error("OnSearchEmpty should fire when search returns zero results")
	}
	if emptyQuery != "nomatch" {
		t.Errorf("OnSearchEmpty got %q, want %q", emptyQuery, "nomatch")
	}
}

// ---------------------------------------------------------------------------
// ResultsCount / IsSearching
// ---------------------------------------------------------------------------

func TestSearchBoxResultsCount(t *testing.T) {
	resetScheduler()
	sb := ui.NewSearchBox("sb", newTestFont(), 14)
	defer sb.Dispose()

	results := ui.NewArray[string]()
	ui.SetSearchBoxFunc(sb, results, func(q string) []string {
		return []string{"a", "b", "c"}
	})

	sb.SetValue("x")
	sb.TriggerSearchNow()

	if sb.ResultsCount() != 3 {
		t.Errorf("ResultsCount() = %d, want 3", sb.ResultsCount())
	}
}

// ---------------------------------------------------------------------------
// BindValue
// ---------------------------------------------------------------------------

func TestSearchBoxBindValue(t *testing.T) {
	resetScheduler()
	sb := ui.NewSearchBox("sb", newTestFont(), 14)
	defer sb.Dispose()

	ref := ui.NewRef("initial")
	sb.BindValue(ref)
	if sb.Value() != "initial" {
		t.Errorf("after BindValue, Value() = %q, want %q", sb.Value(), "initial")
	}

	ref.Set("updated")
	ui.DefaultScheduler.Flush()
	if sb.Value() != "updated" {
		t.Errorf("after ref.Set, Value() = %q, want %q", sb.Value(), "updated")
	}
}

// ---------------------------------------------------------------------------
// InsertText / DeleteBack
// ---------------------------------------------------------------------------

func TestSearchBoxInsertText(t *testing.T) {
	resetScheduler()
	sb := ui.NewSearchBox("sb", newTestFont(), 14)
	defer sb.Dispose()

	sb.InsertText("hello")
	if sb.Value() != "hello" {
		t.Errorf("after InsertText, Value() = %q, want %q", sb.Value(), "hello")
	}
	if sb.GetCursorPos() != 5 {
		t.Errorf("cursor pos = %d, want 5", sb.GetCursorPos())
	}
}

func TestSearchBoxDeleteBack(t *testing.T) {
	resetScheduler()
	sb := ui.NewSearchBox("sb", newTestFont(), 14)
	defer sb.Dispose()

	sb.SetValue("hello")
	sb.DeleteBack()
	if sb.Value() != "hell" {
		t.Errorf("after DeleteBack, Value() = %q, want %q", sb.Value(), "hell")
	}
}

// ---------------------------------------------------------------------------
// onChange / onSubmit callbacks
// ---------------------------------------------------------------------------

func TestSearchBoxOnChange(t *testing.T) {
	resetScheduler()
	sb := ui.NewSearchBox("sb", newTestFont(), 14)
	defer sb.Dispose()

	var changedTo string
	sb.SetOnChange(func(q string) { changedTo = q })

	sb.InsertText("hi")
	if changedTo != "hi" {
		t.Errorf("OnChange got %q, want %q", changedTo, "hi")
	}
}

func TestSearchBoxOnSubmit(t *testing.T) {
	resetScheduler()
	sb := ui.NewSearchBox("sb", newTestFont(), 14)
	defer sb.Dispose()

	var submitted string
	sb.SetOnSubmit(func(q string) { submitted = q })
	sb.SetValue("query")
	sb.Submit()
	if submitted != "query" {
		t.Errorf("OnSubmit got %q, want %q", submitted, "query")
	}
}

func TestSearchBoxSearchOnSubmit(t *testing.T) {
	resetScheduler()
	sb := ui.NewSearchBox("sb", newTestFont(), 14)
	defer sb.Dispose()

	sb.SetSearchOnSubmit(true)
	sb.SetSearchOnChange(false)
	sb.SetDebounce(0)

	searched := false
	results := ui.NewArray[string]()
	ui.SetSearchBoxFunc(sb, results, func(q string) []string {
		searched = true
		return []string{"hit"}
	})

	sb.SetValue("test")
	if searched {
		t.Error("search should not run on change when searchOnChange=false")
	}

	sb.Submit()
	if !searched {
		t.Error("search should run on Submit() when searchOnSubmit=true")
	}
}
