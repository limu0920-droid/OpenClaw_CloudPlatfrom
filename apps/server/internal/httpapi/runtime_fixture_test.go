package httpapi

import (
	"fmt"
	"strings"
	"sync"
	"time"

	runtimeadapterpkg "openclaw/platformapi/internal/runtimeadapter"
)

type testRuntimeAdapter struct {
	mu         sync.RWMutex
	clusters   []runtimeadapterpkg.Cluster
	nodes      []runtimeadapterpkg.Node
	namespaces []runtimeadapterpkg.Namespace
	workloads  []runtimeadapterpkg.Workload
	pods       []runtimeadapterpkg.Pod
	metrics    []runtimeadapterpkg.WorkloadMetrics
}

func newTestRuntimeAdapter() *testRuntimeAdapter {
	now := time.Now().UTC().Format(time.RFC3339)

	workloads := []runtimeadapterpkg.Workload{
		{ID: "wl-100", ClusterID: "cluster-sh", Namespace: "tenant-acme-prod", Name: "openclaw-gateway-prod", Kind: "Deployment", Image: "ghcr.io/openclaw/openclaw:1.6.3", Status: "Running", Desired: 2, Ready: 2, Available: 2, LastActionAt: now},
		{ID: "wl-101", ClusterID: "cluster-sh", Namespace: "tenant-acme-stg", Name: "openclaw-gateway-stg", Kind: "Deployment", Image: "ghcr.io/openclaw/openclaw:1.6.3", Status: "Running", Desired: 1, Ready: 1, Available: 1, LastActionAt: now},
		{ID: "wl-core-1", ClusterID: "cluster-sh", Namespace: "platform-system", Name: "platform-api", Kind: "Deployment", Image: "registry.example.com/platform-api:0.2.0", Status: "Running", Desired: 2, Ready: 2, Available: 2, LastActionAt: now},
		{ID: "wl-200", ClusterID: "cluster-bj", Namespace: "tenant-nova-prod", Name: "openclaw-gateway-prod", Kind: "Deployment", Image: "ghcr.io/openclaw/openclaw:1.6.2", Status: "Running", Desired: 1, Ready: 1, Available: 1, LastActionAt: now},
	}

	return &testRuntimeAdapter{
		clusters: []runtimeadapterpkg.Cluster{
			{ID: "cluster-sh", Code: "c-sh", Name: "Shanghai Cluster", Region: "cn-shanghai", Status: "healthy", Version: "v1.31.2", Nodes: 3, Workloads: 3},
			{ID: "cluster-bj", Code: "c-bj", Name: "Beijing Cluster", Region: "cn-beijing", Status: "warning", Version: "v1.30.8", Nodes: 2, Workloads: 1},
		},
		nodes: []runtimeadapterpkg.Node{
			{ID: "node-sh-1", ClusterID: "cluster-sh", Name: "sh-worker-01", Role: "worker", Status: "Ready", Kubelet: "v1.31.2", CPUCapacity: "8", MemoryCapacity: "32Gi", PodCount: 18},
			{ID: "node-sh-2", ClusterID: "cluster-sh", Name: "sh-worker-02", Role: "worker", Status: "Ready", Kubelet: "v1.31.2", CPUCapacity: "8", MemoryCapacity: "32Gi", PodCount: 16},
			{ID: "node-sh-3", ClusterID: "cluster-sh", Name: "sh-master-01", Role: "control-plane", Status: "Ready", Kubelet: "v1.31.2", CPUCapacity: "4", MemoryCapacity: "16Gi", PodCount: 11},
			{ID: "node-bj-1", ClusterID: "cluster-bj", Name: "bj-worker-01", Role: "worker", Status: "Ready", Kubelet: "v1.30.8", CPUCapacity: "8", MemoryCapacity: "32Gi", PodCount: 13},
			{ID: "node-bj-2", ClusterID: "cluster-bj", Name: "bj-master-01", Role: "control-plane", Status: "MemoryPressure", Kubelet: "v1.30.8", CPUCapacity: "4", MemoryCapacity: "16Gi", PodCount: 9},
		},
		namespaces: []runtimeadapterpkg.Namespace{
			{ID: "ns-acme-prod", ClusterID: "cluster-sh", Name: "tenant-acme-prod", Status: "Active", Workloads: 1, Pods: 2},
			{ID: "ns-acme-stg", ClusterID: "cluster-sh", Name: "tenant-acme-stg", Status: "Active", Workloads: 1, Pods: 1},
			{ID: "ns-platform", ClusterID: "cluster-sh", Name: "platform-system", Status: "Active", Workloads: 1, Pods: 2},
			{ID: "ns-nova", ClusterID: "cluster-bj", Name: "tenant-nova-prod", Status: "Active", Workloads: 1, Pods: 1},
		},
		workloads: workloads,
		pods: []runtimeadapterpkg.Pod{
			{ID: "pod-100-1", WorkloadID: "wl-100", Name: "openclaw-gateway-prod-7b79d6c767-rk8cz", NodeName: "sh-worker-01", Status: "Running", Restarts: 0, StartedAt: now},
			{ID: "pod-100-2", WorkloadID: "wl-100", Name: "openclaw-gateway-prod-7b79d6c767-wkls4", NodeName: "sh-worker-02", Status: "Running", Restarts: 1, StartedAt: now},
			{ID: "pod-101-1", WorkloadID: "wl-101", Name: "openclaw-gateway-stg-69f8cb6b9d-9xpvd", NodeName: "sh-worker-01", Status: "Running", Restarts: 0, StartedAt: now},
			{ID: "pod-core-1", WorkloadID: "wl-core-1", Name: "platform-api-69f8cb6b9d-9xpvd", NodeName: "sh-worker-01", Status: "Running", Restarts: 0, StartedAt: now},
			{ID: "pod-core-2", WorkloadID: "wl-core-1", Name: "platform-api-69f8cb6b9d-mt25k", NodeName: "sh-worker-02", Status: "Running", Restarts: 0, StartedAt: now},
			{ID: "pod-200-1", WorkloadID: "wl-200", Name: "openclaw-gateway-676d98f8d9-zk8pz", NodeName: "bj-worker-01", Status: "Running", Restarts: 2, StartedAt: now},
		},
		metrics: []runtimeadapterpkg.WorkloadMetrics{
			{WorkloadID: "wl-100", CPUUsageMilli: 920, MemoryUsageMB: 1820, NetworkRxKB: 84300, NetworkTxKB: 72900, ErrorRatePercent: 1, RequestsPerMinute: 410},
			{WorkloadID: "wl-101", CPUUsageMilli: 240, MemoryUsageMB: 768, NetworkRxKB: 14100, NetworkTxKB: 11940, ErrorRatePercent: 0, RequestsPerMinute: 64},
			{WorkloadID: "wl-core-1", CPUUsageMilli: 340, MemoryUsageMB: 1024, NetworkRxKB: 24100, NetworkTxKB: 21940, ErrorRatePercent: 0, RequestsPerMinute: 128},
			{WorkloadID: "wl-200", CPUUsageMilli: 780, MemoryUsageMB: 1650, NetworkRxKB: 55900, NetworkTxKB: 50110, ErrorRatePercent: 3, RequestsPerMinute: 265},
		},
	}
}

