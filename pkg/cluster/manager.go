package cluster

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/guilinonline/k8s-kit/pkg/client"
	"github.com/guilinonline/k8s-kit/pkg/tenant"
)

// NewManager creates a new ClusterManager.
func NewManager(factory *client.Factory, healthCfg HealthCheckConfig) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	return &Manager{
		clusters:      make(map[string]*ManagedCluster),
		clientFactory: factory,
		healthConfig:  healthCfg,
		ctx:           ctx,
		cancel:        cancel,
	}
}

// SetEventCallbacks sets event callbacks for cluster events.
func (m *Manager) SetEventCallbacks(callbacks EventCallbacks) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onClusterHealthy = callbacks.OnHealthy
	m.onClusterUnhealthy = callbacks.OnUnhealthy
	m.onClusterReconnected = callbacks.OnReconnected
	m.onInformerRecreate = callbacks.OnInformerRecreate
}

// Register registers a new cluster with the manager.
func (m *Manager) Register(id string, kubeconfig []byte, opts ...RegisterOption) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.clusters[id]; ok {
		return fmt.Errorf("cluster %s is already registered", id)
	}

	options := &RegisterOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// 构建 Factory 选项
	var clientOpts []client.Option
	if options.DialContext != nil {
		clientOpts = append(clientOpts, client.WithDialContext(options.DialContext))
	}

	cli, err := m.clientFactory.CreateFromKubeconfig(kubeconfig, clientOpts...)
	if err != nil {
		return fmt.Errorf("failed to create client for cluster %s: %w", id, err)
	}

	cluster := &ManagedCluster{
		ID:         id,
		TenantID:   options.TenantID,
		client:     cli,
		health:     HealthStatusHealthy,
		lastCheck:  time.Now(),
		lastAccess: time.Now(),
	}

	m.clusters[id] = cluster

	go m.healthCheckLoop(cluster)

	return nil
}

// Unregister unregisters a cluster from the manager.
func (m *Manager) Unregister(id string) error {
	m.mu.Lock()
	cluster, ok := m.clusters[id]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("cluster %s not found", id)
	}
	delete(m.clusters, id)
	m.mu.Unlock()

	cluster.mu.Lock()
	cluster.closed = true
	if cluster.informerMgr != nil {
		cluster.informerMgr.stopAll()
	}
	cluster.mu.Unlock()

	return nil
}

// GetClient gets the client for a registered cluster.
func (m *Manager) GetClient(id string) (*client.ClusterClient, error) {
	m.mu.RLock()
	cluster, ok := m.clusters[id]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("cluster %s not found", id)
	}

	cluster.mu.Lock()
	cluster.lastAccess = time.Now()
	cluster.mu.Unlock()

	return cluster.client, nil
}

// GetClientFromContext gets the client for a registered cluster,
// using tenant and cluster information from the context.
func (m *Manager) GetClientFromContext(ctx context.Context) (*client.ClusterClient, error) {
	tenantID := tenant.FromContext(ctx)
	clusterID := tenant.ClusterFromContext(ctx)

	if clusterID == "" {
		return nil, fmt.Errorf("cluster ID not found in context")
	}

	cli, err := m.GetClient(clusterID)
	if err != nil {
		return nil, err
	}

	_ = tenantID // Used for access validation in extended version

	return cli, nil
}

// List lists all registered clusters.
func (m *Manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.clusters))
	for id := range m.clusters {
		ids = append(ids, id)
	}
	return ids
}

// GetHealthStatus gets the health status of a cluster.
func (m *Manager) GetHealthStatus(id string) (HealthStatus, error) {
	m.mu.RLock()
	cluster, ok := m.clusters[id]
	m.mu.RUnlock()

	if !ok {
		return HealthStatusUnknown, fmt.Errorf("cluster %s not found", id)
	}

	cluster.mu.RLock()
	defer cluster.mu.RUnlock()
	return cluster.health, nil
}

// Stop stops the manager and all health check loops.
func (m *Manager) Stop() {
	m.cancel()

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, cluster := range m.clusters {
		cluster.mu.Lock()
		cluster.closed = true
		if cluster.informerMgr != nil {
			cluster.informerMgr.stopAll()
		}
		cluster.mu.Unlock()
	}
}

func (m *Manager) getCluster(id string) (*ManagedCluster, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cluster, ok := m.clusters[id]
	return cluster, ok
}

func (m *Manager) healthCheckLoop(cluster *ManagedCluster) {
	ticker := time.NewTicker(m.healthConfig.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			cluster.mu.RLock()
			if cluster.closed {
				cluster.mu.RUnlock()
				return
			}
			cluster.mu.RUnlock()

			m.checkHealth(cluster)
		}
	}
}

func (m *Manager) checkHealth(cluster *ManagedCluster) {
	//此处的ctx是否在使用
	_, cancel := context.WithTimeout(m.ctx, m.healthConfig.Timeout)
	defer cancel()

	_, err := cluster.client.Clientset.Discovery().ServerVersion()

	cluster.mu.Lock()
	defer cluster.mu.Unlock()

	cluster.lastCheck = time.Now()

	if err != nil {
		cluster.failCount++

		if cluster.failCount >= m.healthConfig.FailureThreshold {
			if cluster.health != HealthStatusUnhealthy {
				cluster.health = HealthStatusUnhealthy
				if m.onClusterUnhealthy != nil {
					go m.onClusterUnhealthy(cluster.ID)
				}
				if m.healthConfig.AutoReconnect {
					go m.reconnect(cluster)
				}
			}
		} else if cluster.failCount >= m.healthConfig.FailureThreshold/2 {
			if cluster.health == HealthStatusHealthy {
				cluster.health = HealthStatusDegraded
			}
		}
	} else {
		if cluster.failCount > 0 {
			cluster.failCount = 0
			cluster.health = HealthStatusHealthy
			if m.onClusterHealthy != nil {
				go m.onClusterHealthy(cluster.ID)
			}
		}
	}
}

// reconnect attempts to reconnect to an unhealthy cluster with exponential backoff.
func (m *Manager) reconnect(cluster *ManagedCluster) {
	backoff := m.healthConfig.ReconnectBackoff
	attempt := 0

	for {
		select {
		case <-m.ctx.Done():
			return
		default:
		}

		cluster.mu.RLock()
		if cluster.closed || cluster.health == HealthStatusHealthy {
			cluster.mu.RUnlock()
			return
		}
		cluster.mu.RUnlock()

		// Calculate backoff time
		waitTime := time.Duration(float64(backoff.InitialInterval) * math.Pow(backoff.Multiplier, float64(attempt)))
		if waitTime > backoff.MaxInterval {
			waitTime = backoff.MaxInterval
		}

		select {
		case <-m.ctx.Done():
			return
		case <-time.After(waitTime):
		}

		attempt++

		// Check max retries
		if backoff.MaxRetries > 0 && attempt >= backoff.MaxRetries {
			cluster.mu.Lock()
			cluster.health = HealthStatusUnhealthy
			cluster.mu.Unlock()
			return
		}

		// Try to reconnect - in a full implementation, we would get fresh kubeconfig
		// For now, we just mark as healthy and let the next health check verify
		cluster.mu.Lock()
		if cluster.health == HealthStatusUnhealthy {
			cluster.health = HealthStatusHealthy
			cluster.failCount = 0
			if m.onClusterReconnected != nil {
				go m.onClusterReconnected(cluster.ID)
			}
			if m.onInformerRecreate != nil {
				go m.onInformerRecreate(cluster.ID)
			}
		}
		cluster.mu.Unlock()

		return
	}
}
