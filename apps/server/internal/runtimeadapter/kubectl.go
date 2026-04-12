package runtimeadapter

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	lastActionAnnotation       = "openclaw.io/last-action-at"
	previousReplicasAnnotation = "openclaw.io/previous-replicas"
	workloadIDLabel            = "openclaw.io/workload-id"
	defaultOpenClawImage       = "ghcr.io/openclaw/openclaw:latest"
	defaultServiceType         = "NodePort"
	defaultAccessHost          = "localhost"
	defaultAccessScheme        = "http"
	defaultOpenClawPort        = 18789
)

type kubectlRunner interface {
	Run(ctx context.Context, stdin string, args ...string) ([]byte, error)
}

type execKubectlRunner struct {
	binary  string
	context string
}

func (r execKubectlRunner) Run(ctx context.Context, stdin string, args ...string) ([]byte, error) {
	finalArgs := make([]string, 0, len(args)+2)
	if strings.TrimSpace(r.context) != "" {
		finalArgs = append(finalArgs, "--context", r.context)
	}
	finalArgs = append(finalArgs, args...)

	cmd := exec.CommandContext(ctx, r.binary, finalArgs...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("kubectl %s: %w: %s", strings.Join(finalArgs, " "), err, strings.TrimSpace(string(output)))
	}
	return output, nil
}

type KubectlAdapter struct {
	cfg    Config
	runner kubectlRunner
}

func NewKubectlAdapter(cfg Config) *KubectlAdapter {
	if strings.TrimSpace(cfg.KubectlBinary) == "" {
		cfg.KubectlBinary = "kubectl"
	}
	if strings.TrimSpace(cfg.Image) == "" {
		cfg.Image = defaultOpenClawImage
	}
	if strings.TrimSpace(cfg.ServiceType) == "" {
		cfg.ServiceType = defaultServiceType
	}
	if strings.TrimSpace(cfg.AccessHost) == "" {
		cfg.AccessHost = defaultAccessHost
	}
	if strings.TrimSpace(cfg.AccessScheme) == "" {
		cfg.AccessScheme = defaultAccessScheme
	}
	if cfg.Port <= 0 {
		cfg.Port = defaultOpenClawPort
	}

	return &KubectlAdapter{
		cfg: cfg,
		runner: execKubectlRunner{
			binary:  cfg.KubectlBinary,
			context: cfg.KubeContext,
		},
	}
}

func (k *KubectlAdapter) ListClusters() []Cluster {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	clusterID, err := k.currentContext(ctx)
	if err != nil || clusterID == "" {
		return nil
	}

	nodes := k.listNodesForContext(ctx, clusterID)
	workloads := k.listWorkloadsForContext(ctx, clusterID, "")
	version := k.serverVersion(ctx)

	status := "healthy"
	if len(nodes) == 0 {
		status = "warning"
	}

	return []Cluster{{
		ID:        clusterID,
		Code:      clusterID,
		Name:      clusterID,
		Region:    "local",
		Status:    status,
		Version:   version,
		Nodes:     len(nodes),
		Workloads: len(workloads),
	}}
}

func (k *KubectlAdapter) GetCluster(id string) (Cluster, bool) {
	for _, item := range k.ListClusters() {
		if item.ID == id {
			return item, true
		}
	}
	return Cluster{}, false
}

func (k *KubectlAdapter) ListNodes(clusterID string) []Node {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	currentContext, err := k.currentContext(ctx)
	if err != nil || !k.matchesCluster(clusterID, currentContext) {
		return nil
	}
	return k.listNodesForContext(ctx, currentContext)
}

func (k *KubectlAdapter) ListNamespaces(clusterID string) []Namespace {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	currentContext, err := k.currentContext(ctx)
	if err != nil || !k.matchesCluster(clusterID, currentContext) {
		return nil
	}

	var namespaces namespaceListResponse
	if err := k.getJSON(ctx, &namespaces, "get", "namespaces", "-o", "json"); err != nil {
		return nil
	}

	workloads := k.listWorkloadsForContext(ctx, currentContext, "")
	pods := k.listAllPods(ctx)
	workloadCount := make(map[string]int)
	podCount := make(map[string]int)
	for _, item := range workloads {
		workloadCount[item.Namespace]++
	}
	for _, item := range pods.Items {
		podCount[item.Metadata.Namespace]++
	}

	items := make([]Namespace, 0, len(namespaces.Items))
	for _, item := range namespaces.Items {
		items = append(items, Namespace{
			ID:        item.Metadata.Name,
			ClusterID: currentContext,
			Name:      item.Metadata.Name,
			Status:    namespacePhase(item.Status.Phase),
			Workloads: workloadCount[item.Metadata.Name],
			Pods:      podCount[item.Metadata.Name],
		})
	}
	return items
}

