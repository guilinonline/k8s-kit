package cluster

import "time"

// HealthCheckConfig contains configuration for cluster health checking.
type HealthCheckConfig struct {
	// Interval is the interval between health checks.
	// Default: 60 seconds.
	Interval time.Duration

	// Timeout is the timeout for a single health check.
	// Default: 5 seconds.
	Timeout time.Duration

	// FailureThreshold is the number of consecutive failures
	// before marking a cluster as unhealthy.
	// Default: 3.
	FailureThreshold int

	// SuccessThreshold is the number of consecutive successes
	// needed to mark a degraded cluster as healthy again.
	// Default: 2.
	SuccessThreshold int

	// AutoReconnect enables automatic reconnection for unhealthy clusters.
	// Default: true.
	AutoReconnect bool

	// ReconnectBackoff contains the backoff strategy for reconnections.
	ReconnectBackoff BackoffStrategy

	// SyncInterval is the interval for periodic config sync (Pull mode).
	// Default: 5 minutes.
	SyncInterval time.Duration

	// MaxEntries is the maximum number of Informer entries per cluster.
	// 0 means no limit.
	// Default: 100.
	MaxEntries int

	// CleanupInterval is the interval for cleaning up idle Informers.
	// Default: 5 minutes.
	CleanupInterval time.Duration

	// IdleTimeout is the timeout for idle Informers.
	// Default: 10 minutes.
	IdleTimeout time.Duration
}

// BackoffStrategy contains configuration for exponential backoff.
type BackoffStrategy struct {
	// InitialInterval is the initial wait time before the first retry.
	// Default: 1 second.
	InitialInterval time.Duration

	// MaxInterval is the maximum wait time between retries.
	// Default: 60 seconds.
	MaxInterval time.Duration

	// Multiplier is the multiplier applied to the interval after each retry.
	// Default: 2.0.
	Multiplier float64

	// MaxRetries is the maximum number of retries.
	// 0 means unlimited retries.
	// Default: 0.
	MaxRetries int
}

// DefaultHealthCheckConfig is the default health check configuration.
var DefaultHealthCheckConfig = HealthCheckConfig{
	Interval:         60 * time.Second, // 调大间隔，减少 APIServer 压力
	Timeout:          5 * time.Second,
	FailureThreshold: 3,
	SuccessThreshold: 2,
	AutoReconnect:    true,
	ReconnectBackoff: DefaultBackoffStrategy,
	SyncInterval:     5 * time.Minute, // Pull 同步间隔
	MaxEntries:       100,
	CleanupInterval:  5 * time.Minute,
	IdleTimeout:      10 * time.Minute,
}

// DefaultBackoffStrategy is the default backoff strategy.
var DefaultBackoffStrategy = BackoffStrategy{
	InitialInterval: 1 * time.Second,
	MaxInterval:     60 * time.Second,
	Multiplier:      2.0,
	MaxRetries:      0, // Unlimited
}

// RegisterOption is the functional option type for cluster registration.
type RegisterOption func(*RegisterOptions)

// RegisterOptions contains options for cluster registration.
type RegisterOptions struct {
	TenantID string
}

// WithTenantID sets the tenant ID for the cluster.
func WithTenantID(tenantID string) RegisterOption {
	return func(o *RegisterOptions) {
		o.TenantID = tenantID
	}
}

// EventCallbacks contains event callback functions for cluster events.
type EventCallbacks struct {
	OnHealthy          func(id string)
	OnUnhealthy        func(id string)
	OnReconnected      func(id string)
	OnInformerRecreate func(id string)
}

// String returns the string representation of HealthStatus.
func (s HealthStatus) String() string {
	switch s {
	case HealthStatusUnknown:
		return "unknown"
	case HealthStatusHealthy:
		return "healthy"
	case HealthStatusDegraded:
		return "degraded"
	case HealthStatusUnhealthy:
		return "unhealthy"
	case HealthStatusReconnecting:
		return "reconnecting"
	default:
		return "invalid"
	}
}
