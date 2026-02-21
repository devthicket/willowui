package reactive

// ---------------------------------------------------------------------------
// Record
// ---------------------------------------------------------------------------

// onRecordChangedCB wraps a key+value change callback for safe O(n) removal.
type onRecordChangedCB struct{ fn func(key string, value any) }

// Record is a reactive key-value object. Each field is stored as a *Ref[any]
// so that WatchEffect and Computed callers tracking a specific field are not
// disturbed by unrelated field changes.
//
// Record is not goroutine-safe — all mutations must occur on the game-loop goroutine.
type Record struct {
	fields   map[string]*Ref[any]
	onChange []*onRecordChangedCB
}

// NewRecord creates an empty Record.
func NewRecord() *Record {
	return &Record{fields: make(map[string]*Ref[any])}
}

// NewRecordFrom creates a Record pre-populated with fields. The map is copied —
// the caller retains ownership of the original.
func NewRecordFrom(fields map[string]any) *Record {
	r := NewRecord()
	for k, v := range fields {
		r.fields[k] = NewRef[any](v)
	}
	return r
}

// ---------------------------------------------------------------------------
// Field access
// ---------------------------------------------------------------------------

// Set assigns value to key, creating the field if it doesn't exist.
// Fires onChange callbacks only when the value actually changes.
func (r *Record) Set(key string, value any) {
	ref, ok := r.fields[key]
	if !ok {
		r.fields[key] = NewRef[any](value)
		for _, c := range r.onChange {
			c.fn(key, value)
		}
		return
	}
	old := ref.Peek()
	ref.Set(value) // no-op internally if equal; propagates to reactive graph
	if old != value {
		for _, c := range r.onChange {
			c.fn(key, value)
		}
	}
}

// Get returns the current value for key, or nil if the key is not set.
// Registers a reactive dependency when called inside a WatchEffect or Computed.
func (r *Record) Get(key string) any {
	ref, ok := r.fields[key]
	if !ok {
		return nil
	}
	return ref.Get()
}

// Has reports whether key exists without creating it.
func (r *Record) Has(key string) bool {
	_, ok := r.fields[key]
	return ok
}

// Delete removes key from the record and fires OnChange with a nil value.
func (r *Record) Delete(key string) {
	if _, ok := r.fields[key]; !ok {
		return
	}
	delete(r.fields, key)
	for _, c := range r.onChange {
		c.fn(key, nil)
	}
}

// Keys returns a snapshot of all current keys (order is not guaranteed).
func (r *Record) Keys() []string {
	keys := make([]string, 0, len(r.fields))
	for k := range r.fields {
		keys = append(keys, k)
	}
	return keys
}

// SetMany assigns multiple fields. OnChange fires once per changed field.
func (r *Record) SetMany(fields map[string]any) {
	for k, v := range fields {
		r.Set(k, v)
	}
}

// ToMap returns a snapshot copy of all fields as a plain map.
// Mutations to the returned map do not affect the Record.
func (r *Record) ToMap() map[string]any {
	out := make(map[string]any, len(r.fields))
	for k, ref := range r.fields {
		out[k] = ref.Peek()
	}
	return out
}

// ---------------------------------------------------------------------------
// Reactive access
// ---------------------------------------------------------------------------

// Ref returns the internal *Ref[any] for key, creating the field with a nil
// value if it does not yet exist. Use inside WatchEffect or Computed to track
// a specific field reactively.
func (r *Record) Ref(key string) *Ref[any] {
	ref, ok := r.fields[key]
	if !ok {
		ref = NewRef[any](nil)
		r.fields[key] = ref
	}
	return ref
}

// ---------------------------------------------------------------------------
// Change callbacks
// ---------------------------------------------------------------------------

// OnFieldChange registers a callback that fires when the named field changes.
// The callback does NOT fire on registration — only on subsequent changes.
// Returns a WatchHandle — call Stop() to unsubscribe.
func (r *Record) OnFieldChange(key string, fn func(value any)) WatchHandle {
	ref := r.Ref(key)
	w := &watchNode{}
	first := true
	w.fn = func() {
		v := ref.Get() // establishes reactive dependency
		if first {
			first = false
			return
		}
		fn(v)
	}
	w.run() // initial run: captures dep, skips callback via first-flag
	return WatchHandle{node: w}
}

// OnChange registers a callback that fires when any field changes, passing the
// key and new value. Returns a WatchHandle — call Stop() to unsubscribe.
func (r *Record) OnChange(fn func(key string, value any)) WatchHandle {
	cb := &onRecordChangedCB{fn: fn}
	r.onChange = append(r.onChange, cb)
	return WatchHandle{stopFn: func() {
		for i, c := range r.onChange {
			if c == cb {
				last := len(r.onChange) - 1
				r.onChange[i] = r.onChange[last]
				r.onChange[last] = nil
				r.onChange = r.onChange[:last]
				return
			}
		}
	}}
}
