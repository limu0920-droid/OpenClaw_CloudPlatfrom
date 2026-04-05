package runtimeadapter

import (
	"fmt"
	"sync"
	"time"
)

type MockAdapter struct {
	mu         sync.RWMutex
	clusters   []Cluster
	nodes      []Node
	namespaces []Namespace
	workloads  []Workload
	pods       []Pod
	metrics    []WorkloadMetrics
}

func NewMockAdapter() *MockAdapter {
	now := time.Now().UTC().Format(time.RFC3339)

	workloads := []Workload{
		{ID: "wl-100", ClusterID: "cluster-sh", Namespace: "tenant-acme", Name: "openclaw-gateway", Kind: "Deployment", Image: "ghcr.io/openclaw/openclaw:1.6.3", Status: "Running", Desired: 2, Ready: 2, Available: 2, LastActionAt: now},
		{ID: "wl-101", ClusterID: "cluster-sh", Namespace: "tenant-acme", Name: "platform-api", Kind: "Deployment", Image: "registry.example.com/platform-api:0.2.0", Status: "Running", Desired: 2, Ready: 2, Available: 2, LastActionAt: now},
		{ID: "wl-200", ClusterID: "cluster-bj", Namespace: "tenant-nova", Name: "openclaw-gateway", Kind: "Deployment", Image: "ghcr.io/openclaw/openclaw:1.6.2", Status: "Running", Desired: 1, Ready: 1, Available: 1, LastActionAt: now},
	}

	return &MockAdapter{
		clusters: []Cluster{
			{ID: "cluster-sh", Code: "c-sh", Name: "Shanghai Cluster", Region: "cn-shanghai", Status: "healthy", Version: "v1.31.2", Nodes: 3, Workloads: 2},
			{ID: "cluster-bj", Code: "c-bj", Name: "Beijing Cluster", Region: "cn-beijing", Status: "warning", Version: "v1.30.8", Nodes: 2, Workloads: 1},
		},
		nodes: []Node{
			{ID: "node-sh-1", ClusterID: "cluster-sh", Name: "sh-worker-01", Role: "worker", Status: "Ready", Kubelet: "v1.31.2", CPUCapacity: "8", MemoryCapacity: "32Gi", PodCount: 18},
			{ID: "node-sh-2", ClusterID: "cluster-sh", Name: "sh-worker-02", Role: "worker", Status: "Ready", Kubelet: "v1.31.2", CPUCapacity: "8", MemoryCapacity: "32Gi", PodCount: 16},
			{ID: "node-sh-3", ClusterID: "cluster-sh", Name: "sh-master-01", Role: "control-plane", Status: "Ready", Kubelet: "v1.31.2", CPUCapacity: "4", MemoryCapacity: "16Gi", PodCount: 11},
			{ID: "node-bj-1", ClusterID: "cluster-bj", Name: "bj-worker-01", Role: "worker", Status: "Ready", Kubelet: "v1.30.8", CPUCapacity: "8", MemoryCapacity: "32Gi", PodCount: 13},
			{ID: "node-bj-2", ClusterID: "cluster-bj", Name: "bj-master-01", Role: "control-plane", Status: "MemoryPressure", Kubelet: "v1.30.8", CPUCapacity: "4", MemoryCapacity: "16Gi", PodCount: 9},
		},
		namespaces: []Namespace{
			{ID: "ns-acme", ClusterID: "cluster-sh", Name: "tenant-acme", Status: "Active", Workloads: 2, Pods: 4},
			{ID: "ns-nova", ClusterID: "cluster-bj", Name: "tenant-nova", Status: "Active", Workloads: 1, Pods: 1},
		},
		workloads: workloads,
		pods: []Pod{
			{ID: "pod-100-1", WorkloadID: "wl-100", Name: "openclaw-gateway-7b79d6c767-rk8cz", NodeName: "sh-worker-01", Status: "Running", Restarts: 0, StartedAt: now},
			{ID: "pod-100-2", WorkloadID: "wl-100", Name: "openclaw-gateway-7b79d6c767-wkls4", NodeName: "sh-worker-02", Status: "Running", Restarts: 1, StartedAt: now},
			{ID: "pod-101-1", WorkloadID: "wl-101", Name: "platform-api-69f8cb6b9d-9xpvd", NodeName: "sh-worker-01", Status: "Running", Restarts: 0, StartedAt: now},
			{ID: "pod-101-2", WorkloadID: "wl-101", Name: "platform-api-69f8cb6b9d-mt25k", NodeName: "sh-worker-02", Status: "Running", Restarts: 0, StartedAt: now},
			{ID: "pod-200-1", WorkloadID: "wl-200", Name: "openclaw-gateway-676d98f8d9-zk8pz", NodeName: "bj-worker-01", Status: "Running", Restarts: 2, StartedAt: now},
		},
		metrics: []WorkloadMetrics{
			{WorkloadID: "wl-100", CPUUsageMilli: 920, MemoryUsageMB: 1820, NetworkRxKB: 84300, NetworkTxKB: 72900, ErrorRatePercent: 1, RequestsPerMinute: 410},
			{WorkloadID: "wl-101", CPUUsageMilli: 340, MemoryUsageMB: 1024, NetworkRxKB: 24100, NetworkTxKB: 21940, ErrorRatePercent: 0, RequestsPerMinute: 128},
			{WorkloadID: "wl-200", CPUUsageMilli: 780, MemoryUsageMB: 1650, NetworkRxKB: 55900, NetworkTxKB: 50110, ErrorRatePercent: 3, RequestsPerMinute: 265},
		},
	}
}