func (k *KubectlAdapter) ListWorkloads(clusterID string, namespace string) []Workload {
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	currentContext, err := k.currentContext(ctx)
	if err != nil || !k.matchesCluster(clusterID, currentContext) {
		return nil
	}
	return k.listWorkloadsForContext(ctx, currentContext, namespace)
}

func (k *KubectlAdapter) GetWorkload(id string) (Workload, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	item, ok := k.findDeployment(ctx, id)
	if !ok {
		return Workload{}, false
	}
	clusterID, _ := k.currentContext(ctx)
	return deploymentToWorkload(clusterID, item), true
}

func (k *KubectlAdapter) ListPods(workloadID string) []Pod {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	workload, selector, ok := k.findWorkloadSelector(ctx, workloadID)
	if !ok || selector == "" {
		return nil
	}

	pods := k.listPodsBySelector(ctx, workload.Namespace, selector)
	items := make([]Pod, 0, len(pods.Items))
	for _, item := range pods.Items {
		items = append(items, Pod{
			ID:         string(item.Metadata.UID),
			WorkloadID: workload.ID,
			Name:       item.Metadata.Name,
			NodeName:   item.Spec.NodeName,
			Status:     podPhase(item.Status.Phase),
			Restarts:   containerRestarts(item.Status.ContainerStatuses),
			StartedAt:  item.Status.StartTime,
		})
	}
	return items
}

func (k *KubectlAdapter) GetMetrics(workloadID string) (WorkloadMetrics, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	workload, selector, ok := k.findWorkloadSelector(ctx, workloadID)
	if !ok || selector == "" {
		return WorkloadMetrics{}, false
	}

	output, err := k.runner.Run(ctx, "", "top", "pods", "-n", workload.Namespace, "-l", selector, "--no-headers")
	if err != nil {
		return WorkloadMetrics{}, false
	}

	cpuMilli := 0
	memoryMB := 0
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		cpuMilli += parseMilliCPU(fields[1])
		memoryMB += parseMemoryMB(fields[2])
	}

	return WorkloadMetrics{
		WorkloadID:    workload.ID,
		CPUUsageMilli: cpuMilli,
		MemoryUsageMB: memoryMB,
	}, true
}

func (k *KubectlAdapter) StartWorkload(id string) (ActionResult, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	item, ok := k.findDeployment(ctx, id)
	if !ok {
		return ActionResult{}, false
	}

	desired := 1
	if raw := strings.TrimSpace(item.Metadata.Annotations[previousReplicasAnnotation]); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			desired = parsed
		}
	} else if item.Spec.Replicas != nil && *item.Spec.Replicas > 0 {
		desired = *item.Spec.Replicas
	}

	result, err := k.patchReplicas(ctx, item.Metadata.Namespace, item.Metadata.Name, desired, "start", map[string]string{
		previousReplicasAnnotation: strconv.Itoa(desired),
	})
	if err != nil {
		return ActionResult{}, false
	}
	return result, true
}

func (k *KubectlAdapter) StopWorkload(id string) (ActionResult, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	item, ok := k.findDeployment(ctx, id)
	if !ok {
		return ActionResult{}, false
	}

	desired := 1
	if item.Spec.Replicas != nil && *item.Spec.Replicas > 0 {
		desired = *item.Spec.Replicas
	}

	result, err := k.patchReplicas(ctx, item.Metadata.Namespace, item.Metadata.Name, 0, "stop", map[string]string{
		previousReplicasAnnotation: strconv.Itoa(desired),
	})
	if err != nil {
		return ActionResult{}, false
	}
	return result, true
}

