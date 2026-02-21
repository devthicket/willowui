package reactive

// flushable is implemented by computeds and watches that need re-evaluation
// during a scheduler flush.
type flushable interface {
	// markDirty marks this node as needing re-evaluation.
	markDirty()
	// flush re-evaluates this node if it is dirty. The generation counter
	// is used to avoid duplicate evaluations in diamond dependency graphs.
	flush(gen uint64)
	// priority returns the topological priority. Watches run after computeds
	// so that watches see up-to-date computed values.
	// 0 = Computed, 1 = Watch.
	priority() int
}

// Scheduler batches reactive updates and flushes them once per frame.
type Scheduler struct {
	queue []flushable
	gen   uint64
}

// DefaultScheduler is the package-level scheduler instance used by all
// reactive primitives. Call DefaultScheduler.Flush() once per frame
// before the UI update pass.
var DefaultScheduler Scheduler

// Enqueue adds a flushable to the dirty set for the next flush.
// Duplicates are allowed; flush uses a generation counter to skip
// nodes that have already been processed.
func (s *Scheduler) Enqueue(f flushable) {
	s.queue = append(s.queue, f)
}

// Flush processes all enqueued dirty nodes. Computeds are flushed before
// watches so that watch callbacks observe up-to-date values.
// This should be called once per frame before the UI update pass.
func (s *Scheduler) Flush() {
	if len(s.queue) == 0 {
		return
	}

	s.gen++
	gen := s.gen

	// Partition: computeds (priority 0) before watches (priority 1).
	// We do a stable two-pass approach to avoid sort allocations.
	pending := s.queue
	s.queue = s.queue[:0] // reset for next frame, reuse backing array

	// First pass: flush all computeds.
	for _, f := range pending {
		if f.priority() == 0 {
			f.flush(gen)
		}
	}
	// Second pass: flush all watches.
	for _, f := range pending {
		if f.priority() == 1 {
			f.flush(gen)
		}
	}
}
