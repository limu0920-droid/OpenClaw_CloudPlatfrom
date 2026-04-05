package runtimeadapter

type Adapter interface {
	ListClusters() []Cluster
	GetCluster(id string) (Cluster, bool)
	ListNodes(clusterID string) []Node
	ListNamespaces(clusterID string) []Namespace
	ListWorkloads(clusterID string, namespace string) []Workload
	GetWorkload(id string) (Workload, bool)
	ListPods(workloadID string) []Pod
	GetMetrics(workloadID string) (WorkloadMetrics, bool)
	StartWorkload(id string) (ActionResult, bool)
	StopWorkload(id string) (ActionResult, bool)
	RestartWorkload(id string) (ActionResult, bool)
	ScaleWorkload(id string, replicas int) (ActionResult, bool)
}

type Cluster struct {
	ID        string `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	Region    string `json:"region"`
	Status    string `json:"status"`
	Version   string `json:"version"`
	Nodes     int    `json:"nodes"`
	Workloads int    `json:"workloads"`
}

type Node struct {
	ID            string `json:"id"`
	ClusterID     string `json:"clusterId"`
	Name          string `json:"name"`
	Role          string `json:"role"`
	Status        string `json:"status"`
	Kubelet       string `json:"kubelet"`
	CPUCapacity   string `json:"cpuCapacity"`
	MemoryCapacity string `json:"memoryCapacity"`
	PodCount      int    `json:"podCount"`
}

type Namespace struct {
	ID         string `json:"id"`
	ClusterID  string `json:"clusterId"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Workloads  int    `json:"workloads"`
	Pods       int    `json:"pods"`
}

type Workload struct {
	ID          string `json:"id"`
	ClusterID   string `json:"clusterId"`
	Namespace   string `json:"namespace"`
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	Image       string `json:"image"`
	Status      string `json:"status"`
	Desired     int    `json:"desired"`
	Ready       int    `json:"ready"`
	Available   int    `json:"available"`
	LastActionAt string `json:"lastActionAt"`
}

type Pod struct {
	ID         string `json:"id"`
	WorkloadID string `json:"workloadId"`
	Name       string `json:"name"`
	NodeName   string `json:"nodeName"`
	Status     string `json:"status"`
	Restarts   int    `json:"restarts"`
	StartedAt  string `json:"startedAt"`
}

type WorkloadMetrics struct {
	WorkloadID        string `json:"workloadId"`
	CPUUsageMilli     int    `json:"cpuUsageMilli"`
	MemoryUsageMB     int    `json:"memoryUsageMB"`
	NetworkRxKB       int    `json:"networkRxKB"`
	NetworkTxKB       int    `json:"networkTxKB"`
	ErrorRatePercent  int    `json:"errorRatePercent"`
	RequestsPerMinute int    `json:"requestsPerMinute"`
}

type ActionResult struct {
	WorkloadID   string   `json:"workloadId"`
	Action       string   `json:"action"`
	Status       string   `json:"status"`
	Message      string   `json:"message"`
	Desired      int      `json:"desired"`
	Ready        int      `json:"ready"`
	Available    int      `json:"available"`
	AffectedPods []string `json:"affectedPods,omitempty"`
	UpdatedAt    string   `json:"updatedAt"`
}
