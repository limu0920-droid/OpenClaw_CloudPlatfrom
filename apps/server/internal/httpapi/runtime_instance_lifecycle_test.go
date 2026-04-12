package httpapi

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"openclaw/platformapi/internal/runtimeadapter"
)

func TestHandleCreatePortalInstanceWithProvisionerCreatesBindingAndRuntime(t *testing.T) {
	runtime := newFakeRuntimeAdapter()
	router := &Router{
		data: mockdata.Seed(),
		config: ExternalConfig{
			RuntimeProvider:        "kubectl",
			RuntimeNamespacePrefix: "openclaw",
			RuntimeWorkloadPrefix:  "openclaw",
			RuntimeAccessHost:      "localhost",
			RuntimeAccessScheme:    "http",
			RuntimePort:            18789,
		},
		runtime: runtime,
	}

	recorder := performRequest(t, router, http.MethodPost, "/api/v1/portal/instances", map[string]any{
		"name":   "龙虾测试",
		"plan":   "trial",
		"region": "cn-shanghai",
	})
	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected create status 201, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		Instance struct {
			ID     int    `json:"id"`
			Status string `json:"status"`
		} `json:"instance"`
		Binding *struct {
			ClusterID  string `json:"clusterId"`
			Namespace  string `json:"namespace"`
			WorkloadID string `json:"workloadId"`
		} `json:"binding"`
		Runtime *struct {
			PowerState string `json:"powerState"`
		} `json:"runtime"`
		Workload *struct {
			ID string `json:"id"`
		} `json:"workload"`
	}
	decodeResponse(t, recorder, &response)

	if response.Instance.ID == 0 {
		t.Fatal("expected created instance id")
	}
	if response.Binding == nil || response.Workload == nil || response.Runtime == nil {
		t.Fatalf("expected binding/workload/runtime in response, got body: %s", recorder.Body.String())
	}
	if response.Binding.ClusterID != "docker-desktop" {
		t.Fatalf("expected binding cluster docker-desktop, got %q", response.Binding.ClusterID)
	}
	if response.Binding.WorkloadID != response.Workload.ID {
		t.Fatalf("expected workload id %q, got %q", response.Workload.ID, response.Binding.WorkloadID)
	}
	if response.Runtime.PowerState != "running" {
		t.Fatalf("expected runtime power state running, got %q", response.Runtime.PowerState)
	}

	binding := router.findRuntimeBinding(response.Instance.ID)
	if binding == nil {
		t.Fatal("expected runtime binding persisted")
	}
	if runtimeState := router.findRuntime(response.Instance.ID); runtimeState == nil || runtimeState.PowerState != "running" {
		t.Fatalf("expected persisted runtime state running, got %#v", runtimeState)
	}
	accesses := router.filterAccessByInstance(response.Instance.ID)
	if len(accesses) == 0 || !strings.Contains(accesses[0].URL, "localhost") {
		t.Fatalf("expected runtime access endpoint, got %#v", accesses)
	}
	if credential := router.findCredential(response.Instance.ID); credential == nil || credential.AdminUser == "" {
		t.Fatalf("expected instance credential, got %#v", credential)
	}
}

