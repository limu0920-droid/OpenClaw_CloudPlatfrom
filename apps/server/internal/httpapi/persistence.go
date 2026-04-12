package httpapi

import (
	"openclaw/platformapi/internal/corestore"
	"openclaw/platformapi/internal/models"
)

func (r *Router) snapshotInstanceStateLocked(instanceID int) (corestore.InstanceState, bool) {
	if r.store == nil {
		return corestore.InstanceState{}, false
	}

	instance, found := r.findInstance(instanceID)
	if !found {
		return corestore.InstanceState{}, false
	}

	state := corestore.InstanceState{
		Instance: instance,
		Accesses: append([]models.InstanceAccess(nil), r.filterAccessByInstance(instanceID)...),
		Backups:  append([]models.BackupRecord(nil), r.filterBackupsByInstance(instanceID)...),
		Jobs:     append([]models.Job(nil), r.filterJobsByInstance(instanceID)...),
		Audits:   append([]models.AuditEvent(nil), r.filterAuditsByInstance(instanceID)...),
	}
	if config := r.findConfig(instanceID); config != nil {
		copy := *config
		state.Config = &copy
	}
	if runtime := r.findRuntime(instanceID); runtime != nil {
		copy := *runtime
		state.Runtime = &copy
	}
	if credential := r.findCredential(instanceID); credential != nil {
		copy := *credential
		state.Credential = &copy
	}
	if binding := r.findRuntimeBinding(instanceID); binding != nil {
		copy := *binding
		state.RuntimeBinding = &copy
	}
	return state, true
}

func (r *Router) persistInstanceState(instanceID int) error {
	if r.store == nil {
		return nil
	}

	r.mu.RLock()
	state, ok := r.snapshotInstanceStateLocked(instanceID)
	r.mu.RUnlock()
	if !ok {
		return nil
	}
	return r.store.SaveInstanceState(state)
}

func (r *Router) snapshotDataLocked() models.Data {
	copyData := r.data
	copyData.Tenants = append([]models.Tenant(nil), r.data.Tenants...)
	copyData.Users = append([]models.UserProfile(nil), r.data.Users...)
	copyData.AuthIdentities = append([]models.AuthIdentity(nil), r.data.AuthIdentities...)
	copyData.Clusters = append([]models.Cluster(nil), r.data.Clusters...)
	copyData.Instances = append([]models.Instance(nil), r.data.Instances...)
	copyData.Accesses = append([]models.InstanceAccess(nil), r.data.Accesses...)
	copyData.Configs = append([]models.InstanceConfig(nil), r.data.Configs...)
	copyData.Backups = append([]models.BackupRecord(nil), r.data.Backups...)
	copyData.Jobs = append([]models.Job(nil), r.data.Jobs...)
	copyData.Alerts = append([]models.Alert(nil), r.data.Alerts...)
	copyData.Audits = append([]models.AuditEvent(nil), r.data.Audits...)
	copyData.Channels = append([]models.Channel(nil), r.data.Channels...)
	copyData.Activities = append([]models.ChannelActivity(nil), r.data.Activities...)
	copyData.Runtimes = append([]models.InstanceRuntime(nil), r.data.Runtimes...)
	copyData.Credentials = append([]models.InstanceCredential(nil), r.data.Credentials...)
	copyData.PlanOffers = append([]models.PlanOffer(nil), r.data.PlanOffers...)
	copyData.Orders = append([]models.Order(nil), r.data.Orders...)
	copyData.Subscriptions = append([]models.Subscription(nil), r.data.Subscriptions...)
	copyData.Payments = append([]models.PaymentTransaction(nil), r.data.Payments...)
	copyData.Refunds = append([]models.RefundRecord(nil), r.data.Refunds...)
	copyData.Invoices = append([]models.InvoiceRecord(nil), r.data.Invoices...)
	copyData.PaymentCallbackEvents = append([]models.PaymentCallbackEvent(nil), r.data.PaymentCallbackEvents...)
	copyData.AccountSettings = append([]models.AccountSettings(nil), r.data.AccountSettings...)
	copyData.Wallets = append([]models.WalletBalance(nil), r.data.Wallets...)
	copyData.BillingStatements = append([]models.BillingStatement(nil), r.data.BillingStatements...)
	copyData.Tickets = append([]models.Ticket(nil), r.data.Tickets...)
	copyData.Brands = append([]models.OEMBrand(nil), r.data.Brands...)
	copyData.BrandThemes = append([]models.OEMTheme(nil), r.data.BrandThemes...)
	copyData.BrandFeatures = append([]models.OEMFeatureFlags(nil), r.data.BrandFeatures...)
	copyData.BrandBindings = append([]models.TenantBrandBinding(nil), r.data.BrandBindings...)
	copyData.RuntimeBindings = append([]models.RuntimeBinding(nil), r.data.RuntimeBindings...)
	copyData.Approvals = append([]models.ApprovalRecord(nil), r.data.Approvals...)
	copyData.ApprovalActions = append([]models.ApprovalAction(nil), r.data.ApprovalActions...)
	copyData.DiagnosticSessions = append([]models.DiagnosticSession(nil), r.data.DiagnosticSessions...)
	copyData.DiagnosticCommandRecords = append([]models.DiagnosticCommandRecord(nil), r.data.DiagnosticCommandRecords...)
	copyData.WorkspaceSessions = append([]models.WorkspaceSession(nil), r.data.WorkspaceSessions...)
	copyData.WorkspaceArtifacts = append([]models.WorkspaceArtifact(nil), r.data.WorkspaceArtifacts...)
	copyData.WorkspaceMessages = append([]models.WorkspaceMessage(nil), r.data.WorkspaceMessages...)
	copyData.WorkspaceMessageEvents = append([]models.WorkspaceMessageEvent(nil), r.data.WorkspaceMessageEvents...)
	copyData.WorkspaceArtifactLogs = append([]models.WorkspaceArtifactAccessLog(nil), r.data.WorkspaceArtifactLogs...)
	copyData.WorkspaceArtifactFavorites = append([]models.WorkspaceArtifactFavorite(nil), r.data.WorkspaceArtifactFavorites...)
	copyData.WorkspaceArtifactShares = append([]models.WorkspaceArtifactShare(nil), r.data.WorkspaceArtifactShares...)
	return copyData
}

func (r *Router) persistAllData() error {
	if r.store == nil {
		return nil
	}
	r.mu.RLock()
	data := r.snapshotDataLocked()
	r.mu.RUnlock()
	return r.store.SaveData(data)
}

func (r *Router) persistWorkspaceMutation(mutation corestore.WorkspaceMutation) error {
	if r.store == nil {
		return nil
	}
	return r.store.SaveWorkspaceMutation(mutation)
}

func (r *Router) persistDiagnosticsMutation(mutation corestore.DiagnosticsMutation) error {
	if r.store == nil {
		return nil
	}
	return r.store.SaveDiagnosticsMutation(mutation)
}
