package cluster

import (
	"fmt"

	"github.com/guilinonline/k8s-kit/pkg/client"
	"github.com/guilinonline/k8s-kit/pkg/informer"
)

// GetOrCreateInformer gets or creates an Informer for the specified cluster.
// The informer is created lazily on first access.
func (m *Manager) GetOrCreateInformer(clusterID string, opts informer.Options) (*informer.Entry, error) {
	cluster, ok := m.getCluster(clusterID)
	if !ok {
		return nil, fmt.Errorf("cluster %s not found", clusterID)
	}

	cluster.mu.Lock()
	defer cluster.mu.Unlock()

	if cluster.informerMgr == nil {
		cluster.informerMgr = newInformerManager(cluster.client)
	}

	return cluster.informerMgr.getOrCreate(opts)
}

// StopInformer stops a specific informer for a cluster.
func (m *Manager) StopInformer(clusterID string, key string) error {
	cluster, ok := m.getCluster(clusterID)
	if !ok {
		return fmt.Errorf("cluster %s not found", clusterID)
	}

	cluster.mu.RLock()
	informerMgr := cluster.informerMgr
	cluster.mu.RUnlock()

	if informerMgr == nil {
		return nil
	}

	return informerMgr.stop(key)
}

// StopAllInformers stops all informers for a specific cluster.
func (m *Manager) StopAllInformers(clusterID string) error {
	cluster, ok := m.getCluster(clusterID)
	if !ok {
		return fmt.Errorf("cluster %s not found", clusterID)
	}

	cluster.mu.RLock()
	informerMgr := cluster.informerMgr
	cluster.mu.RUnlock()

	if informerMgr != nil {
		informerMgr.stopAll()
	}

	return nil
}

func (im *informerManager) getOrCreate(opts informer.Options) (*informer.Entry, error) {
	key := generateInformerKey(opts)

	im.mu.RLock()
	entry, ok := im.entries[key]
	im.mu.RUnlock()

	if ok {
		entry.UpdateAccessTime()
		return entry, nil
	}

	im.mu.Lock()
	defer im.mu.Unlock()

	// Double check
	if entry, ok := im.entries[key]; ok {
		return entry, nil
	}

	entry, err := informer.NewFactory().Create(im.client, opts)
	if err != nil {
		return nil, err
	}

	im.entries[key] = entry
	return entry, nil
}

func (im *informerManager) stop(key string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	entry, ok := im.entries[key]
	if !ok {
		return nil
	}

	entry.Stop()
	delete(im.entries, key)
	return nil
}

func generateInformerKey(opts informer.Options) string {
	// Generate a unique key based on namespace and resource type
	return fmt.Sprintf("%s/%s", opts.Namespace, "pods")
}

// OnClientRecreated is called when a cluster's client is recreated (e.g., after reconnection).
func (m *Manager) OnClientRecreated(clusterID string, newClient *client.ClusterClient) error {
	cluster, ok := m.getCluster(clusterID)
	if !ok {
		return fmt.Errorf("cluster %s not found", clusterID)
	}

	cluster.mu.Lock()
	cluster.client = newClient
	if cluster.informerMgr != nil {
		cluster.informerMgr.client = newClient
		// Note: In production, you might want to recreate the informers
		// with the new client to ensure cache consistency
	}
	cluster.mu.Unlock()

	// Notify through callback
	if m.onInformerRecreate != nil {
		go m.onInformerRecreate(clusterID)
	}

	return nil
}
