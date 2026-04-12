package runtimeadapter

import "fmt"

type HybridAdapter struct {
	providers []Adapter
}

func NewHybridAdapter(providers ...Adapter) *HybridAdapter {
	filtered := make([]Adapter, 0, len(providers))
	for _, provider := range providers {
		if provider != nil {
			filtered = append(filtered, provider)
		}
	}
	return &HybridAdapter{providers: filtered}
}

func (h *HybridAdapter) ListClusters() []Cluster {
	items := make([]Cluster, 0)
	seen := make(map[string]struct{})
	for _, provider := range h.providers {
		for _, item := range provider.ListClusters() {
			if _, ok := seen[item.ID]; ok {
				continue
			}
			seen[item.ID] = struct{}{}
			items = append(items, item)
		}
	}
	return items
}

func (h *HybridAdapter) GetCluster(id string) (Cluster, bool) {
	for _, provider := range h.providers {
		if item, ok := provider.GetCluster(id); ok {
			return item, true
		}
	}
	return Cluster{}, false
}

func (h *HybridAdapter) ListNodes(clusterID string) []Node {
	items := make([]Node, 0)
	for _, provider := range h.providers {
		items = append(items, provider.ListNodes(clusterID)...)
	}
	return items
}

func (h *HybridAdapter) ListNamespaces(clusterID string) []Namespace {
	items := make([]Namespace, 0)
	for _, provider := range h.providers {
		items = append(items, provider.ListNamespaces(clusterID)...)
	}
	return items
}

func (h *HybridAdapter) ListWorkloads(clusterID string, namespace string) []Workload {
	items := make([]Workload, 0)
	for _, provider := range h.providers {
		items = append(items, provider.ListWorkloads(clusterID, namespace)...)
	}
	return items
}

func (h *HybridAdapter) GetWorkload(id string) (Workload, bool) {
	for _, provider := range h.providers {
		if item, ok := provider.GetWorkload(id); ok {
			return item, true
		}
	}
	return Workload{}, false
}

func (h *HybridAdapter) ListPods(workloadID string) []Pod {
	for _, provider := range h.providers {
		if workload, ok := provider.GetWorkload(workloadID); ok {
			return provider.ListPods(workload.ID)
		}
	}
	return nil
}

func (h *HybridAdapter) GetMetrics(workloadID string) (WorkloadMetrics, bool) {
	for _, provider := range h.providers {
		if metrics, ok := provider.GetMetrics(workloadID); ok {
			return metrics, true
		}
	}
	return WorkloadMetrics{}, false
}

func (h *HybridAdapter) StartWorkload(id string) (ActionResult, bool) {
	for _, provider := range h.providers {
		if result, ok := provider.StartWorkload(id); ok {
			return result, true
		}
	}
	return ActionResult{}, false
}

func (h *HybridAdapter) StopWorkload(id string) (ActionResult, bool) {
	for _, provider := range h.providers {
		if result, ok := provider.StopWorkload(id); ok {
			return result, true
		}
	}
	return ActionResult{}, false
}

func (h *HybridAdapter) RestartWorkload(id string) (ActionResult, bool) {
	for _, provider := range h.providers {
		if result, ok := provider.RestartWorkload(id); ok {
			return result, true
		}
	}
	return ActionResult{}, false
}

func (h *HybridAdapter) ScaleWorkload(id string, replicas int) (ActionResult, bool) {
	for _, provider := range h.providers {
		if result, ok := provider.ScaleWorkload(id, replicas); ok {
			return result, true
		}
	}
	return ActionResult{}, false
}

func (h *HybridAdapter) CreateWorkload(req CreateWorkloadRequest) (CreateWorkloadResult, error) {
	for _, provider := range h.providers {
		if provisioner, ok := provider.(ProvisioningAdapter); ok {
			return provisioner.CreateWorkload(req)
		}
	}
	return CreateWorkloadResult{}, fmt.Errorf("no provisioning adapter available")
}

func (h *HybridAdapter) DeleteWorkload(req DeleteWorkloadRequest) (ActionResult, error) {
	for _, provider := range h.providers {
		if deleter, ok := provider.(DeletionAdapter); ok {
			return deleter.DeleteWorkload(req)
		}
	}
	return ActionResult{}, fmt.Errorf("no deletion adapter available")
}

func (h *HybridAdapter) GetWorkloadAccess(id string) []AccessEndpoint {
	for _, provider := range h.providers {
		accessProvider, ok := provider.(AccessInfoProvider)
		if !ok {
			continue
		}
		items := accessProvider.GetWorkloadAccess(id)
		if len(items) > 0 {
			return items
		}
	}
	return nil
}

func (h *HybridAdapter) ExecuteDiagnosticCommand(req DiagnosticExecRequest) (DiagnosticExecResult, error) {
	for _, provider := range h.providers {
		execProvider, ok := provider.(DiagnosticExecProvider)
		if !ok {
			continue
		}
		result, err := execProvider.ExecuteDiagnosticCommand(req)
		if err == nil {
			return result, nil
		}
	}
	return DiagnosticExecResult{}, fmt.Errorf("no diagnostic exec provider available")
}
