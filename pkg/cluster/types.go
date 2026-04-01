// Package cluster provides multi-cluster Kubernetes management, including cluster lifecycle,
// health checking, auto-reconnection, and dynamic configuration refresh.
package cluster

import (
	"context"
	"sync"
	"time"

	"github.com/seaman/k8s-kit/pkg/client"
	"github.com/seaman/k8s-kit/pkg/informer"
)

// Manager is the multi-cluster Kubernetes manager.
// It provides cluster registration, health checking, auto-reconnection,
// and configuration refresh capabilities.
type Manager struct {
	mu       sync.RWMutex
	clusters map[string]*ManagedCluster

	clientFactory *client.Factory
	healthConfig  HealthCheckConfig

	ctx    context.Context
	cancel context.CancelFunc

	// Event callbacks
	onClusterHealthy     func(id string)
	onClusterUnhealthy   func(id string)
	onClusterReconnected func(id string)
	onInformerRecreate  func(id string)
}

// ManagedCluster represents a single managed Kubernetes cluster.
type ManagedCluster struct {
	ID        string
	TenantID  string // Optional: tenant ID for multi-tenancy

	mu sync.RWMutex
	// These fields require mutex protection
	client        *client.ClusterClient
	informerMgr   *informerManager
	health        HealthStatus
	failCount     int
	lastCheck     time.Time
	lastAccess    time.Time
	version       string
	recreatingInformer bool
	closed        bool
}

// HealthStatus represents the health state of a cluster.
type HealthStatus int

const (
	// HealthStatusUnknown indicates the cluster health is unknown.
	HealthStatusUnknown HealthStatus = iota
	// HealthStatusHealthy indicates the cluster is healthy.
	HealthStatusHealthy
	// HealthStatusDegraded indicates some health check failures but still operational.
	HealthStatusDegraded
	// HealthStatusUnhealthy indicates the cluster is unhealthy.
	HealthStatusUnhealthy
	// HealthStatusReconnecting indicates the cluster is being reconnected.
	HealthStatusReconnecting
)

// ChangeType represents the type of cluster configuration change.
type ChangeType int

const (
	// ChangeTypeAdd represents a new cluster being added.
	ChangeTypeAdd ChangeType = iota
	// ChangeTypeUpdate represents an existing cluster being updated.
	ChangeTypeUpdate
	// ChangeTypeDelete represents an existing cluster being deleted.
	ChangeTypeDelete
)

// ClusterConfigChange represents a cluster configuration change event.
type ClusterConfigChange struct {
	Type       ChangeType
	ClusterID  string
	TenantID   string
	Kubeconfig []byte // Add/Update events have this field
}

// ConfigProvider is the interface for providing cluster configurations.
type ConfigProvider interface {
	// GetAll returns all cluster configurations.
	GetAll(ctx context.Context) ([]ClusterConfig, error)
}

// ConfigWatcher is the optional interface for watching configuration changes.
type ConfigWatcher interface {
	// Watch watches for configuration changes and returns a channel of events.
	Watch(ctx context.Context) (<-chan ClusterConfigChange, error)
}

// ClusterConfig represents a single cluster configuration.
type ClusterConfig struct {
	ID         string
	TenantID   string
	Kubeconfig []byte
}

// informerManager manages Informers for a single cluster.
type informerManager struct {
	client *client.ClusterClient
	mu     sync.RWMutex
	entries map[string]*informer.Entry
}

// NewInformerManager creates a new informer manager for a cluster.
func newInformerManager(client *client.ClusterClient) *informerManager {
	return &informerManager{
		client:  client,
		entries: make(map[string]*informer.Entry),
	}
}

// stopAll stops all informers in the manager.
func (im *informerManager) stopAll() {
	im.mu.Lock()
	defer im.mu.Unlock()

	for _, entry := range im.entries {
		entry.Stop()
	}
	im.entries = make(map[string]*informer.Entry)
}
