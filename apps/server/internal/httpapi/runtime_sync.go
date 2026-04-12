package httpapi

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"openclaw/platformapi/internal/models"
	"openclaw/platformapi/internal/runtimeadapter"
)

type liveInstanceState struct {
	Instance models.Instance
	Binding  *models.RuntimeBinding
	Runtime  *models.InstanceRuntime
	Workload *runtimeadapter.Workload
	Metrics  *runtimeadapter.WorkloadMetrics
}

func buildInstanceRuntime(instanceID int, spec map[string]string, workload runtimeadapter.Workload, metrics *runtimeadapter.WorkloadMetrics) models.InstanceRuntime {
	runtime := models.InstanceRuntime{
		InstanceID: instanceID,
		PowerState: powerStateFromWorkload(workload),
		LastSeenAt: defaultString(workload.LastActionAt, time.Now().UTC().Format(time.RFC3339)),
	}
	if metrics == nil {
		return runtime
	}

	runtime.APIRequests24h = metrics.RequestsPerMinute * 60 * 24
	runtime.APITokens24h = metrics.RequestsPerMinute * 2048
	runtime.CPUUsagePercent = boundedPercent(metrics.CPUUsageMilli, parseCPUCapacityMilli(spec["cpu"]))
	runtime.MemoryUsagePercent = boundedPercent(metrics.MemoryUsageMB, parseMemoryCapacityMB(spec["memory"]))
	return runtime
}

func boundedPercent(used int, capacity int) int {
	if used <= 0 || capacity <= 0 {
		return 0
	}
	value := int(math.Round(float64(used) * 100 / float64(capacity)))
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}

