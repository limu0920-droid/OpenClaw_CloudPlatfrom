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

type ProvisioningAdapter interface {
	CreateWorkload(req CreateWorkloadRequest) (CreateWorkloadResult, error)
}

type DeletionAdapter interface {
	DeleteWorkload(req DeleteWorkloadRequest) (ActionResult, error)
}

type AccessInfoProvider interface {
	GetWorkloadAccess(id string) []AccessEndpoint
}

type DiagnosticExecProvider interface {
	ExecuteDiagnosticCommand(req DiagnosticExecRequest) (DiagnosticExecResult, error)
}

type Config struct {
	StrictMode      bool
	Provider        string
	KubectlBinary   string
	KubeContext     string
	NamespacePrefix string
	WorkloadPrefix  string
	Image           string
	ServiceType     string
	AccessHost      string
	AccessScheme    string
	Port            int
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
	ID             string `json:"id"`
	ClusterID      string `json:"clusterId"`
	Name           string `json:"name"`
	Role           string `json:"role"`
	Status         string `json:"status"`
	Kubelet        string `json:"kubelet"`
	CPUCapacity    string `json:"cpuCapacity"`
	MemoryCapacity string `json:"memoryCapacity"`
	PodCount       int    `json:"podCount"`
}

type Namespace struct {
	ID        string `json:"id"`
	ClusterID string `json:"clusterId"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Workloads int    `json:"workloads"`
	Pods      int    `json:"pods"`
}

type Workload struct {
	ID           string `json:"id"`
	ClusterID    string `json:"clusterId"`
	Namespace    string `json:"namespace"`
	Name         string `json:"name"`
	Kind         string `json:"kind"`
	Image        string `json:"image"`
	Status       string `json:"status"`
	Desired      int    `json:"desired"`
	Ready        int    `json:"ready"`
	Available    int    `json:"available"`
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

type CreateWorkloadRequest struct {
	WorkloadID  string            `json:"workloadId"`
	Namespace   string            `json:"namespace"`
	Name        string            `json:"name"`
	Image       string            `json:"image"`
	Replicas    int               `json:"replicas"`
	Port        int               `json:"port"`
	CPU         string            `json:"cpu"`
	Memory      string            `json:"memory"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
}

type CreateWorkloadResult struct {
	Workload        Workload         `json:"workload"`
	AccessEndpoints []AccessEndpoint `json:"accessEndpoints,omitempty"`
}

type AccessEndpoint struct {
	WorkloadID  string `json:"workloadId"`
	ServiceName string `json:"serviceName"`
	Namespace   string `json:"namespace"`
	Type        string `json:"type"`
	URL         string `json:"url"`
	Host        string `json:"host,omitempty"`
	Port        int    `json:"port,omitempty"`
}

type DeleteWorkloadRequest struct {
	WorkloadID      string `json:"workloadId"`
	Namespace       string `json:"namespace"`
	Name            string `json:"name"`
	DeleteNamespace bool   `json:"deleteNamespace"`
}

type DiagnosticExecRequest struct {
	WorkloadID     string   `json:"workloadId"`
	Namespace      string   `json:"namespace"`
	PodName        string   `json:"podName"`
	ContainerName  string   `json:"containerName"`
	Command        []string `json:"command"`
	TimeoutSeconds int      `json:"timeoutSeconds"`
}

type DiagnosticExecResult struct {
	WorkloadID      string   `json:"workloadId"`
	Namespace       string   `json:"namespace"`
	PodName         string   `json:"podName"`
	ContainerName   string   `json:"containerName,omitempty"`
	Command         []string `json:"command"`
	ExitCode        int      `json:"exitCode"`
	Stdout          string   `json:"stdout,omitempty"`
	Stderr          string   `json:"stderr,omitempty"`
	DurationMs      int      `json:"durationMs"`
	StartedAt       string   `json:"startedAt"`
	FinishedAt      string   `json:"finishedAt"`
	OutputTruncated bool     `json:"outputTruncated,omitempty"`
}
