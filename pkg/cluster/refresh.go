package cluster

import (
	"context"
	"fmt"
	"log"
	"time"
)

// Start starts the ClusterManager with configuration refresh capabilities.
// It loads all clusters from the config provider on startup and watches for changes.
func (m *Manager) Start(ctx context.Context, configProvider ConfigProvider) error {
	// Load all clusters on startup
	if err := m.loadAllClusters(ctx, configProvider); err != nil {
		return fmt.Errorf("failed to load clusters: %w", err)
	}

	// Start watching for config changes (Push mode)
	if watcher, ok := configProvider.(ConfigWatcher); ok {
		go m.watchConfigChanges(ctx, watcher)
	}

	// Start periodic sync as fallback (Pull mode)
	go m.syncLoop(ctx, configProvider)

	return nil
}

// loadAllClusters loads all clusters from the config provider.
func (m *Manager) loadAllClusters(ctx context.Context, provider ConfigProvider) error {
	configs, err := provider.GetAll(ctx)
	if err != nil {
		return err
	}

	for _, cfg := range configs {
		if err := m.Register(cfg.ID, cfg.Kubeconfig, WithTenantID(cfg.TenantID)); err != nil {
			log.Printf("Failed to register cluster %s: %v", cfg.ID, err)
		}
	}

	return nil
}

// watchConfigChanges watches for configuration changes in real-time.
func (m *Manager) watchConfigChanges(ctx context.Context, watcher ConfigWatcher) {
	eventCh, err := watcher.Watch(ctx)
	if err != nil {
		log.Printf("Failed to watch config changes: %v", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.ctx.Done():
			return
		case event := <-eventCh:
			m.handleConfigChange(event)
		}
	}
}

// handleConfigChange handles a cluster configuration change.
func (m *Manager) handleConfigChange(event ClusterConfigChange) {
	switch event.Type {
	case ChangeTypeAdd:
		if err := m.Register(event.ClusterID, event.Kubeconfig, WithTenantID(event.TenantID)); err != nil {
			log.Printf("Failed to register cluster %s: %v", event.ClusterID, err)
		}
	case ChangeTypeUpdate:
		if err := m.Update(event.ClusterID, event.Kubeconfig); err != nil {
			log.Printf("Failed to update cluster %s: %v", event.ClusterID, err)
		}
	case ChangeTypeDelete:
		if err := m.Unregister(event.ClusterID); err != nil {
			log.Printf("Failed to unregister cluster %s: %v", event.ClusterID, err)
		}
	}
}

// syncLoop periodically syncs with the config provider as a fallback.
func (m *Manager) syncLoop(ctx context.Context, provider ConfigProvider) {
	syncInterval := m.healthConfig.SyncInterval
	if syncInterval <= 0 {
		syncInterval = 5 * time.Minute // 默认值
	}
	ticker := time.NewTicker(syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			if err := m.syncWithProvider(ctx, provider); err != nil {
				log.Printf("Failed to sync clusters: %v", err)
			}
		}
	}
}

// syncWithProvider syncs the manager state with the config provider.
func (m *Manager) syncWithProvider(ctx context.Context, provider ConfigProvider) error {
	configs, err := provider.GetAll(ctx)
	if err != nil {
		return err
	}

	currentIDs := make(map[string]bool)
	for _, cfg := range configs {
		currentIDs[cfg.ID] = true
		//此处的getcluster只使用ok吗？
		_, ok := m.getCluster(cfg.ID)
		if !ok {
			// New cluster
			m.Register(cfg.ID, cfg.Kubeconfig, WithTenantID(cfg.TenantID))
		}
		// Note: Update logic would go here in a full implementation
	}

	// Delete clusters that no longer exist
	m.mu.RLock()
	for id := range m.clusters {
		if !currentIDs[id] {
			go m.Unregister(id)
		}
	}
	m.mu.RUnlock()

	return nil
}

// Update updates an existing cluster's configuration.
func (m *Manager) Update(id string, kubeconfig []byte) error {
	m.mu.Lock()
	cluster, ok := m.clusters[id]
	m.mu.Unlock()

	if !ok {
		return fmt.Errorf("cluster %s not found", id)
	}

	// Create new client
	cli, err := m.clientFactory.CreateFromKubeconfig(kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to create client for cluster %s: %w", id, err)
	}

	cluster.mu.Lock()
	oldClient := cluster.client
	cluster.client = cli

	// Recreate informer manager if it exists
	if cluster.informerMgr != nil {
		cluster.informerMgr.client = cli
	}
	cluster.mu.Unlock()

	_ = oldClient // In production, properly close the old client

	return nil
}