func parseCPUCapacityMilli(value string) int {
	value = strings.TrimSpace(strings.ToLower(value))
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

func parseMemoryCapacityMB(value string) int {
	value = strings.TrimSpace(strings.ToUpper(value))
	if value == "" {
		return 0
	}
	units := map[string]float64{
		"KI": 1.0 / 1024.0,
		"MI": 1,
		"GI": 1024,
		"TI": 1024 * 1024,
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

func instanceStatusFromWorkload(workload runtimeadapter.Workload) string {
	switch strings.ToLower(strings.TrimSpace(workload.Status)) {
	case "running":
		return "running"
	case "stopped":
		return "stopped"
	case "pending":
		return "provisioning"
	case "degraded":
		return "updating"
	default:
		return strings.ToLower(strings.TrimSpace(workload.Status))
	}
}

func powerStateFromWorkload(workload runtimeadapter.Workload) string {
	switch strings.ToLower(strings.TrimSpace(workload.Status)) {
	case "running":
		return "running"
	case "stopped":
		return "stopped"
	case "pending":
		return "starting"
	default:
		return strings.ToLower(strings.TrimSpace(workload.Status))
	}
}

func (r *Router) upsertRuntimeLocked(runtime models.InstanceRuntime) {
	for index, item := range r.data.Runtimes {
		if item.InstanceID == runtime.InstanceID {
			r.data.Runtimes[index] = runtime
			return
		}
	}
	r.data.Runtimes = append([]models.InstanceRuntime{runtime}, r.data.Runtimes...)
}

func (r *Router) upsertCredentialLocked(credential models.InstanceCredential) {
	for index, item := range r.data.Credentials {
		if item.InstanceID == credential.InstanceID {
			r.data.Credentials[index] = credential
			return
		}
	}
	r.data.Credentials = append([]models.InstanceCredential{credential}, r.data.Credentials...)
}

func (r *Router) upsertRuntimeBindingLocked(binding models.RuntimeBinding) {
	for index, item := range r.data.RuntimeBindings {
		if item.InstanceID == binding.InstanceID {
			r.data.RuntimeBindings[index] = binding
			return
		}
	}
	r.data.RuntimeBindings = append([]models.RuntimeBinding{binding}, r.data.RuntimeBindings...)
}

func (r *Router) replaceAccessesLocked(instanceID int, items []models.InstanceAccess) {
	filtered := make([]models.InstanceAccess, 0, len(r.data.Accesses))
	for _, item := range r.data.Accesses {
		if item.InstanceID != instanceID {
			filtered = append(filtered, item)
		}
	}
	r.data.Accesses = append(items, filtered...)
}

func (r *Router) removeRuntimeBindingLocked(instanceID int) {
	filtered := r.data.RuntimeBindings[:0]
	for _, item := range r.data.RuntimeBindings {
		if item.InstanceID != instanceID {
			filtered = append(filtered, item)
		}
	}
	r.data.RuntimeBindings = filtered
}

func (r *Router) removeRuntimeLocked(instanceID int) {
	filtered := r.data.Runtimes[:0]
	for _, item := range r.data.Runtimes {
		if item.InstanceID != instanceID {
			filtered = append(filtered, item)
		}
	}
	r.data.Runtimes = filtered
}

func (r *Router) removeCredentialLocked(instanceID int) {
	filtered := r.data.Credentials[:0]
	for _, item := range r.data.Credentials {
		if item.InstanceID != instanceID {
			filtered = append(filtered, item)
		}
	}
	r.data.Credentials = filtered
}

func (r *Router) removeAccessesLocked(instanceID int) {
	filtered := r.data.Accesses[:0]
	for _, item := range r.data.Accesses {
		if item.InstanceID != instanceID {
			filtered = append(filtered, item)
		}
	}
	r.data.Accesses = filtered
}

func buildAccessesFromRuntime(instanceID int, endpoints []runtimeadapter.AccessEndpoint) []models.InstanceAccess {
	if len(endpoints) == 0 {
		return nil
	}

	primary := endpoints[0]
	domain := primary.Host
	items := []models.InstanceAccess{{
		InstanceID: instanceID,
		EntryType:  "web",
		URL:        primary.URL,
		Domain:     domain,
		AccessMode: "direct",
		IsPrimary:  true,
	}}

	adminURL := primary.URL
	adminDomain := domain
	if len(endpoints) > 1 {
		adminURL = endpoints[1].URL
		adminDomain = endpoints[1].Host
	}
	items = append(items, models.InstanceAccess{
		InstanceID: instanceID,
		EntryType:  "admin",
		URL:        adminURL,
		Domain:     adminDomain,
		AccessMode: "direct",
		IsPrimary:  false,
	})

	return items
}

func kubeNamespace(prefix string, tenantID int, name string, instanceID int) string {
	return sanitizeKubeName(fmt.Sprintf("%s-t%d-%s-%d", defaultString(prefix, "openclaw"), tenantID, slugify(name), instanceID), 63)
}

func kubeWorkloadName(prefix string, name string, instanceID int) string {
	return sanitizeKubeName(fmt.Sprintf("%s-%s-%d", defaultString(prefix, "openclaw"), slugify(name), instanceID), 63)
}

func sanitizeKubeName(value string, maxLen int) string {
	cleaned := slugify(value)
	cleaned = strings.Trim(cleaned, "-")
	if cleaned == "" {
		cleaned = "openclaw"
	}
	if len(cleaned) > maxLen {
		cleaned = strings.Trim(cleaned[:maxLen], "-")
	}
	if cleaned == "" {
		return "openclaw"
	}
	return cleaned
}

func visibleInstance(instance models.Instance) bool {
	return !isDeletedInstance(instance)
}

func (r *Router) syncInstanceRuntimeLocked(instanceID int, workload runtimeadapter.Workload, metrics *runtimeadapter.WorkloadMetrics) (models.Instance, models.InstanceRuntime, bool) {
	instanceIndex := r.findInstanceIndex(instanceID)
	if instanceIndex < 0 {
		return models.Instance{}, models.InstanceRuntime{}, false
	}

	r.data.Instances[instanceIndex].Status = instanceStatusFromWorkload(workload)
	runtime := buildInstanceRuntime(instanceID, r.data.Instances[instanceIndex].Spec, workload, metrics)
	r.upsertRuntimeLocked(runtime)
	return r.data.Instances[instanceIndex], runtime, true
}

func (r *Router) loadLiveInstanceState(instanceID int) (liveInstanceState, bool) {
	r.mu.RLock()
	instance, found := r.findInstance(instanceID)
	binding := r.findRuntimeBinding(instanceID)
	runtime := r.findRuntime(instanceID)
	r.mu.RUnlock()
	if !found || !visibleInstance(instance) {
		return liveInstanceState{}, false
	}

	state := liveInstanceState{
		Instance: instance,
		Binding:  binding,
		Runtime:  runtime,
	}
	if binding == nil {
		return state, true
	}

	workload, workloadFound := r.runtime.GetWorkload(binding.WorkloadID)
	if !workloadFound {
		return state, true
	}

	var metrics *runtimeadapter.WorkloadMetrics
	if runtimeMetrics, ok := r.runtime.GetMetrics(binding.WorkloadID); ok {
		metrics = &runtimeMetrics
	}

	r.mu.Lock()
	updatedInstance, updatedRuntime, ok := r.syncInstanceRuntimeLocked(instanceID, workload, metrics)
	r.mu.Unlock()
	if ok {
		if updatedInstance.Status != instance.Status {
			_ = r.persistInstanceState(instanceID)
		}
		state.Instance = updatedInstance
		state.Runtime = &updatedRuntime
	} else {
		runtimeSnapshot := buildInstanceRuntime(instanceID, instance.Spec, workload, metrics)
		state.Runtime = &runtimeSnapshot
	}
	state.Workload = &workload
	state.Metrics = metrics
	return state, true
}
