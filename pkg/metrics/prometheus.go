package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type PromotheusMetrics struct {
	requestsTotal      *prometheus.CounterVec
	requestDuration    *prometheus.HistogramVec
	userSignupsTotal   prometheus.Counter
	userSigninsTotal   prometheus.Counter
	deletedUsersTotal  prometheus.Counter
	createdListsTotal  prometheus.Counter
	deletedListsTotal  prometheus.Counter
	searchedListsTotal prometheus.Counter
	createdItemsTotal  prometheus.Counter
	deletedItemsTotal  prometheus.Counter
	searchedItemsTotal prometheus.Counter
}

// New initializes and returns a PromotheusMetrics instance with counters and histograms for tracking metrics
func New() Interface {
	return &PromotheusMetrics{
		requestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"endpoint", "status"},
		),
		requestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"endpoint"},
		),
		userSignupsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "user_signup_success_total",
			Help: "Total number of successful user sign-ups",
		}),
		userSigninsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "user_signin_success_total",
			Help: "Total number of successful user sign-ins",
		}),
		deletedUsersTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "deleted_users_total",
			Help: "Total number of successful deleted users",
		}),
		createdListsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "created_lists_total",
			Help: "Total number of successful created lists",
		}),
		deletedListsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "deleted_lists_total",
			Help: "Total number of successful deleted lists",
		}),
		searchedListsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "searched_lists_total",
			Help: "Total number of successful searched lists",
		}),
		createdItemsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "created_items_total",
			Help: "Total number of successful created items",
		}),
		deletedItemsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "deleted_items_total",
			Help: "Total number of successful deleted items",
		}),
		searchedItemsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "searched_items_total",
			Help: "Total number of successful searched items",
		}),
	}
}

// TrackRequestCount increments the counter for the number of HTTP requests for a given endpoint and status
func (m *PromotheusMetrics) TrackRequestCount(endpoint string, status string) {
	m.requestsTotal.WithLabelValues(endpoint, status).Inc()
}

// TrackRequestDuration observes the duration of HTTP requests for a given endpoint
func (m *PromotheusMetrics) TrackRequestDuration(endpoint string, duration float64) {
	m.requestDuration.WithLabelValues(endpoint).Observe(duration)
}

// IncrementSignedUpUsers increments the counter for successful user sign-ups
func (m *PromotheusMetrics) IncrementSignedUpUsers() {
	m.userSignupsTotal.Inc()
}

// IncrementSignedInUsers increments the counter for successful user sign-ins
func (m *PromotheusMetrics) IncrementSignedInUsers() {
	m.userSigninsTotal.Inc()
}

// IncrementDeletedUsers increments the counter for successfully deleted users
func (m *PromotheusMetrics) IncrementDeletedUsers() {
	m.deletedUsersTotal.Inc()
}

// IncrementCreatedLists increments the counter for successfully created lists
func (m *PromotheusMetrics) IncrementCreatedLists() {
	m.createdListsTotal.Inc()
}

// IncrementDeletedLists increments the counter for successfully deleted lists
func (m *PromotheusMetrics) IncrementDeletedLists() {
	m.deletedListsTotal.Inc()
}

// IncrementSearchedLists increments the counter for successfully searched lists
func (m *PromotheusMetrics) IncrementSearchedLists() {
	m.searchedListsTotal.Inc()
}

// IncrementCreatedItems increments the counter for successfully created items
func (m *PromotheusMetrics) IncrementCreatedItems() {
	m.createdItemsTotal.Inc()
}

// IncrementDeletedItems increments the counter for successfully deleted items
func (m *PromotheusMetrics) IncrementDeletedItems() {
	m.deletedItemsTotal.Inc()
}

// IncrementSearchedItems increments the counter for successfully searched items
func (m *PromotheusMetrics) IncrementSearchedItems() {
	m.searchedItemsTotal.Inc()
}