func (m *testRuntimeAdapter) ListClusters() []runtimeadapterpkg.Cluster {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]runtimeadapterpkg.Cluster(nil), m.clusters...)
}

func (m *testRuntimeAdapter) GetCluster(id string) (runtimeadapterpkg.Cluster, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, item := range m.clusters {
		if item.ID == id {
			return item, true
		}
	}
	return runtimeadapterpkg.Cluster{}, false
}

func (m *testRuntimeAdapter) ListNodes(clusterID string) []runtimeadapterpkg.Node {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]runtimeadapterpkg.Node, 0)
	for _, item := range m.nodes {
		if clusterID == "" || item.ClusterID == clusterID {
			out = append(out, item)
		}
	}
	return out
}

func (m *testRuntimeAdapter) ListNamespaces(clusterID string) []runtimeadapterpkg.Namespace {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]runtimeadapterpkg.Namespace, 0)
	for _, item := range m.namespaces {
		if clusterID == "" || item.ClusterID == clusterID {
			out = append(out, item)
		}
	}
	return out
}

func (m *testRuntimeAdapter) ListWorkloads(clusterID string, namespace string) []runtimeadapterpkg.Workload {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]runtimeadapterpkg.Workload, 0)
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

func (m *testRuntimeAdapter) GetWorkload(id string) (runtimeadapterpkg.Workload, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, item := range m.workloads {
		if item.ID == id {
			return item, true
		}
	}
	return runtimeadapterpkg.Workload{}, false
}

func (m *testRuntimeAdapter) ListPods(workloadID string) []runtimeadapterpkg.Pod {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]runtimeadapterpkg.Pod, 0)
	for _, item := range m.pods {
		if workloadID == "" || item.WorkloadID == workloadID {
			out = append(out, item)
		}
	}
	return out
}