func (k *KubectlAdapter) RestartWorkload(id string) (ActionResult, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	item, ok := k.findDeployment(ctx, id)
	if !ok {
		return ActionResult{}, false
	}

	if _, err := k.runner.Run(ctx, "", "rollout", "restart", "deployment/"+item.Metadata.Name, "-n", item.Metadata.Namespace); err != nil {
		return ActionResult{}, false
	}

	_, _ = k.patchAnnotations(ctx, item.Metadata.Namespace, item.Metadata.Name, map[string]string{
		lastActionAnnotation: time.Now().UTC().Format(time.RFC3339),
	})

	workload, found := k.GetWorkload(id)
	if !found {
		return ActionResult{}, false
	}

	return ActionResult{
		WorkloadID:   workload.ID,
		Action:       "restart",
		Status:       "accepted",
		Message:      "restart accepted by kubectl adapter",
		Desired:      workload.Desired,
		Ready:        workload.Ready,
		Available:    workload.Available,
		AffectedPods: podNames(k.ListPods(id)),
		UpdatedAt:    workload.LastActionAt,
	}, true
}

func (k *KubectlAdapter) ScaleWorkload(id string, replicas int) (ActionResult, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	item, ok := k.findDeployment(ctx, id)
	if !ok {
		return ActionResult{}, false
	}

	annotations := map[string]string{}
	if replicas == 0 {
		current := 1
		if item.Spec.Replicas != nil && *item.Spec.Replicas > 0 {
			current = *item.Spec.Replicas
		}
		annotations[previousReplicasAnnotation] = strconv.Itoa(current)
	}

	result, err := k.patchReplicas(ctx, item.Metadata.Namespace, item.Metadata.Name, replicas, "scale", annotations)
	if err != nil {
		return ActionResult{}, false
	}
	return result, true
}

