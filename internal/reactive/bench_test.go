package reactive

import "testing"

// ---------------------------------------------------------------------------
// Array benchmarks
// ---------------------------------------------------------------------------

// BenchmarkArrayPush — single-item fast path (no allocation expected).
func BenchmarkArrayPush(b *testing.B) {
	a := NewArrayWithCap[int](b.N + 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Push(i)
	}
}

// BenchmarkArrayPushWithCallback — Push with one OnChange subscriber.
func BenchmarkArrayPushWithCallback(b *testing.B) {
	a := NewArrayWithCap[int](b.N + 1)
	a.OnChange(func() {})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Push(i)
	}
}

// BenchmarkArrayPushWithWatchEffect — Push with one reactive WatchEffect subscriber.
func BenchmarkArrayPushWithWatchEffect(b *testing.B) {
	resetScheduler()
	a := NewArrayWithCap[int](b.N + 1)
	WatchEffect(func() { _ = a.Len() })
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Push(i)
		DefaultScheduler.Flush()
	}
}

// BenchmarkArrayPop — single-item fast path.
func BenchmarkArrayPop(b *testing.B) {
	a := NewArrayWithCap[int](b.N)
	for i := 0; i < b.N; i++ {
		a.items = append(a.items, i) // fill without callbacks
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Pop()
	}
}

// BenchmarkArrayInsertFront — worst case: O(n) shift.
func BenchmarkArrayInsertFront(b *testing.B) {
	a := NewArrayWithCap[int](b.N + 1)
	for i := 0; i < b.N; i++ {
		a.Insert(0, i)
	}
}

// BenchmarkArrayRemoveFront — worst case: O(n) shift.
func BenchmarkArrayRemoveFront(b *testing.B) {
	a := NewArrayWithCap[int](b.N)
	for i := 0; i < b.N; i++ {
		a.items = append(a.items, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Remove(0)
	}
}

// BenchmarkArrayMove — single move on a 1000-element array.
func BenchmarkArrayMove(b *testing.B) {
	const size = 1000
	a := NewArrayWithCap[int](size)
	for i := 0; i < size; i++ {
		a.items = append(a.items, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Move(0, size-1)
		a.Move(size-1, 0) // restore
	}
}

// BenchmarkArraySort — pdqsort on 1000 ints.
func BenchmarkArraySort(b *testing.B) {
	const size = 1000
	src := make([]int, size)
	for i := range src {
		src[i] = size - i
	}
	a := NewArrayFrom(src)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Sort(func(x, y int) int { return x - y })
	}
}

// BenchmarkArrayBatch — 100 pushes coalesced into one notification.
func BenchmarkArrayBatch(b *testing.B) {
	a := NewArray[int]()
	a.OnChange(func() {})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Clear()
		a.Batch(func() {
			for j := 0; j < 100; j++ {
				a.Push(j)
			}
		})
	}
}

// BenchmarkArrayBatchVsUnbatched — compare batched vs unbatched 100-push burst.
func BenchmarkArray100PushUnbatched(b *testing.B) {
	a := NewArray[int]()
	a.OnChange(func() {})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Clear()
		for j := 0; j < 100; j++ {
			a.Push(j)
		}
	}
}

// BenchmarkArrayLen — reactive read (tracks version).
func BenchmarkArrayLen(b *testing.B) {
	a := NewArrayFrom([]int{1, 2, 3, 4, 5})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = a.Len()
	}
}

// BenchmarkArrayAt — single element access.
func BenchmarkArrayAt(b *testing.B) {
	a := NewArrayFrom([]int{1, 2, 3, 4, 5})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = a.At(i % 5)
	}
}

// BenchmarkArrayCallbackStop — registering and stopping callbacks.
func BenchmarkArrayCallbackStop(b *testing.B) {
	a := NewArray[int]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h := a.OnChange(func() {})
		h.Stop()
	}
}

// BenchmarkArrayMap — transform 1000-element array.
func BenchmarkArrayMap(b *testing.B) {
	src := make([]int, 1000)
	for i := range src {
		src[i] = i
	}
	a := NewArrayFrom(src)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ArrayMap(a, func(v int) int { return v * 2 })
	}
}

// BenchmarkIndexOf — linear scan on 1000-element array (worst case: not found).
func BenchmarkIndexOf(b *testing.B) {
	src := make([]int, 1000)
	for i := range src {
		src[i] = i
	}
	a := NewArrayFrom(src)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IndexOf(a, -1) // not found → full scan
	}
}

// ---------------------------------------------------------------------------
// Record benchmarks
// ---------------------------------------------------------------------------