func TestHandlePortalInstancePowerUsesRuntimeBinding(t *testing.T) {
	runtime := newFakeRuntimeAdapter()
	runtime.workloads["wl-100"] = runtimeadapter.Workload{
		ID:           "wl-100",
		ClusterID:    "docker-desktop",
		Namespace:    "tenant-acme-prod",
		Name:         "openclaw-gateway-prod",
		Kind:         "Deployment",
		Image:        "ghcr.io/openclaw/openclaw:latest",
		Status:       "Running",
		Desired:      1,
		Ready:        1,
		Available:    1,
		LastActionAt: time.Now().UTC().Format(time.RFC3339),
	}
	runtime.metrics["wl-100"] = runtimeadapter.WorkloadMetrics{
		WorkloadID:    "wl-100",
		CPUUsageMilli: 250,
		MemoryUsageMB: 512,
	}
	runtime.access["wl-100"] = []runtimeadapter.AccessEndpoint{{
		WorkloadID:  "wl-100",
		ServiceName: "openclaw-gateway-prod",
		Namespace:   "tenant-acme-prod",
		Type:        "nodePort",
		URL:         "http://localhost:32001",
		Host:        "localhost",
		Port:        32001,
	}}

	router := &Router{
		data:    mockdata.Seed(),
		config:  ExternalConfig{},
		runtime: runtime,
	}

	stopRecorder := performRequest(t, router, http.MethodPost, "/api/v1/portal/instances/100/pause", nil)
	if stopRecorder.Code != http.StatusOK {
		t.Fatalf("expected stop status 200, got %d: %s", stopRecorder.Code, stopRecorder.Body.String())
	}
	if instance, _ := router.findInstance(100); instance.Status != "stopped" {
		t.Fatalf("expected instance status stopped, got %q", instance.Status)
	}
	if runtimeState := router.findRuntime(100); runtimeState == nil || runtimeState.PowerState != "stopped" {
		t.Fatalf("expected runtime state stopped, got %#v", runtimeState)
	}

	startRecorder := performRequest(t, router, http.MethodPost, "/api/v1/portal/instances/100/start", nil)
	if startRecorder.Code != http.StatusOK {
		t.Fatalf("expected start status 200, got %d: %s", startRecorder.Code, startRecorder.Body.String())
	}
	if instance, _ := router.findInstance(100); instance.Status != "running" {
		t.Fatalf("expected instance status running, got %q", instance.Status)
	}
	if runtimeState := router.findRuntime(100); runtimeState == nil || runtimeState.PowerState != "running" {
		t.Fatalf("expected runtime state running, got %#v", runtimeState)
	}
}

