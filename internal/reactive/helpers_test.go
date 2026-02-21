package reactive

// resetScheduler clears the default scheduler between tests.
func resetScheduler() {
	DefaultScheduler = Scheduler{}
	TrackingStack = TrackingStack[:0]
}