func (m *MockAdapter) ListClusters() []Cluster {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]Cluster(nil), m.clusters...)
}

func (m *MockAdapter) GetCluster(id string) (Cluster, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, item := range m.clusters {
		if item.ID == id {
			return item, true
		}
	}
	return Cluster{}, false
}

func (m *MockAdapter) ListNodes(clusterID string) []Node {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Node, 0)
	for _, item := range m.nodes {
		if clusterID == "" || item.ClusterID == clusterID {
			out = append(out, item)
		}
	}
	return out
}

func (m *MockAdapter) ListNamespaces(clusterID string) []Namespace {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Namespace, 0)
	for _, item := range m.namespaces {
		if clusterID == "" || item.ClusterID == clusterID {
			out = append(out, item)
		}
	}
	return out
}

func (m *MockAdapter) ListWorkloads(clusterID string, namespace string) []Workload {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Workload, 0)
	for _, item := range m.workloads {
		if clusterID != "" && item.ClusterID != clusterID {
			continue
		}
		if namespace != "" && item.Namespace != namespace {
			continue
		}
		out = append(out, item)
	}
	return out
}

func (m *MockAdapter) GetWorkload(id string) (Workload, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, item := range m.workloads {
		if item.ID == id {
			return item, true
		}
	}
	return Workload{}, false
}

func (m *MockAdapter) ListPods(workloadID string) []Pod {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Pod, 0)
	for _, item := range m.pods {
		if workloadID == "" || item.WorkloadID == workloadID {
			out = append(out, item)
		}
	}
	return out
}

func (m *MockAdapter) GetMetrics(workloadID string) (WorkloadMetrics, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, item := range m.metrics {
		if item.WorkloadID == workloadID {
			return item, true
		}
	}
	return WorkloadMetrics{}, false
}

func (m *MockAdapter) StartWorkload(id string) (ActionResult, bool) {
	return m.updateWorkload(id, "start", func(item *Workload) {
		item.Status = "Running"
		if item.Desired == 0 {
			item.Desired = 1
		}
		item.Ready = item.Desired
		item.Available = item.Desired
	})
}

func (m *MockAdapter) StopWorkload(id string) (ActionResult, bool) {
	return m.updateWorkload(id, "stop", func(item *Workload) {
		item.Status = "Stopped"
		item.Ready = 0
		item.Available = 0
	})
}

func (m *MockAdapter) RestartWorkload(id string) (ActionResult, bool) {
	return m.updateWorkload(id, "restart", func(item *Workload) {
		item.Status = "Running"
		item.Ready = item.Desired
		item.Available = item.Desired
	})
}

func (m *MockAdapter) ScaleWorkload(id string, replicas int) (ActionResult, bool) {
	if replicas < 0 {
		replicas = 0
	}
	return m.updateWorkload(id, "scale", func(item *Workload) {
		item.Desired = replicas
		item.Ready = replicas
		item.Available = replicas
		item.Status = "Running"
	})
}

func (m *MockAdapter) updateWorkload(id string, action string, mutate func(item *Workload)) (ActionResult, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for index := range m.workloads {
		if m.workloads[index].ID != id {
			continue
		}

		mutate(&m.workloads[index])
		m.workloads[index].LastActionAt = time.Now().UTC().Format(time.RFC3339)

		affected := make([]string, 0)
		for _, pod := range m.pods {
			if pod.WorkloadID == id {
				affected = append(affected, pod.Name)
			}
		}

		return ActionResult{
			WorkloadID:   id,
			Action:       action,
			Status:       "accepted",
			Message:      fmt.Sprintf("%s workload accepted by mock adapter", action),
			Desired:      m.workloads[index].Desired,
			Ready:        m.workloads[index].Ready,
			Available:    m.workloads[index].Available,
			AffectedPods: affected,
			UpdatedAt:    m.workloads[index].LastActionAt,
		}, true
	}

	return ActionResult{}, false
}