func TestHandlePortalInstanceRuntimeRefreshesStoredInstanceStatus(t *testing.T) {
	runtime := newFakeRuntimeAdapter()
	runtime.workloads["wl-100"] = runtimeadapter.Workload{
		ID:           "wl-100",
		ClusterID:    "docker-desktop",
		Namespace:    "tenant-acme-prod",
		Name:         "openclaw-gateway-prod",
		Kind:         "Deployment",
		Image:        "ghcr.io/openclaw/openclaw:latest",
		Status:       "Running",
		Desired:      1,
		Ready:        1,
		Available:    1,
		LastActionAt: time.Now().UTC().Format(time.RFC3339),
	}

	router := &Router{
		data:    mockdata.Seed(),
		config:  ExternalConfig{},
		runtime: runtime,
	}
	router.data.Instances[0].Status = "provisioning"

	recorder := performRequest(t, router, http.MethodGet, "/api/v1/portal/instances/100/runtime", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected runtime status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	if instance, _ := router.findInstance(100); instance.Status != "running" {
		t.Fatalf("expected stored instance status running after runtime refresh, got %q", instance.Status)
	}
}

func TestHandleDeletePortalInstanceMarksInstanceDeletedAndHidesIt(t *testing.T) {
	runtime := newFakeRuntimeAdapter()
	runtime.workloads["wl-100"] = runtimeadapter.Workload{
		ID:           "wl-100",
		ClusterID:    "docker-desktop",
		Namespace:    "tenant-acme-prod",
		Name:         "openclaw-gateway-prod",
		Kind:         "Deployment",
		Image:        "ghcr.io/openclaw/openclaw:latest",
		Status:       "Running",
		Desired:      1,
		Ready:        1,
		Available:    1,
		LastActionAt: time.Now().UTC().Format(time.RFC3339),
	}
	runtime.access["wl-100"] = []runtimeadapter.AccessEndpoint{{
		WorkloadID:  "wl-100",
		ServiceName: "openclaw-gateway-prod",
		Namespace:   "tenant-acme-prod",
		Type:        "nodePort",
		URL:         "http://localhost:32001",
		Host:        "localhost",
		Port:        32001,
	}}

	router := &Router{
		data:    mockdata.Seed(),
		config:  ExternalConfig{},
		runtime: runtime,
	}

	recorder := performRequest(t, router, http.MethodDelete, "/api/v1/portal/instances/100?approvalNo=APR-20260409-005", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected delete status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	instance, found := router.findInstance(100)
	if !found || instance.Status != "deleted" {
		t.Fatalf("expected instance status deleted, got found=%v status=%q", found, instance.Status)
	}
	if binding := router.findRuntimeBinding(100); binding != nil {
		t.Fatalf("expected runtime binding removed, got %#v", binding)
	}
	if runtimeState := router.findRuntime(100); runtimeState != nil {
		t.Fatalf("expected runtime removed, got %#v", runtimeState)
	}
	if credential := router.findCredential(100); credential != nil {
		t.Fatalf("expected credential removed, got %#v", credential)
	}
	if accesses := router.filterAccessByInstance(100); len(accesses) != 0 {
		t.Fatalf("expected accesses removed, got %#v", accesses)
	}
	if _, ok := runtime.GetWorkload("wl-100"); ok {
		t.Fatal("expected runtime workload deleted")
	}

	listRecorder := performRequest(t, router, http.MethodGet, "/api/v1/portal/instances", nil)
	if listRecorder.Code != http.StatusOK {
		t.Fatalf("expected list status 200, got %d: %s", listRecorder.Code, listRecorder.Body.String())
	}
	var listResponse struct {
		Items []struct {
			Instance struct {
				ID int `json:"id"`
			} `json:"instance"`
		} `json:"items"`
	}
	decodeResponse(t, listRecorder, &listResponse)
	for _, item := range listResponse.Items {
		if item.Instance.ID == 100 {
			t.Fatal("expected deleted instance to be hidden from portal list")
		}
	}

	detailRecorder := performRequest(t, router, http.MethodGet, "/api/v1/portal/instances/100", nil)
	if detailRecorder.Code != http.StatusNotFound {
		t.Fatalf("expected deleted instance detail to return 404, got %d", detailRecorder.Code)
	}
}

type fakeRuntimeAdapter struct {
	workloads map[string]runtimeadapter.Workload
	metrics   map[string]runtimeadapter.WorkloadMetrics
	access    map[string][]runtimeadapter.AccessEndpoint
}

func newFakeRuntimeAdapter() *fakeRuntimeAdapter {
	return &fakeRuntimeAdapter{
		workloads: make(map[string]runtimeadapter.Workload),
		metrics:   make(map[string]runtimeadapter.WorkloadMetrics),
		access:    make(map[string][]runtimeadapter.AccessEndpoint),
	}
}

func (f *fakeRuntimeAdapter) ListClusters() []runtimeadapter.Cluster {
	return []runtimeadapter.Cluster{{
		ID:        "docker-desktop",
		Code:      "docker-desktop",
		Name:      "docker-desktop",
		Region:    "local",
		Status:    "healthy",
		Version:   "v1.34.1",
		Nodes:     1,
		Workloads: len(f.workloads),
	}}
}

func (f *fakeRuntimeAdapter) GetCluster(id string) (runtimeadapter.Cluster, bool) {
	if id != "docker-desktop" {
		return runtimeadapter.Cluster{}, false
	}
	return f.ListClusters()[0], true
}

func (f *fakeRuntimeAdapter) ListNodes(clusterID string) []runtimeadapter.Node {
	return nil
}

func (f *fakeRuntimeAdapter) ListNamespaces(clusterID string) []runtimeadapter.Namespace {
	return nil
}

func (f *fakeRuntimeAdapter) ListWorkloads(clusterID string, namespace string) []runtimeadapter.Workload {
	items := make([]runtimeadapter.Workload, 0, len(f.workloads))
	for _, item := range f.workloads {
		if namespace != "" && item.Namespace != namespace {
			continue
		}
		items = append(items, item)
	}
	return items
}

func (f *fakeRuntimeAdapter) GetWorkload(id string) (runtimeadapter.Workload, bool) {
	item, ok := f.workloads[id]
	return item, ok
}

func (f *fakeRuntimeAdapter) ListPods(workloadID string) []runtimeadapter.Pod {
	if _, ok := f.workloads[workloadID]; !ok {
		return nil
	}
	return []runtimeadapter.Pod{{
		ID:         fmt.Sprintf("pod-%s-1", workloadID),
		WorkloadID: workloadID,
		Name:       fmt.Sprintf("%s-pod-1", workloadID),
		NodeName:   "docker-desktop",
		Status:     "Running",
		StartedAt:  time.Now().UTC().Format(time.RFC3339),
	}}
}

func (f *fakeRuntimeAdapter) GetMetrics(workloadID string) (runtimeadapter.WorkloadMetrics, bool) {
	item, ok := f.metrics[workloadID]
	return item, ok
}

func (f *fakeRuntimeAdapter) StartWorkload(id string) (runtimeadapter.ActionResult, bool) {
	item, ok := f.workloads[id]
	if !ok {
		return runtimeadapter.ActionResult{}, false
	}
	item.Status = "Running"
	if item.Desired == 0 {
		item.Desired = 1
	}
	item.Ready = item.Desired
	item.Available = item.Desired
	item.LastActionAt = time.Now().UTC().Format(time.RFC3339)
	f.workloads[id] = item
	return runtimeadapter.ActionResult{
		WorkloadID: id,
		Action:     "start",
		Status:     "accepted",
		Desired:    item.Desired,
		Ready:      item.Ready,
		Available:  item.Available,
		UpdatedAt:  item.LastActionAt,
	}, true
}

func (f *fakeRuntimeAdapter) StopWorkload(id string) (runtimeadapter.ActionResult, bool) {
	item, ok := f.workloads[id]
	if !ok {
		return runtimeadapter.ActionResult{}, false
	}
	item.Status = "Stopped"
	item.Ready = 0
	item.Available = 0
	item.LastActionAt = time.Now().UTC().Format(time.RFC3339)
	f.workloads[id] = item
	return runtimeadapter.ActionResult{
		WorkloadID: id,
		Action:     "stop",
		Status:     "accepted",
		Desired:    item.Desired,
		Ready:      item.Ready,
		Available:  item.Available,
		UpdatedAt:  item.LastActionAt,
	}, true
}

func (f *fakeRuntimeAdapter) RestartWorkload(id string) (runtimeadapter.ActionResult, bool) {
	return f.StartWorkload(id)
}

func (f *fakeRuntimeAdapter) ScaleWorkload(id string, replicas int) (runtimeadapter.ActionResult, bool) {
	item, ok := f.workloads[id]
	if !ok {
		return runtimeadapter.ActionResult{}, false
	}
	item.Desired = replicas
	item.Ready = replicas
	item.Available = replicas
	if replicas == 0 {
		item.Status = "Stopped"
	} else {
		item.Status = "Running"
	}
	item.LastActionAt = time.Now().UTC().Format(time.RFC3339)
	f.workloads[id] = item
	return runtimeadapter.ActionResult{
		WorkloadID: id,
		Action:     "scale",
		Status:     "accepted",
		Desired:    item.Desired,
		Ready:      item.Ready,
		Available:  item.Available,
		UpdatedAt:  item.LastActionAt,
	}, true
}

func (f *fakeRuntimeAdapter) CreateWorkload(req runtimeadapter.CreateWorkloadRequest) (runtimeadapter.CreateWorkloadResult, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	workload := runtimeadapter.Workload{
		ID:           req.WorkloadID,
		ClusterID:    "docker-desktop",
		Namespace:    req.Namespace,
		Name:         req.Name,
		Kind:         "Deployment",
		Image:        req.Image,
		Status:       "Running",
		Desired:      1,
		Ready:        1,
		Available:    1,
		LastActionAt: now,
	}
	f.workloads[req.WorkloadID] = workload
	f.metrics[req.WorkloadID] = runtimeadapter.WorkloadMetrics{
		WorkloadID:    req.WorkloadID,
		CPUUsageMilli: 120,
		MemoryUsageMB: 256,
	}
	f.access[req.WorkloadID] = []runtimeadapter.AccessEndpoint{
		{
			WorkloadID:  req.WorkloadID,
			ServiceName: req.Name,
			Namespace:   req.Namespace,
			Type:        "nodePort",
			URL:         "http://localhost:32089",
			Host:        "localhost",
			Port:        32089,
		},
		{
			WorkloadID:  req.WorkloadID,
			ServiceName: req.Name,
			Namespace:   req.Namespace,
			Type:        "clusterIP",
			URL:         fmt.Sprintf("http://%s.%s.svc.cluster.local:80", req.Name, req.Namespace),
			Host:        fmt.Sprintf("%s.%s.svc.cluster.local", req.Name, req.Namespace),
			Port:        80,
		},
	}
	return runtimeadapter.CreateWorkloadResult{
		Workload:        workload,
		AccessEndpoints: f.access[req.WorkloadID],
	}, nil
}

func (f *fakeRuntimeAdapter) DeleteWorkload(req runtimeadapter.DeleteWorkloadRequest) (runtimeadapter.ActionResult, error) {
	delete(f.workloads, req.WorkloadID)
	delete(f.metrics, req.WorkloadID)
	delete(f.access, req.WorkloadID)
	return runtimeadapter.ActionResult{
		WorkloadID: req.WorkloadID,
		Action:     "delete",
		Status:     "accepted",
		UpdatedAt:  time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (f *fakeRuntimeAdapter) GetWorkloadAccess(id string) []runtimeadapter.AccessEndpoint {
	return f.access[id]
}