func (k *KubectlAdapter) CreateWorkload(req CreateWorkloadRequest) (CreateWorkloadResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	clusterID, err := k.currentContext(ctx)
	if err != nil {
		return CreateWorkloadResult{}, err
	}

	replicas := req.Replicas
	if replicas <= 0 {
		replicas = 1
	}

	labels := map[string]string{
		workloadIDLabel:                req.WorkloadID,
		"app.kubernetes.io/name":       "openclaw",
		"app.kubernetes.io/instance":   req.Name,
		"app.kubernetes.io/managed-by": "openclaw-platform",
	}
	for key, value := range req.Labels {
		labels[key] = value
	}

	annotations := map[string]string{
		lastActionAnnotation: time.Now().UTC().Format(time.RFC3339),
	}
	for key, value := range req.Annotations {
		annotations[key] = value
	}

	items := []map[string]any{
		{
			"apiVersion": "v1",
			"kind":       "Namespace",
			"metadata": map[string]any{
				"name": req.Namespace,
				"labels": map[string]string{
					"app.kubernetes.io/managed-by": "openclaw-platform",
				},
			},
		},
		{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]any{
				"name":        req.Name,
				"namespace":   req.Namespace,
				"labels":      labels,
				"annotations": annotations,
			},
			"spec": map[string]any{
				"replicas": replicas,
				"selector": map[string]any{
					"matchLabels": map[string]string{
						workloadIDLabel: req.WorkloadID,
					},
				},
				"template": map[string]any{
					"metadata": map[string]any{
						"labels":      labels,
						"annotations": annotations,
					},
					"spec": map[string]any{
						"containers": []map[string]any{
							{
								"name":  "openclaw",
								"image": defaultString(req.Image, k.cfg.Image),
								"ports": []map[string]any{
									{
										"name":          "http",
										"containerPort": defaultInt(req.Port, k.cfg.Port),
									},
								},
								"env": buildEnvVars(req.Env),
								"resources": map[string]any{
									"requests": map[string]string{
										"cpu":    defaultString(req.CPU, "1"),
										"memory": defaultString(req.Memory, "2Gi"),
									},
									"limits": map[string]string{
										"cpu":    defaultString(req.CPU, "1"),
										"memory": defaultString(req.Memory, "2Gi"),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			"apiVersion": "v1",
			"kind":       "Service",
			"metadata": map[string]any{
				"name":      req.Name,
				"namespace": req.Namespace,
				"labels":    labels,
			},
			"spec": map[string]any{
				"type": k.cfg.ServiceType,
				"selector": map[string]string{
					workloadIDLabel: req.WorkloadID,
				},
				"ports": []map[string]any{
					{
						"name":       "http",
						"port":       80,
						"targetPort": defaultInt(req.Port, k.cfg.Port),
					},
				},
			},
		},
	}

	manifest, err := json.Marshal(map[string]any{
		"apiVersion": "v1",
		"kind":       "List",
		"items":      items,
	})
	if err != nil {
		return CreateWorkloadResult{}, err
	}

	if _, err := k.runner.Run(ctx, string(manifest), "apply", "-f", "-"); err != nil {
		return CreateWorkloadResult{}, err
	}

	workload, found := k.GetWorkload(req.WorkloadID)
	if !found {
		return CreateWorkloadResult{}, fmt.Errorf("created workload %q not found", req.WorkloadID)
	}
	workload.ClusterID = clusterID

	return CreateWorkloadResult{
		Workload:        workload,
		AccessEndpoints: k.GetWorkloadAccess(req.WorkloadID),
	}, nil
}

func (k *KubectlAdapter) DeleteWorkload(req DeleteWorkloadRequest) (ActionResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	workload, found := k.GetWorkload(req.WorkloadID)
	if !found && strings.TrimSpace(req.Namespace) != "" && strings.TrimSpace(req.Name) != "" {
		if item, ok := k.getDeployment(ctx, req.Namespace, req.Name); ok {
			clusterID, _ := k.currentContext(ctx)
			resolved := deploymentToWorkload(clusterID, item)
			workload = resolved
			found = true
		}
	}
	if !found {
		return ActionResult{}, fmt.Errorf("workload %q not found", req.WorkloadID)
	}

	if req.DeleteNamespace && strings.TrimSpace(req.Namespace) != "" {
		if _, err := k.runner.Run(ctx, "", "delete", "namespace", req.Namespace, "--wait=false"); err != nil {
			return ActionResult{}, err
		}
	} else {
		name := defaultString(req.Name, workload.Name)
		namespace := defaultString(req.Namespace, workload.Namespace)
		if _, err := k.runner.Run(ctx, "", "delete", "service", name, "-n", namespace, "--ignore-not-found=true"); err != nil {
			return ActionResult{}, err
		}
		if _, err := k.runner.Run(ctx, "", "delete", "deployment", name, "-n", namespace, "--ignore-not-found=true"); err != nil {
			return ActionResult{}, err
		}
	}

	return ActionResult{
		WorkloadID: workload.ID,
		Action:     "delete",
		Status:     "accepted",
		Message:    fmt.Sprintf("delete accepted by kubectl adapter for %s", workload.Name),
		Desired:    0,
		Ready:      0,
		Available:  0,
		UpdatedAt:  time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (k *KubectlAdapter) GetWorkloadAccess(id string) []AccessEndpoint {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	item, ok := k.findDeployment(ctx, id)
	if !ok {
		return nil
	}

	service, ok := k.getService(ctx, item.Metadata.Namespace, item.Metadata.Name)
	if !ok {
		return nil
	}

	workloadID := deploymentID(item)
	endpoints := []AccessEndpoint{
		{
			WorkloadID:  workloadID,
			ServiceName: service.Metadata.Name,
			Namespace:   service.Metadata.Namespace,
			Type:        "clusterIP",
			URL:         fmt.Sprintf("%s://%s.%s.svc.cluster.local:%d", k.cfg.AccessScheme, service.Metadata.Name, service.Metadata.Namespace, firstServicePort(service)),
			Host:        fmt.Sprintf("%s.%s.svc.cluster.local", service.Metadata.Name, service.Metadata.Namespace),
			Port:        firstServicePort(service),
		},
	}

	if port := firstNodePort(service); port > 0 {
		endpoints = append([]AccessEndpoint{{
			WorkloadID:  workloadID,
			ServiceName: service.Metadata.Name,
			Namespace:   service.Metadata.Namespace,
			Type:        "nodePort",
			URL:         fmt.Sprintf("%s://%s:%d", k.cfg.AccessScheme, k.cfg.AccessHost, port),
			Host:        k.cfg.AccessHost,
			Port:        port,
		}}, endpoints...)
	}

	return endpoints
}

func (k *KubectlAdapter) ExecuteDiagnosticCommand(req DiagnosticExecRequest) (DiagnosticExecResult, error) {
	if strings.TrimSpace(req.Namespace) == "" {
		return DiagnosticExecResult{}, fmt.Errorf("diagnostic namespace is required")
	}
	if strings.TrimSpace(req.PodName) == "" {
		return DiagnosticExecResult{}, fmt.Errorf("diagnostic pod name is required")
	}
	if len(req.Command) == 0 {
		return DiagnosticExecResult{}, fmt.Errorf("diagnostic command is required")
	}

	timeout := req.TimeoutSeconds
	if timeout <= 0 {
		timeout = 8
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	startedAt := time.Now().UTC()
	args := make([]string, 0, len(req.Command)+8)
	if strings.TrimSpace(k.cfg.KubeContext) != "" {
		args = append(args, "--context", k.cfg.KubeContext)
	}
	args = append(args, "exec", "-n", req.Namespace, req.PodName)
	if strings.TrimSpace(req.ContainerName) != "" {
		args = append(args, "-c", req.ContainerName)
	}
	args = append(args, "--")
	args = append(args, req.Command...)

	cmd := exec.CommandContext(ctx, k.cfg.KubectlBinary, args...)
	output, err := cmd.CombinedOutput()
	finishedAt := time.Now().UTC()

	result := DiagnosticExecResult{
		WorkloadID:    req.WorkloadID,
		Namespace:     req.Namespace,
		PodName:       req.PodName,
		ContainerName: strings.TrimSpace(req.ContainerName),
		Command:       append([]string(nil), req.Command...),
		DurationMs:    int(finishedAt.Sub(startedAt).Milliseconds()),
		StartedAt:     startedAt.Format(time.RFC3339),
		FinishedAt:    finishedAt.Format(time.RFC3339),
	}

	if err == nil {
		result.Stdout = string(output)
		return result, nil
	}

	if ctx.Err() == context.DeadlineExceeded {
		result.ExitCode = 124
		result.Stderr = "diagnostic command timed out"
		return result, fmt.Errorf("diagnostic command timed out")
	}

	result.Stderr = strings.TrimSpace(string(output))
	if exitErr, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitErr.ExitCode()
	} else {
		result.ExitCode = 1
	}
	return result, fmt.Errorf("kubectl exec failed: %w", err)
}

func (k *KubectlAdapter) currentContext(ctx context.Context) (string, error) {
	output, err := k.runner.Run(ctx, "", "config", "current-context")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func (k *KubectlAdapter) serverVersion(ctx context.Context) string {
	var response versionResponse
	if err := k.getJSON(ctx, &response, "version", "-o", "json"); err != nil {
		return ""
	}
	return response.ServerVersion.GitVersion
}

func (k *KubectlAdapter) listNodesForContext(ctx context.Context, clusterID string) []Node {
	var nodes nodeListResponse
	if err := k.getJSON(ctx, &nodes, "get", "nodes", "-o", "json"); err != nil {
		return nil
	}

	pods := k.listAllPods(ctx)
	podCount := make(map[string]int)
	for _, item := range pods.Items {
		if item.Spec.NodeName == "" {
			continue
		}
		podCount[item.Spec.NodeName]++
	}

	items := make([]Node, 0, len(nodes.Items))
	for _, item := range nodes.Items {
		items = append(items, Node{
			ID:             item.Metadata.Name,
			ClusterID:      clusterID,
			Name:           item.Metadata.Name,
			Role:           nodeRole(item.Metadata.Labels),
			Status:         nodeStatus(item.Status.Conditions),
			Kubelet:        item.Status.NodeInfo.KubeletVersion,
			CPUCapacity:    item.Status.Capacity.CPU,
			MemoryCapacity: item.Status.Capacity.Memory,
			PodCount:       podCount[item.Metadata.Name],
		})
	}
	return items
}

func (k *KubectlAdapter) listWorkloadsForContext(ctx context.Context, clusterID string, namespace string) []Workload {
	deployments, err := k.listDeployments(ctx, namespace)
	if err != nil {
		return nil
	}

	items := make([]Workload, 0, len(deployments.Items))
	for _, item := range deployments.Items {
		items = append(items, deploymentToWorkload(clusterID, item))
	}
	return items
}

func (k *KubectlAdapter) listDeployments(ctx context.Context, namespace string) (deploymentListResponse, error) {
	args := []string{"get", "deployments"}
	if strings.TrimSpace(namespace) == "" {
		args = append(args, "-A")
	} else {
		args = append(args, "-n", namespace)
	}
	args = append(args, "-o", "json")

	var response deploymentListResponse
	if err := k.getJSON(ctx, &response, args...); err != nil {
		return deploymentListResponse{}, err
	}
	return response, nil
}

func (k *KubectlAdapter) findDeployment(ctx context.Context, workloadID string) (deploymentResponseItem, bool) {
	deployments, err := k.listDeployments(ctx, "")
	if err != nil {
		return deploymentResponseItem{}, false
	}
	for _, item := range deployments.Items {
		if deploymentID(item) == workloadID {
			return item, true
		}
	}
	return deploymentResponseItem{}, false
}

func (k *KubectlAdapter) findWorkloadSelector(ctx context.Context, workloadID string) (Workload, string, bool) {
	item, ok := k.findDeployment(ctx, workloadID)
	if !ok {
		return Workload{}, "", false
	}
	clusterID, _ := k.currentContext(ctx)
	return deploymentToWorkload(clusterID, item), labelSelector(item.Spec.Selector.MatchLabels), true
}

func (k *KubectlAdapter) listPodsBySelector(ctx context.Context, namespace string, selector string) podListResponse {
	args := []string{"get", "pods", "-n", namespace}
	if selector != "" {
		args = append(args, "-l", selector)
	}
	args = append(args, "-o", "json")

	var response podListResponse
	if err := k.getJSON(ctx, &response, args...); err != nil {
		return podListResponse{}
	}
	return response
}

func (k *KubectlAdapter) listAllPods(ctx context.Context) podListResponse {
	var response podListResponse
	_ = k.getJSON(ctx, &response, "get", "pods", "-A", "-o", "json")
	return response
}

func (k *KubectlAdapter) getService(ctx context.Context, namespace string, name string) (serviceResponse, bool) {
	var response serviceResponse
	err := k.getJSON(ctx, &response, "get", "service", name, "-n", namespace, "-o", "json")
	if err != nil {
		return serviceResponse{}, false
	}
	return response, true
}

func (k *KubectlAdapter) getDeployment(ctx context.Context, namespace string, name string) (deploymentResponseItem, bool) {
	var response deploymentResponseItem
	err := k.getJSON(ctx, &response, "get", "deployment", name, "-n", namespace, "-o", "json")
	if err != nil {
		return deploymentResponseItem{}, false
	}
	return response, true
}

func (k *KubectlAdapter) patchReplicas(ctx context.Context, namespace string, name string, replicas int, action string, annotations map[string]string) (ActionResult, error) {
	annotations[lastActionAnnotation] = time.Now().UTC().Format(time.RFC3339)
	if err := k.scaleDeployment(ctx, namespace, name, replicas); err != nil {
		return ActionResult{}, err
	}
	if err := k.annotateDeployment(ctx, namespace, name, annotations); err != nil {
		return ActionResult{}, err
	}

	item, ok := k.getDeployment(ctx, namespace, name)
	if !ok {
		return ActionResult{}, fmt.Errorf("workload %s/%s not found after patch", namespace, name)
	}
	clusterID, _ := k.currentContext(ctx)
	workload := deploymentToWorkload(clusterID, item)

	return ActionResult{
		WorkloadID:   workload.ID,
		Action:       action,
		Status:       "accepted",
		Message:      fmt.Sprintf("%s accepted by kubectl adapter", action),
		Desired:      workload.Desired,
		Ready:        workload.Ready,
		Available:    workload.Available,
		AffectedPods: podNames(k.ListPods(workload.ID)),
		UpdatedAt:    workload.LastActionAt,
	}, nil
}

func (k *KubectlAdapter) patchAnnotations(ctx context.Context, namespace string, name string, annotations map[string]string) (deploymentResponseItem, error) {
	if err := k.annotateDeployment(ctx, namespace, name, annotations); err != nil {
		return deploymentResponseItem{}, err
	}
	item, ok := k.getDeployment(ctx, namespace, name)
	if !ok {
		return deploymentResponseItem{}, fmt.Errorf("patched workload %s/%s not found", namespace, name)
	}
	return item, nil
}

func (k *KubectlAdapter) scaleDeployment(ctx context.Context, namespace string, name string, replicas int) error {
	_, err := k.runner.Run(ctx, "", "scale", "deployment", name, "-n", namespace, fmt.Sprintf("--replicas=%d", replicas))
	return err
}

func (k *KubectlAdapter) annotateDeployment(ctx context.Context, namespace string, name string, annotations map[string]string) error {
	if len(annotations) == 0 {
		return nil
	}

	args := []string{"annotate", "deployment", name, "-n", namespace, "--overwrite"}
	keys := make([]string, 0, len(annotations))
	for key := range annotations {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		args = append(args, fmt.Sprintf("%s=%s", key, annotations[key]))
	}

	_, err := k.runner.Run(ctx, "", args...)
	return err
}

func (k *KubectlAdapter) getJSON(ctx context.Context, target any, args ...string) error {
	output, err := k.runner.Run(ctx, "", args...)
	if err != nil {
		return err
	}
	return json.Unmarshal(output, target)
}

func (k *KubectlAdapter) matchesCluster(requested string, current string) bool {
	return requested == "" || requested == current
}

func deploymentToWorkload(clusterID string, item deploymentResponseItem) Workload {
	desired := 1
	if item.Spec.Replicas != nil {
		desired = *item.Spec.Replicas
	}
	image := ""
	if len(item.Spec.Template.Spec.Containers) > 0 {
		image = item.Spec.Template.Spec.Containers[0].Image
	}

	lastActionAt := item.Metadata.Annotations[lastActionAnnotation]
	if strings.TrimSpace(lastActionAt) == "" {
		lastActionAt = item.Metadata.CreationTimestamp
	}

	return Workload{
		ID:           deploymentID(item),
		ClusterID:    clusterID,
		Namespace:    item.Metadata.Namespace,
		Name:         item.Metadata.Name,
		Kind:         "Deployment",
		Image:        image,
		Status:       deploymentStatus(item, desired),
		Desired:      desired,
		Ready:        item.Status.ReadyReplicas,
		Available:    item.Status.AvailableReplicas,
		LastActionAt: lastActionAt,
	}
}

func deploymentID(item deploymentResponseItem) string {
	if id := strings.TrimSpace(item.Metadata.Labels[workloadIDLabel]); id != "" {
		return id
	}
	return fmt.Sprintf("%s:%s", item.Metadata.Namespace, item.Metadata.Name)
}

func deploymentStatus(item deploymentResponseItem, desired int) string {
	switch {
	case desired == 0:
		return "Stopped"
	case item.Status.AvailableReplicas >= desired && desired > 0:
		return "Running"
	case item.Status.ReadyReplicas > 0:
		return "Degraded"
	default:
		return "Pending"
	}
}

func nodeRole(labels map[string]string) string {
	switch {
	case hasLabelPrefix(labels, "node-role.kubernetes.io/control-plane"), hasLabelPrefix(labels, "node-role.kubernetes.io/master"):
		return "control-plane"
	default:
		return "worker"
	}
}

func nodeStatus(conditions []nodeCondition) string {
	status := "Unknown"
	for _, condition := range conditions {
		if condition.Type == "Ready" {
			if condition.Status == "True" {
				status = "Ready"
			} else {
				status = "NotReady"
			}
		}
		if strings.HasSuffix(condition.Type, "Pressure") && condition.Status == "True" {
			return condition.Type
		}
	}
	return status
}

func namespacePhase(phase string) string {
	if strings.TrimSpace(phase) == "" {
		return "Unknown"
	}
	return phase
}

func podPhase(phase string) string {
	if strings.TrimSpace(phase) == "" {
		return "Unknown"
	}
	return phase
}

func containerRestarts(items []containerStatus) int {
	total := 0
	for _, item := range items {
		total += item.RestartCount
	}
	return total
}

func labelSelector(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}

	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", key, labels[key]))
	}
	return strings.Join(parts, ",")
}

func hasLabelPrefix(labels map[string]string, prefix string) bool {
	for key := range labels {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return false
}

func firstNodePort(service serviceResponse) int {
	for _, item := range service.Spec.Ports {
		if item.NodePort > 0 {
			return item.NodePort
		}
	}
	return 0
}

func firstServicePort(service serviceResponse) int {
	for _, item := range service.Spec.Ports {
		if item.Port > 0 {
			return item.Port
		}
	}
	return 0
}

func podNames(items []Pod) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.Name)
	}
	return out
}

func buildEnvVars(values map[string]string) []map[string]string {
	merged := map[string]string{
		"NODE_ENV": "production",
	}
	for key, value := range values {
		merged[key] = value
	}

	keys := make([]string, 0, len(merged))
	for key := range merged {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	items := make([]map[string]string, 0, len(keys))
	for _, key := range keys {
		items = append(items, map[string]string{
			"name":  key,
			"value": merged[key],
		})
	}
	return items
}

func parseMilliCPU(value string) int {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	if strings.HasSuffix(value, "m") {
		parsed, _ := strconv.Atoi(strings.TrimSuffix(value, "m"))
		return parsed
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return int(parsed * 1000)
}

func parseMemoryMB(value string) int {
	value = strings.TrimSpace(strings.ToUpper(value))
	if value == "" {
		return 0
	}

	units := map[string]float64{
		"KI": 1.0 / 1024.0,
		"MI": 1,
		"GI": 1024,
		"TI": 1024 * 1024,
		"K":  1.0 / 1000.0,
		"M":  1,
		"G":  1000,
	}
	for suffix, factor := range units {
		if !strings.HasSuffix(value, suffix) {
			continue
		}
		raw := strings.TrimSuffix(value, suffix)
		parsed, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return 0
		}
		return int(parsed * factor)
	}

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return int(parsed)
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func defaultInt(value int, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}

type versionResponse struct {
	ServerVersion struct {
		GitVersion string `json:"gitVersion"`
	} `json:"serverVersion"`
}

type metadataResponse struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	UID               string            `json:"uid"`
	Labels            map[string]string `json:"labels"`
	Annotations       map[string]string `json:"annotations"`
	CreationTimestamp string            `json:"creationTimestamp"`
}

type nodeListResponse struct {
	Items []nodeResponseItem `json:"items"`
}

type nodeResponseItem struct {
	Metadata metadataResponse `json:"metadata"`
	Status   struct {
		Capacity struct {
			CPU    string `json:"cpu"`
			Memory string `json:"memory"`
		} `json:"capacity"`
		NodeInfo struct {
			KubeletVersion string `json:"kubeletVersion"`
		} `json:"nodeInfo"`
		Conditions []nodeCondition `json:"conditions"`
	} `json:"status"`
}

type nodeCondition struct {
	Type   string `json:"type"`
	Status string `json:"status"`
}

type namespaceListResponse struct {
	Items []struct {
		Metadata metadataResponse `json:"metadata"`
		Status   struct {
			Phase string `json:"phase"`
		} `json:"status"`
	} `json:"items"`
}

type deploymentListResponse struct {
	Items []deploymentResponseItem `json:"items"`
}

type deploymentResponseItem struct {
	Metadata metadataResponse `json:"metadata"`
	Spec     struct {
		Replicas *int `json:"replicas"`
		Selector struct {
			MatchLabels map[string]string `json:"matchLabels"`
		} `json:"selector"`
		Template struct {
			Spec struct {
				Containers []struct {
					Image string `json:"image"`
				} `json:"containers"`
			} `json:"spec"`
		} `json:"template"`
	} `json:"spec"`
	Status struct {
		ReadyReplicas     int `json:"readyReplicas"`
		AvailableReplicas int `json:"availableReplicas"`
	} `json:"status"`
}

type podListResponse struct {
	Items []podResponseItem `json:"items"`
}

type podResponseItem struct {
	Metadata metadataResponse `json:"metadata"`
	Spec     struct {
		NodeName string `json:"nodeName"`
	} `json:"spec"`
	Status struct {
		Phase             string            `json:"phase"`
		StartTime         string            `json:"startTime"`
		ContainerStatuses []containerStatus `json:"containerStatuses"`
	} `json:"status"`
}

type containerStatus struct {
	RestartCount int `json:"restartCount"`
}

type serviceResponse struct {
	Metadata metadataResponse `json:"metadata"`
	Spec     struct {
		Ports []struct {
			Port     int `json:"port"`
			NodePort int `json:"nodePort"`
		} `json:"ports"`
	} `json:"spec"`
}