func (m *testRuntimeAdapter) GetMetrics(workloadID string) (runtimeadapterpkg.WorkloadMetrics, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, item := range m.metrics {
		if item.WorkloadID == workloadID {
			return item, true
		}
	}
	return runtimeadapterpkg.WorkloadMetrics{}, false
}

func (m *testRuntimeAdapter) StartWorkload(id string) (runtimeadapterpkg.ActionResult, bool) {
	return m.updateWorkload(id, "start", func(item *runtimeadapterpkg.Workload) {
		item.Status = "Running"
		if item.Desired == 0 {
			item.Desired = 1
		}
		item.Ready = item.Desired
		item.Available = item.Desired
	})
}

func (m *testRuntimeAdapter) StopWorkload(id string) (runtimeadapterpkg.ActionResult, bool) {
	return m.updateWorkload(id, "stop", func(item *runtimeadapterpkg.Workload) {
		item.Status = "Stopped"
		item.Ready = 0
		item.Available = 0
	})
}

func (m *testRuntimeAdapter) RestartWorkload(id string) (runtimeadapterpkg.ActionResult, bool) {
	return m.updateWorkload(id, "restart", func(item *runtimeadapterpkg.Workload) {
		item.Status = "Running"
		item.Ready = item.Desired
		item.Available = item.Desired
	})
}

func (m *testRuntimeAdapter) ScaleWorkload(id string, replicas int) (runtimeadapterpkg.ActionResult, bool) {
	if replicas < 0 {
		replicas = 0
	}
	return m.updateWorkload(id, "scale", func(item *runtimeadapterpkg.Workload) {
		item.Desired = replicas
		item.Ready = replicas
		item.Available = replicas
		item.Status = "Running"
	})
}

