package metrics

// NoOpMetrics is a metrics interface that does nothing, useful for testing
type NoOpMetrics struct{}

func (n *NoOpMetrics) TrackRequestCount(endpoint string, status string)       {}
func (n *NoOpMetrics) TrackRequestDuration(endpoint string, duration float64) {}
func (n *NoOpMetrics) IncrementSignedUpUsers()                                {}
func (n *NoOpMetrics) IncrementSignedInUsers()                                {}
func (n *NoOpMetrics) IncrementDeletedUsers()                                 {}
func (n *NoOpMetrics) IncrementCreatedLists()                                 {}
func (n *NoOpMetrics) IncrementDeletedLists()                                 {}
func (n *NoOpMetrics) IncrementSearchedLists()                                {}
func (n *NoOpMetrics) IncrementCreatedItems()                                 {}
func (n *NoOpMetrics) IncrementDeletedItems()                                 {}
func (n *NoOpMetrics) IncrementSearchedItems()                                {}