// BenchmarkRecordSet — set a field on an existing record.
func BenchmarkRecordSet(b *testing.B) {
	r := NewRecord()
	r.Set("score", 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Set("score", i)
	}
}

// BenchmarkRecordSetNew — set a brand-new field each time (creates Ref).
func BenchmarkRecordSetNew(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := NewRecord()
		r.Set("score", i)
	}
}

// BenchmarkRecordGet — tracked read inside WatchEffect.
func BenchmarkRecordGet(b *testing.B) {
	r := NewRecord()
	r.Set("score", 42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Get("score")
	}
}

// BenchmarkRecordSetWithCallback — Set with one OnChange subscriber.
func BenchmarkRecordSetWithCallback(b *testing.B) {
	r := NewRecord()
	r.Set("score", 0)
	r.OnChange(func(string, any) {})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Set("score", i)
	}
}

// BenchmarkRecordSetWithWatchEffect — Set driving a reactive WatchEffect.
func BenchmarkRecordSetWithWatchEffect(b *testing.B) {
	resetScheduler()
	r := NewRecord()
	r.Set("score", 0)
	WatchEffect(func() { _ = r.Get("score") })
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Set("score", i)
		DefaultScheduler.Flush()
	}
}

// BenchmarkRecordSetMany10 — SetMany with 10 fields.
func BenchmarkRecordSetMany10(b *testing.B) {
	r := NewRecord()
	fields := map[string]any{
		"f0": 0, "f1": 1, "f2": 2, "f3": 3, "f4": 4,
		"f5": 5, "f6": 6, "f7": 7, "f8": 8, "f9": 9,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.SetMany(fields)
	}
}

// BenchmarkRecordToMap — snapshot 10-field record.
func BenchmarkRecordToMap(b *testing.B) {
	r := NewRecordFrom(map[string]any{
		"f0": 0, "f1": 1, "f2": 2, "f3": 3, "f4": 4,
		"f5": 5, "f6": 6, "f7": 7, "f8": 8, "f9": 9,
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.ToMap()
	}
}

// BenchmarkRecordOnChangeStop — register + stop OnChange.
func BenchmarkRecordOnChangeStop(b *testing.B) {
	r := NewRecord()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h := r.OnChange(func(string, any) {})
		h.Stop()
	}
}

// BenchmarkNewRecord — construction cost.
func BenchmarkNewRecord(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewRecord()
	}
}

// BenchmarkNewRecordFrom10 — construction from a 10-field map.
func BenchmarkNewRecordFrom10(b *testing.B) {
	fields := map[string]any{
		"f0": 0, "f1": 1, "f2": 2, "f3": 3, "f4": 4,
		"f5": 5, "f6": 6, "f7": 7, "f8": 8, "f9": 9,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewRecordFrom(fields)
	}
}

// ---------------------------------------------------------------------------
// Theory comparison benchmarks
// ---------------------------------------------------------------------------

// Theory 1: Array construction — direct reactiveSource eliminates NewRef alloc.
func BenchmarkTheory1_NewArray(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = NewArray[int]()
	}
}

// Theory 1: Push with no subscribers — bumpVersion is now just a++ + empty range.
func BenchmarkTheory1_PushNoSubscribers(b *testing.B) {
	a := NewArrayWithCap[int](b.N + 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Push(i)
	}
}

// Theory 1: Push with reactive subscriber — bumpVersion now iterates directly.
func BenchmarkTheory1_PushWithWatchEffect(b *testing.B) {
	resetScheduler()
	a := NewArrayWithCap[int](b.N + 1)
	WatchEffect(func() { _ = a.Len() })
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Push(i)
		DefaultScheduler.Flush()
	}
}

// Theory 2: Lazy Version() — array that never calls Version() has zero versionRef alloc.
func BenchmarkTheory2_NewArrayNoVersionCall(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		a := NewArray[int]()
		a.Push(1) // typical usage: never calls Version()
		_ = a.Len()
	}
}

func BenchmarkTheory2_NewArrayWithVersionCall(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		a := NewArray[int]()
		_ = a.Version() // forces lazy versionRef alloc
		a.Push(1)
		_ = a.Len()
	}
}

// Theory 3: Multi-item Push with no OnAdded callbacks — skips event slice build.
func BenchmarkTheory3_PushMultiNoOnAdded(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a := NewArray[int]()
		a.Push(1, 2, 3, 4, 5) // multi-item, no OnAdded subscriber
	}
}

func BenchmarkTheory3_PushMultiWithOnAdded(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a := NewArray[int]()
		a.OnAdded(func(int, int) {})
		a.Push(1, 2, 3, 4, 5) // multi-item, has OnAdded — must build event slice
	}
}