func (m *testRuntimeAdapter) DeleteWorkload(req runtimeadapterpkg.DeleteWorkloadRequest) (runtimeadapterpkg.ActionResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	index := -1
	targetID := req.WorkloadID
	for i, item := range m.workloads {
		if item.ID == req.WorkloadID {
			index = i
			targetID = item.ID
			break
		}
	}
	if index < 0 {
		return runtimeadapterpkg.ActionResult{}, fmt.Errorf("test workload %q not found", req.WorkloadID)
	}

	removed := m.workloads[index]
	m.workloads = append(m.workloads[:index], m.workloads[index+1:]...)

	pods := m.pods[:0]
	for _, item := range m.pods {
		if item.WorkloadID != removed.ID {
			pods = append(pods, item)
		}
	}
	m.pods = pods

	metrics := m.metrics[:0]
	for _, item := range m.metrics {
		if item.WorkloadID != removed.ID {
			metrics = append(metrics, item)
		}
	}
	m.metrics = metrics

	if req.DeleteNamespace {
		namespaces := m.namespaces[:0]
		for _, item := range m.namespaces {
			if item.Name != removed.Namespace {
				namespaces = append(namespaces, item)
			}
		}
		m.namespaces = namespaces
	}

	return runtimeadapterpkg.ActionResult{
		WorkloadID: targetID,
		Action:     "delete",
		Status:     "accepted",
		Message:    fmt.Sprintf("delete workload %s accepted by test runtime adapter", removed.Name),
		UpdatedAt:  time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (m *testRuntimeAdapter) updateWorkload(id string, action string, mutate func(item *runtimeadapterpkg.Workload)) (runtimeadapterpkg.ActionResult, bool) {
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

		return runtimeadapterpkg.ActionResult{
			WorkloadID:   id,
			Action:       action,
			Status:       "accepted",
			Message:      fmt.Sprintf("%s workload accepted by test runtime adapter", action),
			Desired:      m.workloads[index].Desired,
			Ready:        m.workloads[index].Ready,
			Available:    m.workloads[index].Available,
			AffectedPods: affected,
			UpdatedAt:    m.workloads[index].LastActionAt,
		}, true
	}

	return runtimeadapterpkg.ActionResult{}, false
}

func (m *testRuntimeAdapter) ExecuteDiagnosticCommand(req runtimeadapterpkg.DiagnosticExecRequest) (runtimeadapterpkg.DiagnosticExecResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	startedAt := time.Now().UTC()
	workload, workloadOK := m.findWorkloadLocked(req.WorkloadID)
	if !workloadOK {
		return runtimeadapterpkg.DiagnosticExecResult{}, fmt.Errorf("test workload %q not found", req.WorkloadID)
	}

	pod, podOK := m.findPodLocked(req.WorkloadID, req.PodName)
	if !podOK {
		return runtimeadapterpkg.DiagnosticExecResult{}, fmt.Errorf("test pod %q not found", req.PodName)
	}

	commandText := strings.TrimSpace(strings.Join(req.Command, " "))
	if commandText == "" {
		return runtimeadapterpkg.DiagnosticExecResult{}, fmt.Errorf("diagnostic command is empty")
	}

	stdout, stderr, exitCode := m.mockDiagnosticOutput(workload, pod, req.Command)
	finishedAt := startedAt.Add(12 * time.Millisecond)

	return runtimeadapterpkg.DiagnosticExecResult{
		WorkloadID:    workload.ID,
		Namespace:     workload.Namespace,
		PodName:       pod.Name,
		ContainerName: strings.TrimSpace(req.ContainerName),
		Command:       append([]string(nil), req.Command...),
		ExitCode:      exitCode,
		Stdout:        stdout,
		Stderr:        stderr,
		DurationMs:    int(finishedAt.Sub(startedAt).Milliseconds()),
		StartedAt:     startedAt.Format(time.RFC3339),
		FinishedAt:    finishedAt.Format(time.RFC3339),
	}, nil
}

func (m *testRuntimeAdapter) findWorkloadLocked(workloadID string) (runtimeadapterpkg.Workload, bool) {
	for _, item := range m.workloads {
		if item.ID == workloadID {
			return item, true
		}
	}
	return runtimeadapterpkg.Workload{}, false
}

func (m *testRuntimeAdapter) findPodLocked(workloadID string, podName string) (runtimeadapterpkg.Pod, bool) {
	for _, item := range m.pods {
		if item.WorkloadID != workloadID {
			continue
		}
		if strings.TrimSpace(podName) == "" || item.Name == podName {
			return item, true
		}
	}
	return runtimeadapterpkg.Pod{}, false
}

func (m *testRuntimeAdapter) mockDiagnosticOutput(workload runtimeadapterpkg.Workload, pod runtimeadapterpkg.Pod, command []string) (string, string, int) {
	signature := strings.Join(command, " ")
	switch signature {
	case "ps aux":
		return fmt.Sprintf(`USER       PID %%CPU %%MEM COMMAND
root         1  0.2  0.8 openclaw-server --instance=%s
root        28  0.1  0.3 sidecar-log-forwarder
root        63  2.4  1.6 node /srv/openclaw/runtime.js
root       105  0.0  0.1 tail -F /var/log/app.log
`, workload.Name), "", 0
	case "df -h":
		return `Filesystem      Size  Used Avail Use% Mounted on
overlay          40G   14G   24G  37% /
/dev/sda1        80G   31G   46G  41% /data
tmpfs           512M     0  512M   0% /dev/shm
`, "", 0
	case "free -m":
		return `               total        used        free      shared  buff/cache   available
Mem:            4096        2180         734         110        1181        1530
Swap:              0           0           0
`, "", 0
	case "printenv":
		return fmt.Sprintf(`HOSTNAME=%s
NAMESPACE=%s
WORKLOAD_ID=%s
OPENCLAW_MODE=production
OPENCLAW_INSTANCE=%s
`, pod.Name, workload.Namespace, workload.ID, workload.Name), "", 0
	case "ss -tunlp":
		return `Netid State  Recv-Q Send-Q Local Address:Port Peer Address:PortProcess
tcp   LISTEN 0      511          0.0.0.0:8080      0.0.0.0:*    users:(("openclaw",pid=1,fd=18))
tcp   ESTAB  0      0      10.0.12.44:8080    10.0.0.12:54022 users:(("openclaw",pid=1,fd=20))
`, "", 0
	case "uname -a":
		return "Linux " + pod.NodeName + " 6.6.32-openclaw #1 SMP PREEMPT_DYNAMIC x86_64 GNU/Linux\n", "", 0
	case "cat /etc/os-release":
		return `NAME="OpenClaw Runtime"
ID=openclaw
VERSION="2026.04"
PRETTY_NAME="OpenClaw Runtime 2026.04"
`, "", 0
	case "uptime":
		return " 10:32:18 up 18 days,  4:12,  load average: 0.31, 0.28, 0.24\n", "", 0
	default:
		return "", fmt.Sprintf("test runtime adapter does not support command %q", signature), 127
	}
}
