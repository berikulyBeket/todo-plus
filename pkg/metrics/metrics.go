package metrics

// Interface defines methods for tracking various metrics in the application
type Interface interface {
	TrackRequestCount(endpoint string, status string)
	TrackRequestDuration(endpoint string, duration float64)
	IncrementSignedUpUsers()
	IncrementSignedInUsers()
	IncrementDeletedUsers()
	IncrementCreatedLists()
	IncrementSearchedLists()
	IncrementDeletedLists()
	IncrementCreatedItems()
	IncrementDeletedItems()
	IncrementSearchedItems()
}
