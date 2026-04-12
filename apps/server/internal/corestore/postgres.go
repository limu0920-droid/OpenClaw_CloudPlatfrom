package corestore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"openclaw/platformapi/internal/models"
)

type CoreStore interface {
	Migrate(ctx context.Context) error
	Bootstrap(seed models.Data) (models.Data, error)
	Load(seed models.Data) (models.Data, error)
	SaveInstanceState(state InstanceState) error
	SaveData(data models.Data) error
	SaveDiagnosticsMutation(mutation DiagnosticsMutation) error
	SaveWorkspaceMutation(mutation WorkspaceMutation) error
	Ping(ctx context.Context) error
	Close() error
}

type InstanceState struct {
	Instance       models.Instance
	Accesses       []models.InstanceAccess
	Config         *models.InstanceConfig
	Runtime        *models.InstanceRuntime
	Credential     *models.InstanceCredential
	Backups        []models.BackupRecord
	Jobs           []models.Job
	Audits         []models.AuditEvent
	RuntimeBinding *models.RuntimeBinding
}

type WorkspaceMutation struct {
	Sessions     []models.WorkspaceSession
	Artifacts    []models.WorkspaceArtifact
	Messages     []models.WorkspaceMessage
	Events       []models.WorkspaceMessageEvent
	ArtifactLogs []models.WorkspaceArtifactAccessLog
	Favorites    []models.WorkspaceArtifactFavorite
	Shares       []models.WorkspaceArtifactShare
}

type DiagnosticsMutation struct {
	Sessions []models.DiagnosticSession
	Commands []models.DiagnosticCommandRecord
}

type PostgresStore struct {
	db *sql.DB
}

func Open(databaseURL string) (*PostgresStore, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *PostgresStore) Ping(ctx context.Context) error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.PingContext(ctx)
}

func (s *PostgresStore) Bootstrap(seed models.Data) (models.Data, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return models.Data{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if err := lockBootstrap(ctx, tx); err != nil {
		return models.Data{}, err
	}

	if err := s.ensureReferenceSeed(ctx, tx, seed); err != nil {
		return models.Data{}, err
	}

	instanceCount, err := countRows(ctx, tx, "platform.service_instance")
	if err != nil {
		return models.Data{}, err
	}
	if instanceCount == 0 {
		for _, instance := range seed.Instances {
			if err := s.saveInstanceStateTx(ctx, tx, buildSeedInstanceState(seed, instance.ID)); err != nil {
				return models.Data{}, err
			}
		}
	}
	if err := s.seedAuxiliaryDomainsIfEmpty(ctx, tx, seed); err != nil {
		return models.Data{}, err
	}

	if err := resetManagedSequences(ctx, tx); err != nil {
		return models.Data{}, err
	}
	if err := tx.Commit(); err != nil {
		return models.Data{}, err
	}

	return s.load(seed)
}

func (s *PostgresStore) SaveInstanceState(state InstanceState) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if err := s.saveInstanceStateTx(ctx, tx, state); err != nil {
		return err
	}
	if err := resetManagedSequences(ctx, tx); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *PostgresStore) SaveData(data models.Data) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if err := s.ensureReferenceSeed(ctx, tx, data); err != nil {
		return err
	}
	if err := upsertUsers(ctx, tx, data.Users); err != nil {
		return err
	}
	if err := replaceAuthIdentities(ctx, tx, data.AuthIdentities); err != nil {
		return err
	}
	if err := replaceChannels(ctx, tx, data.Channels, data.Activities); err != nil {
		return err
	}
	if err := replaceAccountSettings(ctx, tx, data.AccountSettings); err != nil {
		return err
	}
	if err := replaceWallets(ctx, tx, data.Wallets); err != nil {
		return err
	}
	if err := replaceBillingStatements(ctx, tx, data.BillingStatements); err != nil {
		return err
	}
	if err := replaceRuntimes(ctx, tx, data.Runtimes); err != nil {
		return err
	}
	if err := replaceCredentials(ctx, tx, data.Credentials); err != nil {
		return err
	}
	if err := upsertAudits(ctx, tx, data.Audits); err != nil {
		return err
	}
	if err := replaceAlerts(ctx, tx, data.Alerts); err != nil {
		return err
	}
	if err := replaceTickets(ctx, tx, data.Tickets); err != nil {
		return err
	}
	if err := replaceApprovals(ctx, tx, data.Approvals, data.ApprovalActions); err != nil {
		return err
	}
	if err := replaceDiagnostics(ctx, tx, data.DiagnosticSessions, data.DiagnosticCommandRecords); err != nil {
		return err
	}
	if err := replaceWorkspace(
		ctx,
		tx,
		data.WorkspaceSessions,
		data.WorkspaceArtifacts,
		data.WorkspaceMessages,
		data.WorkspaceMessageEvents,
		data.WorkspaceArtifactLogs,
		data.WorkspaceArtifactFavorites,
		data.WorkspaceArtifactShares,
	); err != nil {
		return err
	}
	if err := replaceOEM(ctx, tx, data.Brands, data.BrandThemes, data.BrandFeatures, data.BrandBindings); err != nil {
		return err
	}
	if err := replaceCommerce(ctx, tx, data.Orders, data.Subscriptions, data.Payments, data.Refunds, data.Invoices, data.PaymentCallbackEvents); err != nil {
		return err
	}
	if err := resetManagedSequences(ctx, tx); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *PostgresStore) SaveDiagnosticsMutation(mutation DiagnosticsMutation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	sequenceTables := make(map[string]struct{})
	for _, item := range mutation.Sessions {
		if err := upsertDiagnosticSessionTx(ctx, tx, item); err != nil {
			return err
		}
		sequenceTables["platform.diagnostic_session"] = struct{}{}
	}
	for _, item := range mutation.Commands {
		if err := upsertDiagnosticCommandRecordTx(ctx, tx, item); err != nil {
			return err
		}
		sequenceTables["platform.diagnostic_command_record"] = struct{}{}
	}
	for table := range sequenceTables {
		if err := resetTableSequence(ctx, tx, table, "id"); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *PostgresStore) SaveWorkspaceMutation(mutation WorkspaceMutation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	sequenceTables := make(map[string]struct{})
	for _, item := range mutation.Sessions {
		if err := upsertWorkspaceSessionTx(ctx, tx, item); err != nil {
			return err
		}
		sequenceTables["platform.workspace_session"] = struct{}{}
	}
	for _, item := range mutation.Artifacts {
		if err := upsertWorkspaceArtifactTx(ctx, tx, item); err != nil {
			return err
		}
		sequenceTables["platform.workspace_artifact"] = struct{}{}
	}
	for _, item := range mutation.Messages {
		if err := upsertWorkspaceMessageTx(ctx, tx, item); err != nil {
			return err
		}
		sequenceTables["platform.workspace_message"] = struct{}{}
	}
	for _, item := range mutation.Events {
		if err := upsertWorkspaceMessageEventTx(ctx, tx, item); err != nil {
			return err
		}
		sequenceTables["platform.workspace_message_event"] = struct{}{}
	}
	for _, item := range mutation.ArtifactLogs {
		if err := upsertWorkspaceArtifactAccessLogTx(ctx, tx, item); err != nil {
			return err
		}
		sequenceTables["platform.workspace_artifact_access_log"] = struct{}{}
	}
	for _, item := range mutation.Favorites {
		if err := upsertWorkspaceArtifactFavoriteTx(ctx, tx, item); err != nil {
			return err
		}
		sequenceTables["platform.workspace_artifact_favorite"] = struct{}{}
	}
	for _, item := range mutation.Shares {
		if err := upsertWorkspaceArtifactShareTx(ctx, tx, item); err != nil {
			return err
		}
		sequenceTables["platform.workspace_artifact_share"] = struct{}{}
	}
	for table := range sequenceTables {
		if err := resetTableSequence(ctx, tx, table, "id"); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *PostgresStore) Load(seed models.Data) (models.Data, error) {
	return s.load(seed)
}

func (s *PostgresStore) saveInstanceStateTx(ctx context.Context, tx *sql.Tx, state InstanceState) error {
	planID, err := lookupPlanID(ctx, tx, state.Instance.Plan)
	if err != nil {
		return err
	}

	if err := upsertServiceInstance(ctx, tx, state.Instance, planID); err != nil {
		return err
	}
	if err := replaceInstanceAccesses(ctx, tx, state.Instance.ID, state.Accesses); err != nil {
		return err
	}
	if err := syncRuntimeBinding(ctx, tx, state.Instance, state.RuntimeBinding); err != nil {
		return err
	}
	if state.Config != nil {
		if err := upsertInstanceConfig(ctx, tx, *state.Config); err != nil {
			return err
		}
	}
	if err := upsertBackups(ctx, tx, state.Backups); err != nil {
		return err
	}
	if err := upsertRuntimeState(ctx, tx, state.Runtime); err != nil {
		return err
	}
	if err := upsertCredential(ctx, tx, state.Credential); err != nil {
		return err
	}
	if err := upsertJobs(ctx, tx, state.Instance.TenantID, state.Jobs); err != nil {
		return err
	}
	if err := upsertAudits(ctx, tx, state.Audits); err != nil {
		return err
	}
	return nil
}

func buildSeedInstanceState(seed models.Data, instanceID int) InstanceState {
	state := InstanceState{}
	for _, item := range seed.Instances {
		if item.ID == instanceID {
			state.Instance = item
			break
		}
	}
	for _, item := range seed.Accesses {
		if item.InstanceID == instanceID {
			state.Accesses = append(state.Accesses, item)
		}
	}
	for _, item := range seed.Configs {
		if item.InstanceID == instanceID {
			copy := item
			state.Config = &copy
			break
		}
	}
	for _, item := range seed.Runtimes {
		if item.InstanceID == instanceID {
			copy := item
			state.Runtime = &copy
			break
		}
	}
	for _, item := range seed.Credentials {
		if item.InstanceID == instanceID {
			copy := item
			state.Credential = &copy
			break
		}
	}
	for _, item := range seed.Backups {
		if item.InstanceID == instanceID {
			state.Backups = append(state.Backups, item)
		}
	}
	for _, item := range seed.Jobs {
		if item.TargetType == "instance" && item.TargetID == instanceID {
			state.Jobs = append(state.Jobs, item)
		}
	}
	for _, item := range seed.Audits {
		if item.Target == "instance" && item.TargetID == instanceID {
			state.Audits = append(state.Audits, item)
		}
	}
	for _, item := range seed.RuntimeBindings {
		if item.InstanceID == instanceID {
			copy := item
			state.RuntimeBinding = &copy
			break
		}
	}
	return state
}

func (s *PostgresStore) ensureReferenceSeed(ctx context.Context, tx *sql.Tx, seed models.Data) error {
	productID, err := upsertProduct(ctx, tx)
	if err != nil {
		return err
	}
	for _, offer := range seed.PlanOffers {
		if err := upsertPlan(ctx, tx, productID, offer); err != nil {
			return err
		}
	}
	for _, tenant := range seed.Tenants {
		if err := upsertTenant(ctx, tx, tenant); err != nil {
			return err
		}
	}
	for _, cluster := range seed.Clusters {
		if err := upsertCluster(ctx, tx, cluster); err != nil {
			return err
		}
	}
	return nil
}

func (s *PostgresStore) seedAuxiliaryDomainsIfEmpty(ctx context.Context, tx *sql.Tx, seed models.Data) error {
	checks := []struct {
		table string
		run   func() error
	}{
		{"platform.user_account", func() error { return upsertUsers(ctx, tx, seed.Users) }},
		{"platform.auth_identity", func() error { return replaceAuthIdentities(ctx, tx, seed.AuthIdentities) }},
		{"platform.channel", func() error { return replaceChannels(ctx, tx, seed.Channels, seed.Activities) }},
		{"platform.account_settings", func() error { return replaceAccountSettings(ctx, tx, seed.AccountSettings) }},
		{"platform.wallet_balance", func() error { return replaceWallets(ctx, tx, seed.Wallets) }},
		{"platform.billing_statement", func() error { return replaceBillingStatements(ctx, tx, seed.BillingStatements) }},
		{"platform.instance_runtime_state", func() error { return replaceRuntimes(ctx, tx, seed.Runtimes) }},
		{"platform.instance_credential", func() error { return replaceCredentials(ctx, tx, seed.Credentials) }},
		{"platform.alert_record", func() error { return replaceAlerts(ctx, tx, seed.Alerts) }},
		{"platform.ticket_record", func() error { return replaceTickets(ctx, tx, seed.Tickets) }},
		{"platform.approval_record", func() error { return replaceApprovals(ctx, tx, seed.Approvals, seed.ApprovalActions) }},
		{"platform.diagnostic_session", func() error {
			return replaceDiagnostics(ctx, tx, seed.DiagnosticSessions, seed.DiagnosticCommandRecords)
		}},
		{"platform.workspace_session", func() error {
			return replaceWorkspace(
				ctx,
				tx,
				seed.WorkspaceSessions,
				seed.WorkspaceArtifacts,
				seed.WorkspaceMessages,
				seed.WorkspaceMessageEvents,
				seed.WorkspaceArtifactLogs,
				seed.WorkspaceArtifactFavorites,
				seed.WorkspaceArtifactShares,
			)
		}},
		{"platform.oem_brand", func() error {
			return replaceOEM(ctx, tx, seed.Brands, seed.BrandThemes, seed.BrandFeatures, seed.BrandBindings)
		}},
		{"platform.order_main", func() error {
			return replaceCommerce(ctx, tx, seed.Orders, seed.Subscriptions, seed.Payments, seed.Refunds, seed.Invoices, seed.PaymentCallbackEvents)
		}},
	}
	for _, check := range checks {
		count, err := countRows(ctx, tx, check.table)
		if err != nil {
			return err
		}
		if count == 0 {
			if err := check.run(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *PostgresStore) load(seed models.Data) (models.Data, error) {
	_ = seed
	data := newEmptyData()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if items, err := loadPlanOffers(ctx, s.db); err != nil {
		return models.Data{}, err
	} else {
		data.PlanOffers = items
	}
	if items, err := loadTenants(ctx, s.db); err != nil {
		return models.Data{}, err
	} else {
		data.Tenants = items
	}
	if items, err := loadClusters(ctx, s.db); err != nil {
		return models.Data{}, err
	} else {
		data.Clusters = items
	}
	if items, err := loadInstances(ctx, s.db); err != nil {
		return models.Data{}, err
	} else {
		data.Instances = items
	}
	if items, err := loadInstanceAccesses(ctx, s.db); err != nil {
		return models.Data{}, err
	} else {
		data.Accesses = items
	}
	if items, err := loadInstanceConfigs(ctx, s.db); err != nil {
		return models.Data{}, err
	} else {
		data.Configs = items
	}
	if items, err := loadBackups(ctx, s.db); err != nil {
		return models.Data{}, err
	} else {
		data.Backups = items
	}
	if items, err := loadJobs(ctx, s.db); err != nil {
		return models.Data{}, err
	} else {
		data.Jobs = items
	}
	if items, err := loadAudits(ctx, s.db); err != nil {
		return models.Data{}, err
	} else {
		data.Audits = items
	}
	if items, err := loadRuntimeBindings(ctx, s.db); err != nil {
		return models.Data{}, err
	} else {
		data.RuntimeBindings = items
	}
	if items, err := loadUsers(ctx, s.db); err != nil {
		return models.Data{}, err
	} else {
		data.Users = items
	}
	if items, err := loadAuthIdentities(ctx, s.db); err != nil {
		return models.Data{}, err
	} else {
		data.AuthIdentities = items
	}
	if channels, activities, err := loadChannels(ctx, s.db); err != nil {
		return models.Data{}, err
	} else {
		data.Channels = channels
		data.Activities = activities
	}
	if items, err := loadAccountSettings(ctx, s.db); err != nil {
		return models.Data{}, err
	} else {
		data.AccountSettings = items
	}
	if items, err := loadWallets(ctx, s.db); err != nil {
		return models.Data{}, err
	} else {
		data.Wallets = items
	}
	if items, err := loadBillingStatements(ctx, s.db); err != nil {
		return models.Data{}, err
	} else {
		data.BillingStatements = items
	}
	if items, err := loadRuntimes(ctx, s.db); err != nil {
		return models.Data{}, err
	} else {
		data.Runtimes = items
	}
	if items, err := loadCredentials(ctx, s.db); err != nil {
		return models.Data{}, err
	} else {
		data.Credentials = items
	}
	if items, err := loadAlerts(ctx, s.db); err != nil {
		return models.Data{}, err
	} else {
		data.Alerts = items
	}
	if items, err := loadTickets(ctx, s.db); err != nil {
		return models.Data{}, err
	} else {
		data.Tickets = items
	}
	if approvals, actions, err := loadApprovals(ctx, s.db); err != nil {
		return models.Data{}, err
	} else {
		data.Approvals = approvals
		data.ApprovalActions = actions
	}
	if sessions, commands, err := loadDiagnostics(ctx, s.db); err != nil {
		return models.Data{}, err
	} else {
		data.DiagnosticSessions = sessions
		data.DiagnosticCommandRecords = commands
	}
	if sessions, artifacts, messages, events, logs, favorites, shares, err := loadWorkspace(ctx, s.db); err != nil {
		return models.Data{}, err
	} else {
		data.WorkspaceSessions = sessions
		data.WorkspaceArtifacts = artifacts
		data.WorkspaceMessages = messages
		data.WorkspaceMessageEvents = events
		data.WorkspaceArtifactLogs = logs
		data.WorkspaceArtifactFavorites = favorites
		data.WorkspaceArtifactShares = shares
	}
	if brands, themes, features, bindings, err := loadOEM(ctx, s.db); err != nil {
		return models.Data{}, err
	} else {
		data.Brands = brands
		data.BrandThemes = themes
		data.BrandFeatures = features
		data.BrandBindings = bindings
	}
	if orders, subscriptions, payments, refunds, invoices, callbackEvents, err := loadCommerce(ctx, s.db); err != nil {
		return models.Data{}, err
	} else {
		data.Orders = orders
		data.Subscriptions = subscriptions
		data.Payments = payments
		data.Refunds = refunds
		data.Invoices = invoices
		data.PaymentCallbackEvents = callbackEvents
	}
	return data, nil
}

func newEmptyData() models.Data {
	return models.Data{
		Tenants:                    []models.Tenant{},
		Users:                      []models.UserProfile{},
		AuthIdentities:             []models.AuthIdentity{},
		Clusters:                   []models.Cluster{},
		Instances:                  []models.Instance{},
		Accesses:                   []models.InstanceAccess{},
		Configs:                    []models.InstanceConfig{},
		Backups:                    []models.BackupRecord{},
		Jobs:                       []models.Job{},
		Alerts:                     []models.Alert{},
		Audits:                     []models.AuditEvent{},
		Channels:                   []models.Channel{},
		Activities:                 []models.ChannelActivity{},
		Runtimes:                   []models.InstanceRuntime{},
		Credentials:                []models.InstanceCredential{},
		PlanOffers:                 []models.PlanOffer{},
		Orders:                     []models.Order{},
		Subscriptions:              []models.Subscription{},
		Payments:                   []models.PaymentTransaction{},
		Refunds:                    []models.RefundRecord{},
		Invoices:                   []models.InvoiceRecord{},
		PaymentCallbackEvents:      []models.PaymentCallbackEvent{},
		AccountSettings:            []models.AccountSettings{},
		Wallets:                    []models.WalletBalance{},
		BillingStatements:          []models.BillingStatement{},
		Tickets:                    []models.Ticket{},
		Brands:                     []models.OEMBrand{},
		BrandThemes:                []models.OEMTheme{},
		BrandFeatures:              []models.OEMFeatureFlags{},
		BrandBindings:              []models.TenantBrandBinding{},
		RuntimeBindings:            []models.RuntimeBinding{},
		Approvals:                  []models.ApprovalRecord{},
		ApprovalActions:            []models.ApprovalAction{},
		DiagnosticSessions:         []models.DiagnosticSession{},
		DiagnosticCommandRecords:   []models.DiagnosticCommandRecord{},
		WorkspaceSessions:          []models.WorkspaceSession{},
		WorkspaceArtifacts:         []models.WorkspaceArtifact{},
		WorkspaceMessages:          []models.WorkspaceMessage{},
		WorkspaceMessageEvents:     []models.WorkspaceMessageEvent{},
		WorkspaceArtifactLogs:      []models.WorkspaceArtifactAccessLog{},
		WorkspaceArtifactFavorites: []models.WorkspaceArtifactFavorite{},
		WorkspaceArtifactShares:    []models.WorkspaceArtifactShare{},
	}
}

func upsertProduct(ctx context.Context, tx *sql.Tx) (int, error) {
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO platform.product (id, code, name, type, status, description)
		VALUES (1, 'openclaw', 'OpenClaw', 'saas', 'active', 'OpenClaw platform')
		ON CONFLICT (id) DO UPDATE
		SET code = EXCLUDED.code,
		    name = EXCLUDED.name,
		    type = EXCLUDED.type,
		    status = EXCLUDED.status,
		    description = EXCLUDED.description,
		    updated_at = NOW()`); err != nil {
		return 0, err
	}
	return 1, nil
}

func upsertPlan(ctx context.Context, tx *sql.Tx, productID int, offer models.PlanOffer) error {
	resourceSpec, _ := json.Marshal(map[string]any{
		"cpu":     offer.CPU,
		"memory":  offer.Memory,
		"storage": offer.Storage,
	})
	featureSpec, _ := json.Marshal(map[string]any{
		"highlight": offer.Highlight,
		"features":  offer.Features,
	})

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO platform.service_plan (id, product_id, code, name, status, billing_mode, trial_supported, resource_spec, feature_spec)
		VALUES ($1, $2, $3, $4, 'active', 'subscription', $5, $6::jsonb, $7::jsonb)
		ON CONFLICT (id) DO UPDATE
		SET product_id = EXCLUDED.product_id,
		    code = EXCLUDED.code,
		    name = EXCLUDED.name,
		    trial_supported = EXCLUDED.trial_supported,
		    resource_spec = EXCLUDED.resource_spec,
		    feature_spec = EXCLUDED.feature_spec,
		    updated_at = NOW()
	`, offer.ID, productID, offer.Code, offer.Name, offer.Code == "trial", string(resourceSpec), string(featureSpec)); err != nil {
		return err
	}

	_, err := tx.ExecContext(ctx, `
		INSERT INTO platform.plan_price (plan_id, billing_cycle, currency, amount, status)
		VALUES ($1, 'monthly', 'CNY', $2, 'active')
		ON CONFLICT (plan_id, billing_cycle, currency) DO UPDATE
		SET amount = EXCLUDED.amount,
		    status = EXCLUDED.status,
		    updated_at = NOW()
	`, offer.ID, offer.MonthlyPrice)
	return err
}

func upsertTenant(ctx context.Context, tx *sql.Tx, tenant models.Tenant) error {
	planID, err := lookupPlanID(ctx, tx, tenant.Plan)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO platform.tenant (id, tenant_code, name, status, plan_id, expired_at, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, '{}'::jsonb, $7, $8)
		ON CONFLICT (id) DO UPDATE
		SET tenant_code = EXCLUDED.tenant_code,
		    name = EXCLUDED.name,
		    status = EXCLUDED.status,
		    plan_id = EXCLUDED.plan_id,
		    expired_at = EXCLUDED.expired_at,
		    updated_at = EXCLUDED.updated_at
	`, tenant.ID, tenant.Code, tenant.Name, tenant.Status, planID, nullTime(tenant.ExpiredAt), tenant.CreatedAt, tenant.UpdatedAt)
	return err
}

func upsertCluster(ctx context.Context, tx *sql.Tx, cluster models.Cluster) error {
	metadata, _ := json.Marshal(map[string]any{"nodeCount": cluster.NodeCount})
	_, err := tx.ExecContext(ctx, `
		INSERT INTO platform.cluster (id, code, name, region, status, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8)
		ON CONFLICT (id) DO UPDATE
		SET code = EXCLUDED.code,
		    name = EXCLUDED.name,
		    region = EXCLUDED.region,
		    status = EXCLUDED.status,
		    metadata = EXCLUDED.metadata,
		    updated_at = EXCLUDED.updated_at
	`, cluster.ID, cluster.Code, cluster.Name, cluster.Region, cluster.Status, string(metadata), cluster.CreatedAt, cluster.UpdatedAt)
	return err
}

func lookupPlanID(ctx context.Context, tx *sql.Tx, code string) (int, error) {
	var id int
	if err := tx.QueryRowContext(ctx, `SELECT id FROM platform.service_plan WHERE code = $1`, code).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func upsertServiceInstance(ctx context.Context, tx *sql.Tx, instance models.Instance, planID int) error {
	resourceSpec, _ := json.Marshal(instance.Spec)
	metadata, _ := json.Marshal(map[string]any{
		"planCode": instance.Plan,
		"region":   instance.Region,
	})

	_, err := tx.ExecContext(ctx, `
		INSERT INTO platform.service_instance (
			id, tenant_id, cluster_id, plan_id, instance_code, display_name, status, version,
			runtime_type, resource_spec, metadata, activated_at, expired_at, created_at, updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8,
			$9, $10::jsonb, $11::jsonb, $12, $13, $14, $15
		)
		ON CONFLICT (id) DO UPDATE
		SET tenant_id = EXCLUDED.tenant_id,
		    cluster_id = EXCLUDED.cluster_id,
		    plan_id = EXCLUDED.plan_id,
		    instance_code = EXCLUDED.instance_code,
		    display_name = EXCLUDED.display_name,
		    status = EXCLUDED.status,
		    version = EXCLUDED.version,
		    runtime_type = EXCLUDED.runtime_type,
		    resource_spec = EXCLUDED.resource_spec,
		    metadata = EXCLUDED.metadata,
		    activated_at = EXCLUDED.activated_at,
		    expired_at = EXCLUDED.expired_at,
		    updated_at = EXCLUDED.updated_at
	`, instance.ID, instance.TenantID, instance.ClusterID, planID, instance.Code, instance.Name, instance.Status, instance.Version,
		instance.RuntimeType, string(resourceSpec), string(metadata), nullTime(instance.ActivatedAt), nullTime(instance.ExpiredAt),
		instance.CreatedAt, instance.UpdatedAt)
	return err
}

func replaceInstanceAccesses(ctx context.Context, tx *sql.Tx, instanceID int, accesses []models.InstanceAccess) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform.instance_access WHERE instance_id = $1`, instanceID); err != nil {
		return err
	}
	for _, access := range accesses {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO platform.instance_access (instance_id, entry_type, url, domain, access_mode, is_primary, metadata)
			VALUES ($1, $2, $3, $4, $5, $6, '{}'::jsonb)
		`, instanceID, access.EntryType, access.URL, access.Domain, access.AccessMode, access.IsPrimary); err != nil {
			return err
		}
	}
	return nil
}

func syncRuntimeBinding(ctx context.Context, tx *sql.Tx, instance models.Instance, binding *models.RuntimeBinding) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform.runtime_container WHERE instance_id = $1`, instance.ID); err != nil {
		return err
	}
	if binding == nil {
		return nil
	}

	metadata, _ := json.Marshal(map[string]any{
		"clusterId":    binding.ClusterID,
		"namespace":    binding.Namespace,
		"workloadId":   binding.WorkloadID,
		"workloadName": binding.WorkloadName,
	})
	_, err := tx.ExecContext(ctx, `
		INSERT INTO platform.runtime_container (
			instance_id, runtime_type, container_ref, container_name, status, image_ref, metadata
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb)
	`, instance.ID, instance.RuntimeType, binding.WorkloadID, binding.WorkloadName, instance.Status, "", string(metadata))
	return err
}

func upsertInstanceConfig(ctx context.Context, tx *sql.Tx, config models.InstanceConfig) error {
	payload, _ := json.Marshal(map[string]any{
		"settings":  config.Settings,
		"updatedBy": config.UpdatedBy,
	})

	_, err := tx.ExecContext(ctx, `
		INSERT INTO platform.instance_config (
			instance_id, config_version, config_json, config_hash, published_at, updated_at
		)
		VALUES ($1, $2, $3::jsonb, $4, $5, NOW())
		ON CONFLICT (instance_id) DO UPDATE
		SET config_version = EXCLUDED.config_version,
		    config_json = EXCLUDED.config_json,
		    config_hash = EXCLUDED.config_hash,
		    published_at = EXCLUDED.published_at,
		    updated_at = NOW()
	`, config.InstanceID, config.Version, string(payload), config.Hash, nullTime(config.PublishedAt))
	return err
}

func upsertBackups(ctx context.Context, tx *sql.Tx, backups []models.BackupRecord) error {
	for _, backup := range backups {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO platform.backup_record (
				id, instance_id, backup_no, backup_type, status, size_bytes, started_at, finished_at, created_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, COALESCE($7, NOW()))
			ON CONFLICT (id) DO UPDATE
			SET instance_id = EXCLUDED.instance_id,
			    backup_no = EXCLUDED.backup_no,
			    backup_type = EXCLUDED.backup_type,
			    status = EXCLUDED.status,
			    size_bytes = EXCLUDED.size_bytes,
			    started_at = EXCLUDED.started_at,
			    finished_at = EXCLUDED.finished_at
		`, backup.ID, backup.InstanceID, backup.BackupNo, backup.Type, backup.Status, backup.SizeBytes, nullTime(backup.StartedAt), nullTime(backup.FinishedAt))
		if err != nil {
			return err
		}
	}
	return nil
}

func upsertJobs(ctx context.Context, tx *sql.Tx, tenantID int, jobs []models.Job) error {
	for _, job := range jobs {
		payload, _ := json.Marshal(map[string]any{"summary": job.Summary})
		_, err := tx.ExecContext(ctx, `
			INSERT INTO platform.operation_job (
				id, job_no, job_type, target_type, target_id, tenant_id, status, payload_json,
				started_at, finished_at, created_at, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, $9, $10, COALESCE($9, NOW()), NOW())
			ON CONFLICT (id) DO UPDATE
			SET job_no = EXCLUDED.job_no,
			    job_type = EXCLUDED.job_type,
			    target_type = EXCLUDED.target_type,
			    target_id = EXCLUDED.target_id,
			    tenant_id = EXCLUDED.tenant_id,
			    status = EXCLUDED.status,
			    payload_json = EXCLUDED.payload_json,
			    started_at = EXCLUDED.started_at,
			    finished_at = EXCLUDED.finished_at,
			    updated_at = NOW()
		`, job.ID, job.JobNo, job.Type, job.TargetType, job.TargetID, tenantID, job.Status, string(payload), nullTime(job.StartedAt), nullTime(job.FinishedAt))
		if err != nil {
			return err
		}
	}
	return nil
}

func upsertAudits(ctx context.Context, tx *sql.Tx, audits []models.AuditEvent) error {
	for _, audit := range audits {
		payload, _ := json.Marshal(map[string]any{
			"actor":    audit.Actor,
			"action":   audit.Action,
			"result":   audit.Result,
			"metadata": audit.Metadata,
		})
		_, err := tx.ExecContext(ctx, `
			INSERT INTO platform.audit_event (
				id, tenant_id, event_type, target_type, target_id, content_json, created_at
			)
			VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7)
			ON CONFLICT (id) DO UPDATE
			SET tenant_id = EXCLUDED.tenant_id,
			    event_type = EXCLUDED.event_type,
			    target_type = EXCLUDED.target_type,
			    target_id = EXCLUDED.target_id,
			    content_json = EXCLUDED.content_json,
			    created_at = EXCLUDED.created_at
		`, audit.ID, audit.TenantID, audit.Action, audit.Target, audit.TargetID, string(payload), audit.CreatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func loadTenants(ctx context.Context, db *sql.DB) ([]models.Tenant, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT t.id, t.tenant_code, t.name, t.status, COALESCE(sp.code, ''), t.expired_at, t.created_at, t.updated_at
		FROM platform.tenant t
		LEFT JOIN platform.service_plan sp ON sp.id = t.plan_id
		ORDER BY t.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.Tenant, 0)
	for rows.Next() {
		var item models.Tenant
		var expiredAt sql.NullTime
		var createdAt time.Time
		var updatedAt time.Time
		if err := rows.Scan(&item.ID, &item.Code, &item.Name, &item.Status, &item.Plan, &expiredAt, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		item.ExpiredAt = formatNullTime(expiredAt)
		item.CreatedAt = formatTime(createdAt)
		item.UpdatedAt = formatTime(updatedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func loadPlanOffers(ctx context.Context, db *sql.DB) ([]models.PlanOffer, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT sp.id, sp.code, sp.name, COALESCE(pp.amount, 0), sp.resource_spec, sp.feature_spec
		FROM platform.service_plan sp
		LEFT JOIN platform.plan_price pp
		  ON pp.plan_id = sp.id
		 AND pp.billing_cycle = 'monthly'
		 AND pp.currency = 'CNY'
		ORDER BY sp.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.PlanOffer, 0)
	for rows.Next() {
		var item models.PlanOffer
		var resourceSpecRaw []byte
		var featureSpecRaw []byte
		if err := rows.Scan(&item.ID, &item.Code, &item.Name, &item.MonthlyPrice, &resourceSpecRaw, &featureSpecRaw); err != nil {
			return nil, err
		}

		var resourceSpec struct {
			CPU     string `json:"cpu"`
			Memory  string `json:"memory"`
			Storage string `json:"storage"`
		}
		var featureSpec struct {
			Highlight string   `json:"highlight"`
			Features  []string `json:"features"`
		}
		_ = json.Unmarshal(resourceSpecRaw, &resourceSpec)
		_ = json.Unmarshal(featureSpecRaw, &featureSpec)

		item.CPU = resourceSpec.CPU
		item.Memory = resourceSpec.Memory
		item.Storage = resourceSpec.Storage
		item.Highlight = featureSpec.Highlight
		item.Features = append([]string(nil), featureSpec.Features...)
		if item.Features == nil {
			item.Features = []string{}
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func loadClusters(ctx context.Context, db *sql.DB) ([]models.Cluster, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, code, name, region, status, metadata, created_at, updated_at
		FROM platform.cluster
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.Cluster, 0)
	for rows.Next() {
		var item models.Cluster
		var metadataRaw []byte
		var createdAt time.Time
		var updatedAt time.Time
		if err := rows.Scan(&item.ID, &item.Code, &item.Name, &item.Region, &item.Status, &metadataRaw, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		var metadata map[string]any
		_ = json.Unmarshal(metadataRaw, &metadata)
		if nodeCount, ok := metadata["nodeCount"].(float64); ok {
			item.NodeCount = int(nodeCount)
		}
		item.CreatedAt = formatTime(createdAt)
		item.UpdatedAt = formatTime(updatedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func loadInstances(ctx context.Context, db *sql.DB) ([]models.Instance, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT si.id, si.tenant_id, si.cluster_id, si.instance_code, si.display_name, si.status, si.version,
		       COALESCE(sp.code, ''), si.runtime_type, si.resource_spec, si.metadata, si.activated_at,
		       si.expired_at, si.created_at, si.updated_at
		FROM platform.service_instance si
		LEFT JOIN platform.service_plan sp ON sp.id = si.plan_id
		ORDER BY si.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.Instance, 0)
	for rows.Next() {
		var item models.Instance
		var specRaw []byte
		var metadataRaw []byte
		var activatedAt sql.NullTime
		var expiredAt sql.NullTime
		var createdAt time.Time
		var updatedAt time.Time
		if err := rows.Scan(
			&item.ID, &item.TenantID, &item.ClusterID, &item.Code, &item.Name, &item.Status, &item.Version,
			&item.Plan, &item.RuntimeType, &specRaw, &metadataRaw, &activatedAt, &expiredAt, &createdAt, &updatedAt,
		); err != nil {
			return nil, err
		}
		item.Spec = make(map[string]string)
		_ = json.Unmarshal(specRaw, &item.Spec)
		var metadata map[string]any
		_ = json.Unmarshal(metadataRaw, &metadata)
		if item.Region == "" {
			if value, ok := metadata["region"].(string); ok {
				item.Region = value
			}
		}
		item.ActivatedAt = formatNullTime(activatedAt)
		item.ExpiredAt = formatNullTime(expiredAt)
		item.CreatedAt = formatTime(createdAt)
		item.UpdatedAt = formatTime(updatedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func loadInstanceAccesses(ctx context.Context, db *sql.DB) ([]models.InstanceAccess, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT instance_id, entry_type, url, domain, access_mode, is_primary
		FROM platform.instance_access
		ORDER BY instance_id, is_primary DESC, entry_type
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.InstanceAccess, 0)
	for rows.Next() {
		var item models.InstanceAccess
		if err := rows.Scan(&item.InstanceID, &item.EntryType, &item.URL, &item.Domain, &item.AccessMode, &item.IsPrimary); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func loadInstanceConfigs(ctx context.Context, db *sql.DB) ([]models.InstanceConfig, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT instance_id, config_version, config_json, config_hash, published_at
		FROM platform.instance_config
		ORDER BY instance_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.InstanceConfig, 0)
	for rows.Next() {
		var item models.InstanceConfig
		var payloadRaw []byte
		var publishedAt sql.NullTime
		if err := rows.Scan(&item.InstanceID, &item.Version, &payloadRaw, &item.Hash, &publishedAt); err != nil {
			return nil, err
		}
		var payload struct {
			Settings  models.ConfigSettings `json:"settings"`
			UpdatedBy string                `json:"updatedBy"`
		}
		_ = json.Unmarshal(payloadRaw, &payload)
		item.Settings = payload.Settings
		item.UpdatedBy = payload.UpdatedBy
		item.PublishedAt = formatNullTime(publishedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func loadBackups(ctx context.Context, db *sql.DB) ([]models.BackupRecord, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, instance_id, backup_no, backup_type, status, COALESCE(size_bytes, 0), started_at, finished_at
		FROM platform.backup_record
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.BackupRecord, 0)
	for rows.Next() {
		var item models.BackupRecord
		var startedAt sql.NullTime
		var finishedAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.InstanceID, &item.BackupNo, &item.Type, &item.Status, &item.SizeBytes, &startedAt, &finishedAt); err != nil {
			return nil, err
		}
		item.StartedAt = formatNullTime(startedAt)
		item.FinishedAt = formatNullTime(finishedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func loadJobs(ctx context.Context, db *sql.DB) ([]models.Job, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, job_no, job_type, target_type, COALESCE(target_id, 0), status, payload_json, started_at, finished_at
		FROM platform.operation_job
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.Job, 0)
	for rows.Next() {
		var item models.Job
		var payloadRaw []byte
		var startedAt sql.NullTime
		var finishedAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.JobNo, &item.Type, &item.TargetType, &item.TargetID, &item.Status, &payloadRaw, &startedAt, &finishedAt); err != nil {
			return nil, err
		}
		var payload map[string]string
		_ = json.Unmarshal(payloadRaw, &payload)
		item.Summary = payload["summary"]
		item.StartedAt = formatNullTime(startedAt)
		item.FinishedAt = formatNullTime(finishedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func loadAudits(ctx context.Context, db *sql.DB) ([]models.AuditEvent, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, tenant_id, event_type, target_type, COALESCE(target_id, 0), content_json, created_at
		FROM platform.audit_event
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.AuditEvent, 0)
	for rows.Next() {
		var item models.AuditEvent
		var payloadRaw []byte
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.TenantID, &item.Action, &item.Target, &item.TargetID, &payloadRaw, &createdAt); err != nil {
			return nil, err
		}
		var payload struct {
			Actor    string            `json:"actor"`
			Result   string            `json:"result"`
			Metadata map[string]string `json:"metadata"`
		}
		_ = json.Unmarshal(payloadRaw, &payload)
		item.Actor = payload.Actor
		item.Result = payload.Result
		item.Metadata = payload.Metadata
		item.CreatedAt = formatTime(createdAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func loadRuntimeBindings(ctx context.Context, db *sql.DB) ([]models.RuntimeBinding, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT instance_id, metadata
		FROM platform.runtime_container
		ORDER BY instance_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.RuntimeBinding, 0)
	for rows.Next() {
		var instanceID int
		var metadataRaw []byte
		if err := rows.Scan(&instanceID, &metadataRaw); err != nil {
			return nil, err
		}
		var payload struct {
			ClusterID    string `json:"clusterId"`
			Namespace    string `json:"namespace"`
			WorkloadID   string `json:"workloadId"`
			WorkloadName string `json:"workloadName"`
		}
		_ = json.Unmarshal(metadataRaw, &payload)
		items = append(items, models.RuntimeBinding{
			InstanceID:   instanceID,
			ClusterID:    payload.ClusterID,
			Namespace:    payload.Namespace,
			WorkloadID:   payload.WorkloadID,
			WorkloadName: payload.WorkloadName,
		})
	}
	return items, rows.Err()
}

func upsertUsers(ctx context.Context, tx *sql.Tx, users []models.UserProfile) error {
	for _, user := range users {
		profileJSON, _ := json.Marshal(map[string]any{
			"avatarUrl":  user.AvatarURL,
			"locale":     user.Locale,
			"timezone":   user.Timezone,
			"department": user.Department,
			"title":      user.Title,
			"bio":        user.Bio,
		})
		_, err := tx.ExecContext(ctx, `
			INSERT INTO platform.user_account (
				id, tenant_id, login_name, display_name, email, phone, password_hash, status,
				profile, created_at, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, NOW(), $10)
			ON CONFLICT (id) DO UPDATE
			SET tenant_id = EXCLUDED.tenant_id,
			    login_name = EXCLUDED.login_name,
			    display_name = EXCLUDED.display_name,
			    email = EXCLUDED.email,
			    phone = EXCLUDED.phone,
			    password_hash = EXCLUDED.password_hash,
			    status = EXCLUDED.status,
			    profile = EXCLUDED.profile,
			    updated_at = EXCLUDED.updated_at
		`, user.ID, user.TenantID, user.LoginName, user.DisplayName, user.Email, user.Phone, user.PasswordMasked, user.Status, string(profileJSON), user.UpdatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func replaceAuthIdentities(ctx context.Context, tx *sql.Tx, identities []models.AuthIdentity) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform.auth_identity`); err != nil {
		return err
	}
	for _, identity := range identities {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO platform.auth_identity (
				id, user_id, tenant_id, provider, is_primary, status, status_reason,
				subject, email, open_id, union_id, external_name, last_login_at, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		`, identity.ID, identity.UserID, identity.TenantID, identity.Provider, identity.IsPrimary, identity.Status, identity.StatusReason,
			identity.Subject, identity.Email, identity.OpenID, identity.UnionID, identity.ExternalName, nullTime(identity.LastLoginAt), identity.UpdatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func replaceChannels(ctx context.Context, tx *sql.Tx, channels []models.Channel, activities []models.ChannelActivity) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform.channel_activity`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform.channel`); err != nil {
		return err
	}
	for _, channel := range channels {
		healthJSON, _ := json.Marshal(channel.Health)
		statsJSON, _ := json.Marshal(channel.Stats)
		entryJSON, _ := json.Marshal(channel.EntryPoints)
		settingsJSON, _ := json.Marshal(channel.Settings)
		_, err := tx.ExecContext(ctx, `
			INSERT INTO platform.channel (
				id, code, name, platform, status, connect_method, auth_url, webhook_url, qrcode_url,
				token_masked, callback_secret, health_json, stats_json, entry_points_json, settings_json,
				last_error, notes, updated_at, created_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12::jsonb, $13::jsonb, $14::jsonb, $15::jsonb, $16, $17, $18, $19)
		`, channel.ID, channel.Code, channel.Name, channel.Platform, channel.Status, channel.ConnectMethod, channel.AuthURL, channel.WebhookURL, channel.QrCodeURL,
			channel.TokenMasked, channel.CallbackSecret, string(healthJSON), string(statsJSON), string(entryJSON), string(settingsJSON),
			channel.LastError, channel.Notes, channel.UpdatedAt, channel.CreatedAt)
		if err != nil {
			return err
		}
	}
	for _, activity := range activities {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO platform.channel_activity (id, channel_id, activity_type, title, summary, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, activity.ID, activity.ChannelID, activity.Type, activity.Title, activity.Summary, activity.CreatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func replaceAccountSettings(ctx context.Context, tx *sql.Tx, items []models.AccountSettings) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform.account_settings`); err != nil {
		return err
	}
	for _, item := range items {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO platform.account_settings (
				tenant_id, primary_email, billing_email, alert_email, preferred_locale, secondary_locale,
				timezone, email_verified, marketing_opt_in, notify_on_alert, notify_on_payment, notify_on_expiry,
				notify_channel_email, notify_channel_webhook, notify_channel_in_app, notification_webhook_url,
				portal_headline, portal_subtitle, workspace_callout, experiment_badge, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		`, item.TenantID, item.PrimaryEmail, item.BillingEmail, item.AlertEmail, item.PreferredLocale, item.SecondaryLocale,
			item.Timezone, item.EmailVerified, item.MarketingOptIn, item.NotifyOnAlert, item.NotifyOnPayment, item.NotifyOnExpiry,
			item.NotifyChannelEmail, item.NotifyChannelWebhook, item.NotifyChannelInApp, item.NotificationWebhookURL,
			item.PortalHeadline, item.PortalSubtitle, item.WorkspaceCallout, item.ExperimentBadge, item.UpdatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func replaceWallets(ctx context.Context, tx *sql.Tx, items []models.WalletBalance) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform.wallet_balance`); err != nil {
		return err
	}
	for _, item := range items {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO platform.wallet_balance (
				tenant_id, currency, available_amount, frozen_amount, credit_limit, auto_recharge, last_settlement_at, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, item.TenantID, item.Currency, item.AvailableAmount, item.FrozenAmount, item.CreditLimit, item.AutoRecharge, nullTime(item.LastSettlementAt), item.UpdatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func replaceBillingStatements(ctx context.Context, tx *sql.Tx, items []models.BillingStatement) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform.billing_statement`); err != nil {
		return err
	}
	for _, item := range items {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO platform.billing_statement (
				id, tenant_id, statement_no, billing_month, status, currency, opening_balance,
				charge_amount, refund_amount, closing_balance, paid_amount, due_at, created_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		`, item.ID, item.TenantID, item.StatementNo, item.BillingMonth, item.Status, item.Currency, item.OpeningBalance,
			item.ChargeAmount, item.RefundAmount, item.ClosingBalance, item.PaidAmount, nullTime(item.DueAt), item.CreatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func replaceRuntimes(ctx context.Context, tx *sql.Tx, items []models.InstanceRuntime) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform.instance_runtime_state`); err != nil {
		return err
	}
	for _, item := range items {
		if err := upsertRuntimeState(ctx, tx, &item); err != nil {
			return err
		}
	}
	return nil
}

func replaceCredentials(ctx context.Context, tx *sql.Tx, items []models.InstanceCredential) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform.instance_credential`); err != nil {
		return err
	}
	for _, item := range items {
		if err := upsertCredential(ctx, tx, &item); err != nil {
			return err
		}
	}
	return nil
}

func upsertRuntimeState(ctx context.Context, tx *sql.Tx, item *models.InstanceRuntime) error {
	if item == nil {
		return nil
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO platform.instance_runtime_state (
			instance_id, power_state, cpu_usage_percent, memory_usage_percent, disk_usage_percent,
			api_requests_24h, api_tokens_24h, last_seen_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		ON CONFLICT (instance_id) DO UPDATE
		SET power_state = EXCLUDED.power_state,
		    cpu_usage_percent = EXCLUDED.cpu_usage_percent,
		    memory_usage_percent = EXCLUDED.memory_usage_percent,
		    disk_usage_percent = EXCLUDED.disk_usage_percent,
		    api_requests_24h = EXCLUDED.api_requests_24h,
		    api_tokens_24h = EXCLUDED.api_tokens_24h,
		    last_seen_at = EXCLUDED.last_seen_at,
		    updated_at = NOW()
	`, item.InstanceID, item.PowerState, item.CPUUsagePercent, item.MemoryUsagePercent, item.DiskUsagePercent, item.APIRequests24h, item.APITokens24h, nullTime(item.LastSeenAt))
	return err
}

func upsertCredential(ctx context.Context, tx *sql.Tx, item *models.InstanceCredential) error {
	if item == nil {
		return nil
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO platform.instance_credential (
			instance_id, admin_user, password_masked, last_rotated_at, requires_reset, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (instance_id) DO UPDATE
		SET admin_user = EXCLUDED.admin_user,
		    password_masked = EXCLUDED.password_masked,
		    last_rotated_at = EXCLUDED.last_rotated_at,
		    requires_reset = EXCLUDED.requires_reset,
		    updated_at = NOW()
	`, item.InstanceID, item.AdminUser, item.PasswordMasked, nullTime(item.LastRotatedAt), item.RequiresReset)
	return err
}

func loadUsers(ctx context.Context, db *sql.DB) ([]models.UserProfile, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, tenant_id, login_name, display_name, email, phone, status, profile, password_hash, updated_at
		FROM platform.user_account
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.UserProfile, 0)
	for rows.Next() {
		var item models.UserProfile
		var profileRaw []byte
		var updatedAt time.Time
		if err := rows.Scan(&item.ID, &item.TenantID, &item.LoginName, &item.DisplayName, &item.Email, &item.Phone, &item.Status, &profileRaw, &item.PasswordMasked, &updatedAt); err != nil {
			return nil, err
		}
		var profile map[string]string
		_ = json.Unmarshal(profileRaw, &profile)
		item.AvatarURL = profile["avatarUrl"]
		item.Locale = profile["locale"]
		item.Timezone = profile["timezone"]
		item.Department = profile["department"]
		item.Title = profile["title"]
		item.Bio = profile["bio"]
		item.UpdatedAt = formatTime(updatedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func loadAuthIdentities(ctx context.Context, db *sql.DB) ([]models.AuthIdentity, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, user_id, tenant_id, provider, is_primary, status, COALESCE(status_reason, ''),
		       subject, COALESCE(email, ''), COALESCE(open_id, ''), COALESCE(union_id, ''),
		       COALESCE(external_name, ''), last_login_at, updated_at
		FROM platform.auth_identity
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.AuthIdentity, 0)
	for rows.Next() {
		var item models.AuthIdentity
		var lastLoginAt sql.NullTime
		var updatedAt time.Time
		if err := rows.Scan(&item.ID, &item.UserID, &item.TenantID, &item.Provider, &item.IsPrimary, &item.Status, &item.StatusReason,
			&item.Subject, &item.Email, &item.OpenID, &item.UnionID, &item.ExternalName, &lastLoginAt, &updatedAt); err != nil {
			return nil, err
		}
		item.LastLoginAt = formatNullTime(lastLoginAt)
		item.UpdatedAt = formatTime(updatedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func loadChannels(ctx context.Context, db *sql.DB) ([]models.Channel, []models.ChannelActivity, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, code, name, platform, status, connect_method, COALESCE(auth_url, ''), COALESCE(webhook_url, ''),
		       COALESCE(qrcode_url, ''), COALESCE(token_masked, ''), COALESCE(callback_secret, ''),
		       health_json, stats_json, entry_points_json, settings_json, COALESCE(last_error, ''), COALESCE(notes, ''),
		       updated_at, created_at
		FROM platform.channel
		ORDER BY id
	`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	channels := make([]models.Channel, 0)
	for rows.Next() {
		var item models.Channel
		var healthRaw, statsRaw, entryRaw, settingsRaw []byte
		var updatedAt, createdAt time.Time
		if err := rows.Scan(&item.ID, &item.Code, &item.Name, &item.Platform, &item.Status, &item.ConnectMethod, &item.AuthURL, &item.WebhookURL,
			&item.QrCodeURL, &item.TokenMasked, &item.CallbackSecret, &healthRaw, &statsRaw, &entryRaw, &settingsRaw, &item.LastError, &item.Notes,
			&updatedAt, &createdAt); err != nil {
			return nil, nil, err
		}
		_ = json.Unmarshal(healthRaw, &item.Health)
		_ = json.Unmarshal(statsRaw, &item.Stats)
		_ = json.Unmarshal(entryRaw, &item.EntryPoints)
		if string(settingsRaw) != "" && string(settingsRaw) != "null" {
			_ = json.Unmarshal(settingsRaw, &item.Settings)
		}
		item.UpdatedAt = formatTime(updatedAt)
		item.CreatedAt = formatTime(createdAt)
		channels = append(channels, item)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	activityRows, err := db.QueryContext(ctx, `
		SELECT id, channel_id, activity_type, title, COALESCE(summary, ''), created_at
		FROM platform.channel_activity
		ORDER BY id
	`)
	if err != nil {
		return nil, nil, err
	}
	defer activityRows.Close()

	activities := make([]models.ChannelActivity, 0)
	for activityRows.Next() {
		var item models.ChannelActivity
		var createdAt time.Time
		if err := activityRows.Scan(&item.ID, &item.ChannelID, &item.Type, &item.Title, &item.Summary, &createdAt); err != nil {
			return nil, nil, err
		}
		item.CreatedAt = formatTime(createdAt)
		activities = append(activities, item)
	}
	return channels, activities, activityRows.Err()
}

func loadAccountSettings(ctx context.Context, db *sql.DB) ([]models.AccountSettings, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT tenant_id, COALESCE(primary_email, ''), COALESCE(billing_email, ''), COALESCE(alert_email, ''),
		       preferred_locale, secondary_locale, timezone, email_verified, marketing_opt_in,
		       notify_on_alert, notify_on_payment, notify_on_expiry,
		       notify_channel_email, notify_channel_webhook, notify_channel_in_app,
		       COALESCE(notification_webhook_url, ''), COALESCE(portal_headline, ''), COALESCE(portal_subtitle, ''),
		       COALESCE(workspace_callout, ''), COALESCE(experiment_badge, ''), updated_at
		FROM platform.account_settings
		ORDER BY tenant_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.AccountSettings, 0)
	for rows.Next() {
		var item models.AccountSettings
		var updatedAt time.Time
		if err := rows.Scan(&item.TenantID, &item.PrimaryEmail, &item.BillingEmail, &item.AlertEmail,
			&item.PreferredLocale, &item.SecondaryLocale, &item.Timezone, &item.EmailVerified, &item.MarketingOptIn,
			&item.NotifyOnAlert, &item.NotifyOnPayment, &item.NotifyOnExpiry,
			&item.NotifyChannelEmail, &item.NotifyChannelWebhook, &item.NotifyChannelInApp,
			&item.NotificationWebhookURL, &item.PortalHeadline, &item.PortalSubtitle,
			&item.WorkspaceCallout, &item.ExperimentBadge, &updatedAt); err != nil {
			return nil, err
		}
		item.UpdatedAt = formatTime(updatedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func loadWallets(ctx context.Context, db *sql.DB) ([]models.WalletBalance, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT tenant_id, currency, available_amount, frozen_amount, credit_limit, auto_recharge, last_settlement_at, updated_at
		FROM platform.wallet_balance
		ORDER BY tenant_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.WalletBalance, 0)
	for rows.Next() {
		var item models.WalletBalance
		var lastSettlementAt sql.NullTime
		var updatedAt time.Time
		if err := rows.Scan(&item.TenantID, &item.Currency, &item.AvailableAmount, &item.FrozenAmount, &item.CreditLimit, &item.AutoRecharge, &lastSettlementAt, &updatedAt); err != nil {
			return nil, err
		}
		item.LastSettlementAt = formatNullTime(lastSettlementAt)
		item.UpdatedAt = formatTime(updatedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func loadBillingStatements(ctx context.Context, db *sql.DB) ([]models.BillingStatement, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, tenant_id, statement_no, billing_month, status, currency, opening_balance,
		       charge_amount, refund_amount, closing_balance, paid_amount, due_at, created_at
		FROM platform.billing_statement
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.BillingStatement, 0)
	for rows.Next() {
		var item models.BillingStatement
		var dueAt sql.NullTime
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.TenantID, &item.StatementNo, &item.BillingMonth, &item.Status, &item.Currency, &item.OpeningBalance,
			&item.ChargeAmount, &item.RefundAmount, &item.ClosingBalance, &item.PaidAmount, &dueAt, &createdAt); err != nil {
			return nil, err
		}
		item.DueAt = formatNullTime(dueAt)
		item.CreatedAt = formatTime(createdAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func loadRuntimes(ctx context.Context, db *sql.DB) ([]models.InstanceRuntime, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT instance_id, power_state, cpu_usage_percent, memory_usage_percent, disk_usage_percent,
		       api_requests_24h, api_tokens_24h, last_seen_at
		FROM platform.instance_runtime_state
		ORDER BY instance_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.InstanceRuntime, 0)
	for rows.Next() {
		var item models.InstanceRuntime
		var lastSeenAt sql.NullTime
		if err := rows.Scan(&item.InstanceID, &item.PowerState, &item.CPUUsagePercent, &item.MemoryUsagePercent, &item.DiskUsagePercent, &item.APIRequests24h, &item.APITokens24h, &lastSeenAt); err != nil {
			return nil, err
		}
		item.LastSeenAt = formatNullTime(lastSeenAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func loadCredentials(ctx context.Context, db *sql.DB) ([]models.InstanceCredential, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT instance_id, admin_user, password_masked, last_rotated_at, requires_reset
		FROM platform.instance_credential
		ORDER BY instance_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.InstanceCredential, 0)
	for rows.Next() {
		var item models.InstanceCredential
		var lastRotatedAt sql.NullTime
		if err := rows.Scan(&item.InstanceID, &item.AdminUser, &item.PasswordMasked, &lastRotatedAt, &item.RequiresReset); err != nil {
			return nil, err
		}
		item.LastRotatedAt = formatNullTime(lastRotatedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func replaceAlerts(ctx context.Context, tx *sql.Tx, items []models.Alert) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform.alert_record`); err != nil {
		return err
	}
	for _, item := range items {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO platform.alert_record (
				id, tenant_id, instance_id, metric_key, severity, status, summary, detail, triggered_at, created_at
			)
			VALUES ($1, (SELECT tenant_id FROM platform.service_instance WHERE id = $2), $2, $3, $4, $5, $6, '{}'::jsonb, $7, $7)
		`, item.ID, item.InstanceID, item.MetricKey, item.Severity, item.Status, item.Summary, nullTime(item.TriggeredAt))
		if err != nil {
			return err
		}
	}
	return nil
}

func replaceTickets(ctx context.Context, tx *sql.Tx, items []models.Ticket) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform.ticket_record`); err != nil {
		return err
	}
	for _, item := range items {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO platform.ticket_record (
				id, ticket_no, tenant_id, instance_id, title, category, severity, status,
				reporter, assignee, description, created_at, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NULLIF($10, ''), $11, $12, $13)
		`, item.ID, item.TicketNo, item.TenantID, nullableInt(item.InstanceID), item.Title, item.Category, item.Severity, item.Status,
			item.Reporter, item.Assignee, item.Description, nullTime(item.CreatedAt), nullTime(item.UpdatedAt))
		if err != nil {
			return err
		}
	}
	return nil
}

func replaceApprovals(ctx context.Context, tx *sql.Tx, approvals []models.ApprovalRecord, actions []models.ApprovalAction) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform.approval_action`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform.approval_record`); err != nil {
		return err
	}

	for _, item := range approvals {
		metadata := cloneStringMap(item.Metadata)
		if item.ApplicantName != "" {
			metadata["applicantName"] = item.ApplicantName
		}
		if item.ApproverName != "" {
			metadata["approverName"] = item.ApproverName
		}
		if item.ExecutorName != "" {
			metadata["executorName"] = item.ExecutorName
		}
		metadataJSON, _ := json.Marshal(metadata)

		_, err := tx.ExecContext(ctx, `
			INSERT INTO platform.approval_record (
				id, approval_no, tenant_id, instance_id, approval_type, target_type, target_id,
				applicant_id, approver_id, executor_id, status, risk_level, reason, approval_comment,
				reject_reason, approved_at, executed_at, expired_at, metadata_json, created_at, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NULLIF($13, ''), NULLIF($14, ''),
			        NULLIF($15, ''), $16, $17, $18, $19::jsonb, $20, $21)
		`, item.ID, item.ApprovalNo, nullableInt(item.TenantID), nullableInt(item.InstanceID), item.ApprovalType, item.TargetType,
			nullableInt(item.TargetID), item.ApplicantID, nullableInt(item.ApproverID), nullableInt(item.ExecutorID), item.Status,
			defaultStringValue(item.RiskLevel, "high"), item.Reason, item.ApprovalComment, item.RejectReason, nullTime(item.ApprovedAt),
			nullTime(item.ExecutedAt), nullTime(item.ExpiredAt), string(metadataJSON), nullTime(item.CreatedAt), nullTime(item.UpdatedAt))
		if err != nil {
			return err
		}
	}

	for _, item := range actions {
		metadataJSON, _ := json.Marshal(item.Metadata)
		_, err := tx.ExecContext(ctx, `
			INSERT INTO platform.approval_action (
				id, approval_id, actor_id, actor_name, action, comment, metadata_json, created_at
			)
			VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), $7::jsonb, $8)
		`, item.ID, item.ApprovalID, nullableInt(item.ActorID), item.ActorName, item.Action, item.Comment, string(metadataJSON), nullTime(item.CreatedAt))
		if err != nil {
			return err
		}
	}
	return nil
}

func upsertDiagnosticSessionTx(ctx context.Context, tx *sql.Tx, item models.DiagnosticSession) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO platform.diagnostic_session (
			id, session_no, tenant_id, instance_id, cluster_id, namespace, workload_id, workload_name,
			pod_name, container_name, access_mode, status, approval_ticket, approved_by, operator,
			operator_user_id, reason, close_reason, expires_at, last_command_at, started_at, ended_at,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''), NULLIF($8, ''),
		        $9, NULLIF($10, ''), $11, $12, NULLIF($13, ''), NULLIF($14, ''), $15, $16, NULLIF($17, ''),
		        NULLIF($18, ''), $19, $20, $21, $22, $23, $24)
		ON CONFLICT (id) DO UPDATE
		SET session_no = EXCLUDED.session_no,
		    tenant_id = EXCLUDED.tenant_id,
		    instance_id = EXCLUDED.instance_id,
		    cluster_id = EXCLUDED.cluster_id,
		    namespace = EXCLUDED.namespace,
		    workload_id = EXCLUDED.workload_id,
		    workload_name = EXCLUDED.workload_name,
		    pod_name = EXCLUDED.pod_name,
		    container_name = EXCLUDED.container_name,
		    access_mode = EXCLUDED.access_mode,
		    status = EXCLUDED.status,
		    approval_ticket = EXCLUDED.approval_ticket,
		    approved_by = EXCLUDED.approved_by,
		    operator = EXCLUDED.operator,
		    operator_user_id = EXCLUDED.operator_user_id,
		    reason = EXCLUDED.reason,
		    close_reason = EXCLUDED.close_reason,
		    expires_at = EXCLUDED.expires_at,
		    last_command_at = EXCLUDED.last_command_at,
		    started_at = EXCLUDED.started_at,
		    ended_at = EXCLUDED.ended_at,
		    created_at = EXCLUDED.created_at,
		    updated_at = EXCLUDED.updated_at
	`, item.ID, item.SessionNo, item.TenantID, item.InstanceID, item.ClusterID, item.Namespace, item.WorkloadID, item.WorkloadName,
		item.PodName, item.ContainerName, item.AccessMode, item.Status, item.ApprovalTicket, item.ApprovedBy, item.Operator,
		nullableInt(item.OperatorUserID), item.Reason, item.CloseReason, nullTime(item.ExpiresAt), nullTime(item.LastCommandAt),
		nullTime(item.StartedAt), nullTime(item.EndedAt), nullTime(item.CreatedAt), nullTime(item.UpdatedAt))
	return err
}

func upsertDiagnosticCommandRecordTx(ctx context.Context, tx *sql.Tx, item models.DiagnosticCommandRecord) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO platform.diagnostic_command_record (
			id, session_id, tenant_id, instance_id, command_key, command_text, status, exit_code,
			duration_ms, output, error_output, output_truncated, executed_at
		)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6, $7, $8, $9, NULLIF($10, ''), NULLIF($11, ''), $12, $13)
		ON CONFLICT (id) DO UPDATE
		SET session_id = EXCLUDED.session_id,
		    tenant_id = EXCLUDED.tenant_id,
		    instance_id = EXCLUDED.instance_id,
		    command_key = EXCLUDED.command_key,
		    command_text = EXCLUDED.command_text,
		    status = EXCLUDED.status,
		    exit_code = EXCLUDED.exit_code,
		    duration_ms = EXCLUDED.duration_ms,
		    output = EXCLUDED.output,
		    error_output = EXCLUDED.error_output,
		    output_truncated = EXCLUDED.output_truncated,
		    executed_at = EXCLUDED.executed_at
	`, item.ID, item.SessionID, item.TenantID, item.InstanceID, item.CommandKey, item.CommandText, item.Status,
		item.ExitCode, item.DurationMs, item.Output, item.ErrorOutput, item.OutputTruncated, nullTime(item.ExecutedAt))
	return err
}

func replaceDiagnostics(ctx context.Context, tx *sql.Tx, sessions []models.DiagnosticSession, commands []models.DiagnosticCommandRecord) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform.diagnostic_command_record`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform.diagnostic_session`); err != nil {
		return err
	}
	for _, item := range sessions {
		if err := upsertDiagnosticSessionTx(ctx, tx, item); err != nil {
			return err
		}
	}
	for _, item := range commands {
		if err := upsertDiagnosticCommandRecordTx(ctx, tx, item); err != nil {
			return err
		}
	}
	return nil
}

func upsertWorkspaceSessionTx(ctx context.Context, tx *sql.Tx, item models.WorkspaceSession) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO platform.workspace_session (
			id, session_no, tenant_id, instance_id, title, status, workspace_url,
			protocol_version, last_opened_at, last_artifact_at, last_synced_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''), $8, $9, $10, $11, $12, $13)
		ON CONFLICT (id) DO UPDATE
		SET session_no = EXCLUDED.session_no,
		    tenant_id = EXCLUDED.tenant_id,
		    instance_id = EXCLUDED.instance_id,
		    title = EXCLUDED.title,
		    status = EXCLUDED.status,
		    workspace_url = EXCLUDED.workspace_url,
		    protocol_version = EXCLUDED.protocol_version,
		    last_opened_at = EXCLUDED.last_opened_at,
		    last_artifact_at = EXCLUDED.last_artifact_at,
		    last_synced_at = EXCLUDED.last_synced_at,
		    created_at = EXCLUDED.created_at,
		    updated_at = EXCLUDED.updated_at
	`, item.ID, item.SessionNo, item.TenantID, item.InstanceID, item.Title, item.Status, item.WorkspaceURL,
		defaultStringValue(item.ProtocolVersion, "openclaw-lobster-bridge/v2"), nullTime(item.LastOpenedAt), nullTime(item.LastArtifactAt), nullTime(item.LastSyncedAt), nullTime(item.CreatedAt), nullTime(item.UpdatedAt))
	return err
}

func upsertWorkspaceArtifactTx(ctx context.Context, tx *sql.Tx, item models.WorkspaceArtifact) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO platform.workspace_artifact (
			id, session_id, tenant_id, instance_id, message_id, title, kind, external_id, origin,
			source_url, preview_url, archive_status, content_type, size_bytes, storage_bucket,
			storage_key, filename, checksum_sha256, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''), $9, $10, NULLIF($11, ''), $12, NULLIF($13, ''), $14, NULLIF($15, ''), NULLIF($16, ''), NULLIF($17, ''), NULLIF($18, ''), $19, $20)
		ON CONFLICT (id) DO UPDATE
		SET session_id = EXCLUDED.session_id,
		    tenant_id = EXCLUDED.tenant_id,
		    instance_id = EXCLUDED.instance_id,
		    message_id = EXCLUDED.message_id,
		    title = EXCLUDED.title,
		    kind = EXCLUDED.kind,
		    external_id = EXCLUDED.external_id,
		    origin = EXCLUDED.origin,
		    source_url = EXCLUDED.source_url,
		    preview_url = EXCLUDED.preview_url,
		    archive_status = EXCLUDED.archive_status,
		    content_type = EXCLUDED.content_type,
		    size_bytes = EXCLUDED.size_bytes,
		    storage_bucket = EXCLUDED.storage_bucket,
		    storage_key = EXCLUDED.storage_key,
		    filename = EXCLUDED.filename,
		    checksum_sha256 = EXCLUDED.checksum_sha256,
		    created_at = EXCLUDED.created_at,
		    updated_at = EXCLUDED.updated_at
	`, item.ID, item.SessionID, item.TenantID, item.InstanceID, nullableInt(item.MessageID), item.Title, item.Kind, item.ExternalID, defaultStringValue(item.Origin, "manual"),
		item.SourceURL, item.PreviewURL, defaultStringValue(item.ArchiveStatus, "pending"), item.ContentType, item.SizeBytes, item.StorageBucket,
		item.StorageKey, item.Filename, item.ChecksumSHA256, nullTime(item.CreatedAt), nullTime(item.UpdatedAt))
	return err
}

func upsertWorkspaceMessageTx(ctx context.Context, tx *sql.Tx, item models.WorkspaceMessage) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO platform.workspace_message (
			id, session_id, tenant_id, instance_id, parent_message_id, role, status, external_id, origin, trace_id,
			error_code, error_message, delivery_attempt, content, delivered_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''), $9, NULLIF($10, ''), NULLIF($11, ''), NULLIF($12, ''), $13, $14, $15, $16, $17)
		ON CONFLICT (id) DO UPDATE
		SET session_id = EXCLUDED.session_id,
		    tenant_id = EXCLUDED.tenant_id,
		    instance_id = EXCLUDED.instance_id,
		    parent_message_id = EXCLUDED.parent_message_id,
		    role = EXCLUDED.role,
		    status = EXCLUDED.status,
		    external_id = EXCLUDED.external_id,
		    origin = EXCLUDED.origin,
		    trace_id = EXCLUDED.trace_id,
		    error_code = EXCLUDED.error_code,
		    error_message = EXCLUDED.error_message,
		    delivery_attempt = EXCLUDED.delivery_attempt,
		    content = EXCLUDED.content,
		    delivered_at = EXCLUDED.delivered_at,
		    created_at = EXCLUDED.created_at,
		    updated_at = EXCLUDED.updated_at
	`, item.ID, item.SessionID, item.TenantID, item.InstanceID, nullableInt(item.ParentMessageID), item.Role, item.Status, item.ExternalID, defaultStringValue(item.Origin, "platform"), item.TraceID,
		item.ErrorCode, item.ErrorMessage, item.DeliveryAttempt, item.Content, nullTime(item.DeliveredAt), nullTime(item.CreatedAt), nullTime(item.UpdatedAt))
	return err
}

func upsertWorkspaceMessageEventTx(ctx context.Context, tx *sql.Tx, item models.WorkspaceMessageEvent) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO platform.workspace_message_event (
			id, session_id, message_id, tenant_id, instance_id, event_type, external_id, origin, trace_id, payload_json, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''), NULLIF($8, ''), NULLIF($9, ''), COALESCE(NULLIF($10, ''), '{}')::jsonb, $11)
		ON CONFLICT (id) DO UPDATE
		SET session_id = EXCLUDED.session_id,
		    message_id = EXCLUDED.message_id,
		    tenant_id = EXCLUDED.tenant_id,
		    instance_id = EXCLUDED.instance_id,
		    event_type = EXCLUDED.event_type,
		    external_id = EXCLUDED.external_id,
		    origin = EXCLUDED.origin,
		    trace_id = EXCLUDED.trace_id,
		    payload_json = EXCLUDED.payload_json,
		    created_at = EXCLUDED.created_at
	`, item.ID, item.SessionID, nullableInt(item.MessageID), item.TenantID, item.InstanceID, item.EventType, item.ExternalID, defaultStringValue(item.Origin, "platform"), item.TraceID, item.PayloadJSON, nullTime(item.CreatedAt))
	return err
}

func upsertWorkspaceArtifactAccessLogTx(ctx context.Context, tx *sql.Tx, item models.WorkspaceArtifactAccessLog) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO platform.workspace_artifact_access_log (
			id, artifact_id, session_id, tenant_id, instance_id, action, scope, actor,
			remote_addr, user_agent, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NULLIF($9, ''), NULLIF($10, ''), $11)
		ON CONFLICT (id) DO UPDATE
		SET artifact_id = EXCLUDED.artifact_id,
		    session_id = EXCLUDED.session_id,
		    tenant_id = EXCLUDED.tenant_id,
		    instance_id = EXCLUDED.instance_id,
		    action = EXCLUDED.action,
		    scope = EXCLUDED.scope,
		    actor = EXCLUDED.actor,
		    remote_addr = EXCLUDED.remote_addr,
		    user_agent = EXCLUDED.user_agent,
		    created_at = EXCLUDED.created_at
	`, item.ID, item.ArtifactID, item.SessionID, item.TenantID, item.InstanceID, item.Action, item.Scope, item.Actor,
		item.RemoteAddr, item.UserAgent, nullTime(item.CreatedAt))
	return err
}

func upsertWorkspaceArtifactFavoriteTx(ctx context.Context, tx *sql.Tx, item models.WorkspaceArtifactFavorite) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO platform.workspace_artifact_favorite (
			id, artifact_id, session_id, tenant_id, instance_id, user_id, actor, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE
		SET artifact_id = EXCLUDED.artifact_id,
		    session_id = EXCLUDED.session_id,
		    tenant_id = EXCLUDED.tenant_id,
		    instance_id = EXCLUDED.instance_id,
		    user_id = EXCLUDED.user_id,
		    actor = EXCLUDED.actor,
		    created_at = EXCLUDED.created_at
	`, item.ID, item.ArtifactID, item.SessionID, item.TenantID, item.InstanceID, nullableInt(item.UserID), item.Actor, nullTime(item.CreatedAt))
	return err
}

func upsertWorkspaceArtifactShareTx(ctx context.Context, tx *sql.Tx, item models.WorkspaceArtifactShare) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO platform.workspace_artifact_share (
			id, artifact_id, session_id, tenant_id, instance_id, scope, token, note, created_by,
			created_by_user_id, use_count, expires_at, last_opened_at, revoked_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''), $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (id) DO UPDATE
		SET artifact_id = EXCLUDED.artifact_id,
		    session_id = EXCLUDED.session_id,
		    tenant_id = EXCLUDED.tenant_id,
		    instance_id = EXCLUDED.instance_id,
		    scope = EXCLUDED.scope,
		    token = EXCLUDED.token,
		    note = EXCLUDED.note,
		    created_by = EXCLUDED.created_by,
		    created_by_user_id = EXCLUDED.created_by_user_id,
		    use_count = EXCLUDED.use_count,
		    expires_at = EXCLUDED.expires_at,
		    last_opened_at = EXCLUDED.last_opened_at,
		    revoked_at = EXCLUDED.revoked_at,
		    created_at = EXCLUDED.created_at,
		    updated_at = EXCLUDED.updated_at
	`, item.ID, item.ArtifactID, item.SessionID, item.TenantID, item.InstanceID, item.Scope, item.Token, item.Note, item.CreatedBy,
		nullableInt(item.CreatedByUserID), item.UseCount, nullTime(item.ExpiresAt), nullTime(item.LastOpenedAt), nullTime(item.RevokedAt), nullTime(item.CreatedAt), nullTime(item.UpdatedAt))
	return err
}

func replaceWorkspace(
	ctx context.Context,
	tx *sql.Tx,
	sessions []models.WorkspaceSession,
	artifacts []models.WorkspaceArtifact,
	messages []models.WorkspaceMessage,
	events []models.WorkspaceMessageEvent,
	logs []models.WorkspaceArtifactAccessLog,
	favorites []models.WorkspaceArtifactFavorite,
	shares []models.WorkspaceArtifactShare,
) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform.workspace_artifact_share`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform.workspace_artifact_favorite`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform.workspace_artifact_access_log`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform.workspace_message_event`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform.workspace_message`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform.workspace_artifact`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform.workspace_session`); err != nil {
		return err
	}

	for _, item := range sessions {
		if err := upsertWorkspaceSessionTx(ctx, tx, item); err != nil {
			return err
		}
	}

	for _, item := range artifacts {
		if err := upsertWorkspaceArtifactTx(ctx, tx, item); err != nil {
			return err
		}
	}

	for _, item := range messages {
		if err := upsertWorkspaceMessageTx(ctx, tx, item); err != nil {
			return err
		}
	}

	for _, item := range events {
		if err := upsertWorkspaceMessageEventTx(ctx, tx, item); err != nil {
			return err
		}
	}

	for _, item := range logs {
		if err := upsertWorkspaceArtifactAccessLogTx(ctx, tx, item); err != nil {
			return err
		}
	}
	for _, item := range favorites {
		if err := upsertWorkspaceArtifactFavoriteTx(ctx, tx, item); err != nil {
			return err
		}
	}
	for _, item := range shares {
		if err := upsertWorkspaceArtifactShareTx(ctx, tx, item); err != nil {
			return err
		}
	}

	return nil
}

func replaceOEM(ctx context.Context, tx *sql.Tx, brands []models.OEMBrand, themes []models.OEMTheme, features []models.OEMFeatureFlags, bindings []models.TenantBrandBinding) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform.tenant_brand_binding`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform.oem_feature_flags`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform.oem_theme`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform.oem_brand`); err != nil {
		return err
	}
	for _, item := range brands {
		domainsJSON, _ := json.Marshal(item.Domains)
		_, err := tx.ExecContext(ctx, `
			INSERT INTO platform.oem_brand (
				id, code, name, status, logo_url, favicon_url, support_email, support_url, domains_json, updated_at, created_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, $10, $11)
		`, item.ID, item.Code, item.Name, item.Status, item.LogoURL, item.FaviconURL, item.SupportEmail, item.SupportURL, string(domainsJSON), nullTime(item.UpdatedAt), nullTime(item.CreatedAt))
		if err != nil {
			return err
		}
	}
	for _, item := range themes {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO platform.oem_theme (
				brand_id, primary_color, secondary_color, accent_color, surface_mode, font_family, radius
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, item.BrandID, item.PrimaryColor, item.SecondaryColor, item.AccentColor, item.SurfaceMode, item.FontFamily, item.Radius)
		if err != nil {
			return err
		}
	}
	for _, item := range features {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO platform.oem_feature_flags (
				brand_id, portal_enabled, admin_enabled, channels_enabled, tickets_enabled, purchase_enabled,
				runtime_control_enabled, audit_enabled, sso_enabled
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, item.BrandID, item.PortalEnabled, item.AdminEnabled, item.ChannelsEnabled, item.TicketsEnabled, item.PurchaseEnabled,
			item.RuntimeControlEnabled, item.AuditEnabled, item.SSOEnabled)
		if err != nil {
			return err
		}
	}
	for _, item := range bindings {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO platform.tenant_brand_binding (tenant_id, brand_id, binding_mode, updated_at)
			VALUES ($1, $2, $3, $4)
		`, item.TenantID, item.BrandID, item.BindingMode, nullTime(item.UpdatedAt))
		if err != nil {
			return err
		}
	}
	return nil
}

func replaceCommerce(ctx context.Context, tx *sql.Tx, orders []models.Order, subscriptions []models.Subscription, payments []models.PaymentTransaction, refunds []models.RefundRecord, invoices []models.InvoiceRecord, callbacks []models.PaymentCallbackEvent) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform.payment_callback_event`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform.invoice_record`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform.refund_record`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform.payment_transaction`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform.subscription`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform.order_item`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform.order_main`); err != nil {
		return err
	}

	for _, item := range orders {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO platform.order_main (
				id, tenant_id, order_no, source_platform, order_type, status, currency, total_amount,
				payable_amount, paid_at, created_at, updated_at, instance_id, plan_code
			)
			VALUES ($1, $2, $3, 'portal', $4, $5, $6, $7, $7, NULL, $8, $9, $10, $11)
		`, item.ID, item.TenantID, item.OrderNo, item.Action, item.Status, item.Currency, item.Amount, nullTime(item.CreatedAt), nullTime(item.UpdatedAt), nullableInt(item.InstanceID), item.PlanCode)
		if err != nil {
			return err
		}
	}

	for _, item := range subscriptions {
		planID, err := lookupPlanID(ctx, tx, item.PlanCode)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO platform.subscription (
				id, subscription_no, tenant_id, product_id, plan_id, instance_id, status, renew_mode,
				current_period_start, current_period_end, expired_at, metadata, created_at, updated_at,
				product_code, plan_code
			)
			VALUES ($1, $2, $3, 1, $4, $5, $6, $7, $8, $9, $10, '{}'::jsonb, $11, $12, $13, $14)
		`, item.ID, item.SubscriptionNo, item.TenantID, planID, nullableInt(item.InstanceID), item.Status, item.RenewMode,
			nullTime(item.CurrentPeriodStart), nullTime(item.CurrentPeriodEnd), nullTime(item.ExpiredAt),
			nullTime(item.CreatedAt), nullTime(item.UpdatedAt), item.ProductCode, item.PlanCode)
		if err != nil {
			return err
		}
	}

	for _, item := range payments {
		rawJSON, _ := json.Marshal(item.Raw)
		_, err := tx.ExecContext(ctx, `
			INSERT INTO platform.payment_transaction (
				id, order_id, channel, trade_no, channel_order_no, status, amount, request_payload, callback_payload,
				paid_at, created_at, updated_at, currency, pay_mode, pay_url, code_url, prepay_id, app_id, mch_id, raw_json
			)
			VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6, $7, '{}'::jsonb, '{}'::jsonb, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18::jsonb)
		`, item.ID, item.OrderID, item.Channel, item.TradeNo, item.ChannelOrderNo, item.Status, item.Amount,
			nullTime(item.PaidAt), nullTime(item.CreatedAt), nullTime(item.UpdatedAt), item.Currency, item.PayMode, item.PayURL, item.CodeURL, item.PrepayID, item.AppID, item.MchID, string(rawJSON))
		if err != nil {
			return err
		}
	}

	for _, item := range refunds {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO platform.refund_record (
				id, refund_no, order_id, payment_id, status, amount, reason, channel_refund_no, created_at, updated_at, notify_url
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''), $9, $10, $11)
		`, item.ID, item.RefundNo, item.OrderID, item.PaymentID, item.Status, item.Amount, item.Reason, item.ChannelRefundNo, nullTime(item.CreatedAt), nullTime(item.UpdatedAt), item.NotifyURL)
		if err != nil {
			return err
		}
	}

	for _, item := range invoices {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO platform.invoice_record (
				id, tenant_id, order_id, invoice_type, status, amount, title, tax_no, metadata, created_at, updated_at, email, invoice_no, pdf_url
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''), '{}'::jsonb, $9, $10, NULLIF($11, ''), NULLIF($12, ''), NULLIF($13, ''))
		`, item.ID, item.TenantID, item.OrderID, item.InvoiceType, item.Status, item.Amount, item.Title, item.TaxNo, nullTime(item.CreatedAt), nullTime(item.UpdatedAt), item.Email, item.InvoiceNo, item.PDFURL)
		if err != nil {
			return err
		}
	}

	for _, item := range callbacks {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO platform.payment_callback_event (
				id, channel, event_type, out_trade_no, out_refund_no, signature_status, decrypt_status,
				process_status, request_serial, raw_body, created_at
			)
			VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''), $6, $7, $8, NULLIF($9, ''), $10, $11)
		`, item.ID, item.Channel, item.EventType, item.OutTradeNo, item.OutRefundNo, item.SignatureStatus, item.DecryptStatus, item.ProcessStatus, item.RequestSerial, item.RawBody, nullTime(item.CreatedAt))
		if err != nil {
			return err
		}
	}
	return nil
}

func loadAlerts(ctx context.Context, db *sql.DB) ([]models.Alert, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, instance_id, severity, status, metric_key, summary, triggered_at
		FROM platform.alert_record
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.Alert, 0)
	for rows.Next() {
		var item models.Alert
		var triggeredAt time.Time
		if err := rows.Scan(&item.ID, &item.InstanceID, &item.Severity, &item.Status, &item.MetricKey, &item.Summary, &triggeredAt); err != nil {
			return nil, err
		}
		item.TriggeredAt = formatTime(triggeredAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func loadTickets(ctx context.Context, db *sql.DB) ([]models.Ticket, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, ticket_no, tenant_id, COALESCE(instance_id, 0), title, category, severity, status,
		       reporter, COALESCE(assignee, ''), COALESCE(description, ''), created_at, updated_at
		FROM platform.ticket_record
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.Ticket, 0)
	for rows.Next() {
		var item models.Ticket
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&item.ID, &item.TicketNo, &item.TenantID, &item.InstanceID, &item.Title, &item.Category, &item.Severity, &item.Status,
			&item.Reporter, &item.Assignee, &item.Description, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		item.CreatedAt = formatTime(createdAt)
		item.UpdatedAt = formatTime(updatedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func loadApprovals(ctx context.Context, db *sql.DB) ([]models.ApprovalRecord, []models.ApprovalAction, error) {
	approvalRows, err := db.QueryContext(ctx, `
		SELECT id, approval_no, COALESCE(tenant_id, 0), COALESCE(instance_id, 0), approval_type, target_type,
		       COALESCE(target_id, 0), applicant_id, COALESCE(approver_id, 0), COALESCE(executor_id, 0),
		       status, COALESCE(risk_level, 'high'), COALESCE(reason, ''), COALESCE(approval_comment, ''),
		       COALESCE(reject_reason, ''), approved_at, executed_at, expired_at,
		       COALESCE(metadata_json, '{}'::jsonb), created_at, updated_at
		FROM platform.approval_record
		ORDER BY created_at DESC, id DESC
	`)
	if err != nil {
		return nil, nil, err
	}
	defer approvalRows.Close()

	approvals := make([]models.ApprovalRecord, 0)
	for approvalRows.Next() {
		var item models.ApprovalRecord
		var approvedAt, executedAt, expiredAt sql.NullTime
		var metadataRaw []byte
		var createdAt, updatedAt time.Time
		if err := approvalRows.Scan(&item.ID, &item.ApprovalNo, &item.TenantID, &item.InstanceID, &item.ApprovalType, &item.TargetType,
			&item.TargetID, &item.ApplicantID, &item.ApproverID, &item.ExecutorID, &item.Status, &item.RiskLevel, &item.Reason,
			&item.ApprovalComment, &item.RejectReason, &approvedAt, &executedAt, &expiredAt, &metadataRaw, &createdAt, &updatedAt); err != nil {
			return nil, nil, err
		}
		_ = json.Unmarshal(metadataRaw, &item.Metadata)
		if item.Metadata == nil {
			item.Metadata = map[string]string{}
		}
		item.ApplicantName = item.Metadata["applicantName"]
		item.ApproverName = item.Metadata["approverName"]
		item.ExecutorName = item.Metadata["executorName"]
		item.ApprovedAt = formatNullTime(approvedAt)
		item.ExecutedAt = formatNullTime(executedAt)
		item.ExpiredAt = formatNullTime(expiredAt)
		item.CreatedAt = formatTime(createdAt)
		item.UpdatedAt = formatTime(updatedAt)
		approvals = append(approvals, item)
	}
	if err := approvalRows.Err(); err != nil {
		return nil, nil, err
	}

	actionRows, err := db.QueryContext(ctx, `
		SELECT id, approval_id, COALESCE(actor_id, 0), actor_name, action, COALESCE(comment, ''),
		       COALESCE(metadata_json, '{}'::jsonb), created_at
		FROM platform.approval_action
		ORDER BY created_at ASC, id ASC
	`)
	if err != nil {
		return nil, nil, err
	}
	defer actionRows.Close()

	actions := make([]models.ApprovalAction, 0)
	for actionRows.Next() {
		var item models.ApprovalAction
		var metadataRaw []byte
		var createdAt time.Time
		if err := actionRows.Scan(&item.ID, &item.ApprovalID, &item.ActorID, &item.ActorName, &item.Action, &item.Comment, &metadataRaw, &createdAt); err != nil {
			return nil, nil, err
		}
		_ = json.Unmarshal(metadataRaw, &item.Metadata)
		if item.Metadata == nil {
			item.Metadata = map[string]string{}
		}
		item.CreatedAt = formatTime(createdAt)
		actions = append(actions, item)
	}
	if err := actionRows.Err(); err != nil {
		return nil, nil, err
	}

	return approvals, actions, nil
}

func loadDiagnostics(ctx context.Context, db *sql.DB) ([]models.DiagnosticSession, []models.DiagnosticCommandRecord, error) {
	sessionRows, err := db.QueryContext(ctx, `
		SELECT id, session_no, tenant_id, instance_id, COALESCE(cluster_id, ''), COALESCE(namespace, ''),
		       COALESCE(workload_id, ''), COALESCE(workload_name, ''), pod_name, COALESCE(container_name, ''),
		       access_mode, status, COALESCE(approval_ticket, ''), COALESCE(approved_by, ''), operator,
		       COALESCE(operator_user_id, 0), COALESCE(reason, ''), COALESCE(close_reason, ''), expires_at,
		       last_command_at, started_at, ended_at, created_at, updated_at
		FROM platform.diagnostic_session
		ORDER BY updated_at DESC, id DESC
	`)
	if err != nil {
		return nil, nil, err
	}
	defer sessionRows.Close()

	sessions := make([]models.DiagnosticSession, 0)
	for sessionRows.Next() {
		var item models.DiagnosticSession
		var expiresAt, lastCommandAt, startedAt, endedAt sql.NullTime
		var createdAt, updatedAt time.Time
		if err := sessionRows.Scan(&item.ID, &item.SessionNo, &item.TenantID, &item.InstanceID, &item.ClusterID, &item.Namespace,
			&item.WorkloadID, &item.WorkloadName, &item.PodName, &item.ContainerName, &item.AccessMode, &item.Status,
			&item.ApprovalTicket, &item.ApprovedBy, &item.Operator, &item.OperatorUserID, &item.Reason, &item.CloseReason,
			&expiresAt, &lastCommandAt, &startedAt, &endedAt, &createdAt, &updatedAt); err != nil {
			return nil, nil, err
		}
		item.ExpiresAt = formatNullTime(expiresAt)
		item.LastCommandAt = formatNullTime(lastCommandAt)
		item.StartedAt = formatNullTime(startedAt)
		item.EndedAt = formatNullTime(endedAt)
		item.CreatedAt = formatTime(createdAt)
		item.UpdatedAt = formatTime(updatedAt)
		sessions = append(sessions, item)
	}
	if err := sessionRows.Err(); err != nil {
		return nil, nil, err
	}

	commandRows, err := db.QueryContext(ctx, `
		SELECT id, session_id, tenant_id, instance_id, COALESCE(command_key, ''), command_text, status, exit_code,
		       duration_ms, COALESCE(output, ''), COALESCE(error_output, ''), output_truncated, executed_at
		FROM platform.diagnostic_command_record
		ORDER BY executed_at ASC, id ASC
	`)
	if err != nil {
		return nil, nil, err
	}
	defer commandRows.Close()

	commands := make([]models.DiagnosticCommandRecord, 0)
	for commandRows.Next() {
		var item models.DiagnosticCommandRecord
		var executedAt time.Time
		if err := commandRows.Scan(&item.ID, &item.SessionID, &item.TenantID, &item.InstanceID, &item.CommandKey,
			&item.CommandText, &item.Status, &item.ExitCode, &item.DurationMs, &item.Output, &item.ErrorOutput,
			&item.OutputTruncated, &executedAt); err != nil {
			return nil, nil, err
		}
		item.ExecutedAt = formatTime(executedAt)
		commands = append(commands, item)
	}
	if err := commandRows.Err(); err != nil {
		return nil, nil, err
	}

	return sessions, commands, nil
}

func loadWorkspace(ctx context.Context, db *sql.DB) ([]models.WorkspaceSession, []models.WorkspaceArtifact, []models.WorkspaceMessage, []models.WorkspaceMessageEvent, []models.WorkspaceArtifactAccessLog, []models.WorkspaceArtifactFavorite, []models.WorkspaceArtifactShare, error) {
	sessions := make([]models.WorkspaceSession, 0)
	sessionRows, err := db.QueryContext(ctx, `
		SELECT id, session_no, tenant_id, instance_id, title, status, COALESCE(workspace_url, ''),
		       COALESCE(protocol_version, 'openclaw-lobster-bridge/v2'),
		       last_opened_at, last_artifact_at, last_synced_at, created_at, updated_at
		FROM platform.workspace_session
		ORDER BY updated_at DESC, id DESC
	`)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}
	defer sessionRows.Close()
	for sessionRows.Next() {
		var item models.WorkspaceSession
		var lastOpenedAt sql.NullTime
		var lastArtifactAt sql.NullTime
		var lastSyncedAt sql.NullTime
		var createdAt time.Time
		var updatedAt time.Time
		if err := sessionRows.Scan(&item.ID, &item.SessionNo, &item.TenantID, &item.InstanceID, &item.Title, &item.Status, &item.WorkspaceURL, &item.ProtocolVersion, &lastOpenedAt, &lastArtifactAt, &lastSyncedAt, &createdAt, &updatedAt); err != nil {
			return nil, nil, nil, nil, nil, nil, nil, err
		}
		item.LastOpenedAt = formatNullTime(lastOpenedAt)
		item.LastArtifactAt = formatNullTime(lastArtifactAt)
		item.LastSyncedAt = formatNullTime(lastSyncedAt)
		item.CreatedAt = formatTime(createdAt)
		item.UpdatedAt = formatTime(updatedAt)
		sessions = append(sessions, item)
	}
	if err := sessionRows.Err(); err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}

	artifacts := make([]models.WorkspaceArtifact, 0)
	artifactRows, err := db.QueryContext(ctx, `
		SELECT id, session_id, tenant_id, instance_id, COALESCE(message_id, 0), title, kind, COALESCE(external_id, ''),
		       COALESCE(origin, 'manual'), source_url, COALESCE(preview_url, ''), COALESCE(archive_status, ''),
		       COALESCE(content_type, ''), COALESCE(size_bytes, 0), COALESCE(storage_bucket, ''),
		       COALESCE(storage_key, ''), COALESCE(filename, ''), COALESCE(checksum_sha256, ''), created_at, updated_at
		FROM platform.workspace_artifact
		ORDER BY created_at DESC, id DESC
	`)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}
	defer artifactRows.Close()
	for artifactRows.Next() {
		var item models.WorkspaceArtifact
		var createdAt time.Time
		var updatedAt time.Time
		if err := artifactRows.Scan(&item.ID, &item.SessionID, &item.TenantID, &item.InstanceID, &item.MessageID, &item.Title, &item.Kind, &item.ExternalID, &item.Origin, &item.SourceURL, &item.PreviewURL, &item.ArchiveStatus, &item.ContentType, &item.SizeBytes, &item.StorageBucket, &item.StorageKey, &item.Filename, &item.ChecksumSHA256, &createdAt, &updatedAt); err != nil {
			return nil, nil, nil, nil, nil, nil, nil, err
		}
		item.CreatedAt = formatTime(createdAt)
		item.UpdatedAt = formatTime(updatedAt)
		artifacts = append(artifacts, item)
	}
	if err := artifactRows.Err(); err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}

	messages := make([]models.WorkspaceMessage, 0)
	messageRows, err := db.QueryContext(ctx, `
		SELECT id, session_id, tenant_id, instance_id, COALESCE(parent_message_id, 0), role, status,
		       COALESCE(external_id, ''), COALESCE(origin, 'platform'),
		       COALESCE(trace_id, ''), COALESCE(error_code, ''), COALESCE(error_message, ''), delivery_attempt,
		       content, delivered_at, created_at, updated_at
		FROM platform.workspace_message
		ORDER BY created_at ASC, id ASC
	`)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}
	defer messageRows.Close()
	for messageRows.Next() {
		var item models.WorkspaceMessage
		var deliveredAt sql.NullTime
		var createdAt time.Time
		var updatedAt time.Time
		if err := messageRows.Scan(&item.ID, &item.SessionID, &item.TenantID, &item.InstanceID, &item.ParentMessageID, &item.Role, &item.Status, &item.ExternalID, &item.Origin, &item.TraceID, &item.ErrorCode, &item.ErrorMessage, &item.DeliveryAttempt, &item.Content, &deliveredAt, &createdAt, &updatedAt); err != nil {
			return nil, nil, nil, nil, nil, nil, nil, err
		}
		item.DeliveredAt = formatNullTime(deliveredAt)
		item.CreatedAt = formatTime(createdAt)
		item.UpdatedAt = formatTime(updatedAt)
		messages = append(messages, item)
	}
	if err := messageRows.Err(); err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}

	events := make([]models.WorkspaceMessageEvent, 0)
	eventRows, err := db.QueryContext(ctx, `
		SELECT id, session_id, COALESCE(message_id, 0), tenant_id, instance_id, event_type, COALESCE(external_id, ''),
		       COALESCE(origin, 'platform'), COALESCE(trace_id, ''), payload_json, created_at
		FROM platform.workspace_message_event
		ORDER BY id ASC
	`)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}
	defer eventRows.Close()
	for eventRows.Next() {
		var item models.WorkspaceMessageEvent
		var payloadRaw []byte
		var createdAt time.Time
		if err := eventRows.Scan(&item.ID, &item.SessionID, &item.MessageID, &item.TenantID, &item.InstanceID, &item.EventType, &item.ExternalID, &item.Origin, &item.TraceID, &payloadRaw, &createdAt); err != nil {
			return nil, nil, nil, nil, nil, nil, nil, err
		}
		item.PayloadJSON = string(payloadRaw)
		item.CreatedAt = formatTime(createdAt)
		events = append(events, item)
	}
	if err := eventRows.Err(); err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}

	logs := make([]models.WorkspaceArtifactAccessLog, 0)
	logRows, err := db.QueryContext(ctx, `
		SELECT id, artifact_id, session_id, tenant_id, instance_id, action, scope, actor,
		       COALESCE(remote_addr, ''), COALESCE(user_agent, ''), created_at
		FROM platform.workspace_artifact_access_log
		ORDER BY created_at DESC, id DESC
	`)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}
	defer logRows.Close()
	for logRows.Next() {
		var item models.WorkspaceArtifactAccessLog
		var createdAt time.Time
		if err := logRows.Scan(&item.ID, &item.ArtifactID, &item.SessionID, &item.TenantID, &item.InstanceID, &item.Action, &item.Scope, &item.Actor, &item.RemoteAddr, &item.UserAgent, &createdAt); err != nil {
			return nil, nil, nil, nil, nil, nil, nil, err
		}
		item.CreatedAt = formatTime(createdAt)
		logs = append(logs, item)
	}
	if err := logRows.Err(); err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}

	favorites := make([]models.WorkspaceArtifactFavorite, 0)
	favoriteRows, err := db.QueryContext(ctx, `
		SELECT id, artifact_id, session_id, tenant_id, instance_id, COALESCE(user_id, 0), actor, created_at
		FROM platform.workspace_artifact_favorite
		ORDER BY created_at DESC, id DESC
	`)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}
	defer favoriteRows.Close()
	for favoriteRows.Next() {
		var item models.WorkspaceArtifactFavorite
		var createdAt time.Time
		if err := favoriteRows.Scan(&item.ID, &item.ArtifactID, &item.SessionID, &item.TenantID, &item.InstanceID, &item.UserID, &item.Actor, &createdAt); err != nil {
			return nil, nil, nil, nil, nil, nil, nil, err
		}
		item.CreatedAt = formatTime(createdAt)
		favorites = append(favorites, item)
	}
	if err := favoriteRows.Err(); err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}

	shares := make([]models.WorkspaceArtifactShare, 0)
	shareRows, err := db.QueryContext(ctx, `
		SELECT id, artifact_id, session_id, tenant_id, instance_id, scope, token, COALESCE(note, ''), created_by,
		       COALESCE(created_by_user_id, 0), use_count, expires_at, last_opened_at, revoked_at, created_at, updated_at
		FROM platform.workspace_artifact_share
		ORDER BY created_at DESC, id DESC
	`)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}
	defer shareRows.Close()
	for shareRows.Next() {
		var item models.WorkspaceArtifactShare
		var expiresAt, lastOpenedAt, revokedAt sql.NullTime
		var createdAt, updatedAt time.Time
		if err := shareRows.Scan(&item.ID, &item.ArtifactID, &item.SessionID, &item.TenantID, &item.InstanceID, &item.Scope, &item.Token, &item.Note, &item.CreatedBy,
			&item.CreatedByUserID, &item.UseCount, &expiresAt, &lastOpenedAt, &revokedAt, &createdAt, &updatedAt); err != nil {
			return nil, nil, nil, nil, nil, nil, nil, err
		}
		item.ExpiresAt = formatNullTime(expiresAt)
		item.LastOpenedAt = formatNullTime(lastOpenedAt)
		item.RevokedAt = formatNullTime(revokedAt)
		item.CreatedAt = formatTime(createdAt)
		item.UpdatedAt = formatTime(updatedAt)
		shares = append(shares, item)
	}
	if err := shareRows.Err(); err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}

	return sessions, artifacts, messages, events, logs, favorites, shares, nil
}

func loadOEM(ctx context.Context, db *sql.DB) ([]models.OEMBrand, []models.OEMTheme, []models.OEMFeatureFlags, []models.TenantBrandBinding, error) {
	brandRows, err := db.QueryContext(ctx, `
		SELECT id, code, name, status, COALESCE(logo_url, ''), COALESCE(favicon_url, ''), COALESCE(support_email, ''),
		       COALESCE(support_url, ''), domains_json, updated_at, created_at
		FROM platform.oem_brand
		ORDER BY id
	`)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	defer brandRows.Close()

	brands := make([]models.OEMBrand, 0)
	for brandRows.Next() {
		var item models.OEMBrand
		var domainsRaw []byte
		var updatedAt, createdAt time.Time
		if err := brandRows.Scan(&item.ID, &item.Code, &item.Name, &item.Status, &item.LogoURL, &item.FaviconURL, &item.SupportEmail, &item.SupportURL, &domainsRaw, &updatedAt, &createdAt); err != nil {
			return nil, nil, nil, nil, err
		}
		_ = json.Unmarshal(domainsRaw, &item.Domains)
		item.UpdatedAt = formatTime(updatedAt)
		item.CreatedAt = formatTime(createdAt)
		brands = append(brands, item)
	}
	if err := brandRows.Err(); err != nil {
		return nil, nil, nil, nil, err
	}

	themes := make([]models.OEMTheme, 0)
	themeRows, err := db.QueryContext(ctx, `SELECT brand_id, primary_color, secondary_color, accent_color, surface_mode, font_family, radius FROM platform.oem_theme ORDER BY brand_id`)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	defer themeRows.Close()
	for themeRows.Next() {
		var item models.OEMTheme
		if err := themeRows.Scan(&item.BrandID, &item.PrimaryColor, &item.SecondaryColor, &item.AccentColor, &item.SurfaceMode, &item.FontFamily, &item.Radius); err != nil {
			return nil, nil, nil, nil, err
		}
		themes = append(themes, item)
	}

	features := make([]models.OEMFeatureFlags, 0)
	featureRows, err := db.QueryContext(ctx, `
		SELECT brand_id, portal_enabled, admin_enabled, channels_enabled, tickets_enabled, purchase_enabled,
		       runtime_control_enabled, audit_enabled, sso_enabled
		FROM platform.oem_feature_flags
		ORDER BY brand_id
	`)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	defer featureRows.Close()
	for featureRows.Next() {
		var item models.OEMFeatureFlags
		if err := featureRows.Scan(&item.BrandID, &item.PortalEnabled, &item.AdminEnabled, &item.ChannelsEnabled, &item.TicketsEnabled, &item.PurchaseEnabled, &item.RuntimeControlEnabled, &item.AuditEnabled, &item.SSOEnabled); err != nil {
			return nil, nil, nil, nil, err
		}
		features = append(features, item)
	}

	bindings := make([]models.TenantBrandBinding, 0)
	bindingRows, err := db.QueryContext(ctx, `SELECT tenant_id, brand_id, binding_mode, updated_at FROM platform.tenant_brand_binding ORDER BY tenant_id`)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	defer bindingRows.Close()
	for bindingRows.Next() {
		var item models.TenantBrandBinding
		var updatedAt time.Time
		if err := bindingRows.Scan(&item.TenantID, &item.BrandID, &item.BindingMode, &updatedAt); err != nil {
			return nil, nil, nil, nil, err
		}
		item.UpdatedAt = formatTime(updatedAt)
		bindings = append(bindings, item)
	}

	return brands, themes, features, bindings, nil
}

func loadCommerce(ctx context.Context, db *sql.DB) ([]models.Order, []models.Subscription, []models.PaymentTransaction, []models.RefundRecord, []models.InvoiceRecord, []models.PaymentCallbackEvent, error) {
	orders := make([]models.Order, 0)
	orderRows, err := db.QueryContext(ctx, `
		SELECT id, tenant_id, COALESCE(instance_id, 0), COALESCE(plan_code, ''), order_type, status, total_amount, currency, order_no, created_at, updated_at
		FROM platform.order_main
		ORDER BY id
	`)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	defer orderRows.Close()
	for orderRows.Next() {
		var item models.Order
		var createdAt, updatedAt time.Time
		if err := orderRows.Scan(&item.ID, &item.TenantID, &item.InstanceID, &item.PlanCode, &item.Action, &item.Status, &item.Amount, &item.Currency, &item.OrderNo, &createdAt, &updatedAt); err != nil {
			return nil, nil, nil, nil, nil, nil, err
		}
		item.CreatedAt = formatTime(createdAt)
		item.UpdatedAt = formatTime(updatedAt)
		orders = append(orders, item)
	}

	subscriptions := make([]models.Subscription, 0)
	subRows, err := db.QueryContext(ctx, `
		SELECT id, subscription_no, tenant_id, COALESCE(instance_id, 0), COALESCE(product_code, ''), COALESCE(plan_code, ''),
		       status, renew_mode, current_period_start, current_period_end, expired_at, created_at, updated_at
		FROM platform.subscription
		ORDER BY id
	`)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	defer subRows.Close()
	for subRows.Next() {
		var item models.Subscription
		var startAt, endAt, expiredAt sql.NullTime
		var createdAt, updatedAt time.Time
		if err := subRows.Scan(&item.ID, &item.SubscriptionNo, &item.TenantID, &item.InstanceID, &item.ProductCode, &item.PlanCode, &item.Status, &item.RenewMode, &startAt, &endAt, &expiredAt, &createdAt, &updatedAt); err != nil {
			return nil, nil, nil, nil, nil, nil, err
		}
		item.CurrentPeriodStart = formatNullTime(startAt)
		item.CurrentPeriodEnd = formatNullTime(endAt)
		item.ExpiredAt = formatNullTime(expiredAt)
		item.CreatedAt = formatTime(createdAt)
		item.UpdatedAt = formatTime(updatedAt)
		subscriptions = append(subscriptions, item)
	}

	payments := make([]models.PaymentTransaction, 0)
	paymentRows, err := db.QueryContext(ctx, `
		SELECT id, order_id, channel, COALESCE(pay_mode, ''), trade_no, COALESCE(channel_order_no, ''), amount, currency,
		       status, COALESCE(pay_url, ''), COALESCE(code_url, ''), COALESCE(prepay_id, ''), COALESCE(app_id, ''),
		       COALESCE(mch_id, ''), paid_at, created_at, updated_at, raw_json
		FROM platform.payment_transaction
		ORDER BY id
	`)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	defer paymentRows.Close()
	for paymentRows.Next() {
		var item models.PaymentTransaction
		var paidAt sql.NullTime
		var createdAt, updatedAt time.Time
		var rawJSON []byte
		if err := paymentRows.Scan(&item.ID, &item.OrderID, &item.Channel, &item.PayMode, &item.TradeNo, &item.ChannelOrderNo, &item.Amount, &item.Currency, &item.Status, &item.PayURL, &item.CodeURL, &item.PrepayID, &item.AppID, &item.MchID, &paidAt, &createdAt, &updatedAt, &rawJSON); err != nil {
			return nil, nil, nil, nil, nil, nil, err
		}
		item.PaidAt = formatNullTime(paidAt)
		item.CreatedAt = formatTime(createdAt)
		item.UpdatedAt = formatTime(updatedAt)
		_ = json.Unmarshal(rawJSON, &item.Raw)
		payments = append(payments, item)
	}

	refunds := make([]models.RefundRecord, 0)
	refundRows, err := db.QueryContext(ctx, `
		SELECT id, order_id, payment_id, refund_no, COALESCE(channel_refund_no, ''), status, amount, reason, COALESCE(notify_url, ''), created_at, updated_at
		FROM platform.refund_record
		ORDER BY id
	`)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	defer refundRows.Close()
	for refundRows.Next() {
		var item models.RefundRecord
		var createdAt, updatedAt time.Time
		if err := refundRows.Scan(&item.ID, &item.OrderID, &item.PaymentID, &item.RefundNo, &item.ChannelRefundNo, &item.Status, &item.Amount, &item.Reason, &item.NotifyURL, &createdAt, &updatedAt); err != nil {
			return nil, nil, nil, nil, nil, nil, err
		}
		item.CreatedAt = formatTime(createdAt)
		item.UpdatedAt = formatTime(updatedAt)
		refunds = append(refunds, item)
	}

	invoices := make([]models.InvoiceRecord, 0)
	invoiceRows, err := db.QueryContext(ctx, `
		SELECT id, tenant_id, order_id, invoice_type, status, amount, COALESCE(title, ''), COALESCE(tax_no, ''),
		       COALESCE(email, ''), COALESCE(invoice_no, ''), COALESCE(pdf_url, ''), created_at, updated_at
		FROM platform.invoice_record
		ORDER BY id
	`)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	defer invoiceRows.Close()
	for invoiceRows.Next() {
		var item models.InvoiceRecord
		var createdAt, updatedAt time.Time
		if err := invoiceRows.Scan(&item.ID, &item.TenantID, &item.OrderID, &item.InvoiceType, &item.Status, &item.Amount, &item.Title, &item.TaxNo, &item.Email, &item.InvoiceNo, &item.PDFURL, &createdAt, &updatedAt); err != nil {
			return nil, nil, nil, nil, nil, nil, err
		}
		item.CreatedAt = formatTime(createdAt)
		item.UpdatedAt = formatTime(updatedAt)
		invoices = append(invoices, item)
	}

	callbacks := make([]models.PaymentCallbackEvent, 0)
	callbackRows, err := db.QueryContext(ctx, `
		SELECT id, channel, event_type, COALESCE(out_trade_no, ''), COALESCE(out_refund_no, ''), signature_status,
		       decrypt_status, process_status, COALESCE(request_serial, ''), created_at, raw_body
		FROM platform.payment_callback_event
		ORDER BY id
	`)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	defer callbackRows.Close()
	for callbackRows.Next() {
		var item models.PaymentCallbackEvent
		var createdAt time.Time
		if err := callbackRows.Scan(&item.ID, &item.Channel, &item.EventType, &item.OutTradeNo, &item.OutRefundNo, &item.SignatureStatus, &item.DecryptStatus, &item.ProcessStatus, &item.RequestSerial, &createdAt, &item.RawBody); err != nil {
			return nil, nil, nil, nil, nil, nil, err
		}
		item.CreatedAt = formatTime(createdAt)
		callbacks = append(callbacks, item)
	}

	return orders, subscriptions, payments, refunds, invoices, callbacks, nil
}

func countRows(ctx context.Context, tx *sql.Tx, table string) (int, error) {
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	if err := tx.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func resetManagedSequences(ctx context.Context, tx *sql.Tx) error {
	tables := []struct {
		table  string
		column string
	}{
		{"platform.product", "id"},
		{"platform.service_plan", "id"},
		{"platform.tenant", "id"},
		{"platform.cluster", "id"},
		{"platform.service_instance", "id"},
		{"platform.backup_record", "id"},
		{"platform.operation_job", "id"},
		{"platform.audit_event", "id"},
		{"platform.user_account", "id"},
		{"platform.auth_identity", "id"},
		{"platform.channel", "id"},
		{"platform.channel_activity", "id"},
		{"platform.billing_statement", "id"},
		{"platform.alert_record", "id"},
		{"platform.ticket_record", "id"},
		{"platform.approval_record", "id"},
		{"platform.approval_action", "id"},
		{"platform.diagnostic_session", "id"},
		{"platform.diagnostic_command_record", "id"},
		{"platform.oem_brand", "id"},
		{"platform.order_main", "id"},
		{"platform.subscription", "id"},
		{"platform.payment_transaction", "id"},
		{"platform.refund_record", "id"},
		{"platform.invoice_record", "id"},
		{"platform.payment_callback_event", "id"},
		{"platform.workspace_session", "id"},
		{"platform.workspace_artifact", "id"},
		{"platform.workspace_message", "id"},
		{"platform.workspace_message_event", "id"},
		{"platform.workspace_artifact_access_log", "id"},
		{"platform.workspace_artifact_favorite", "id"},
		{"platform.workspace_artifact_share", "id"},
	}
	for _, item := range tables {
		if err := resetTableSequence(ctx, tx, item.table, item.column); err != nil {
			return err
		}
	}
	return nil
}

func resetTableSequence(ctx context.Context, tx *sql.Tx, table string, column string) error {
	query := fmt.Sprintf(`
		SELECT setval(
			pg_get_serial_sequence('%s', '%s'),
			COALESCE((SELECT MAX(%s) FROM %s), 1),
			true
		)
	`, table, column, column, table)
	_, err := tx.ExecContext(ctx, query)
	return err
}

func nullTime(value string) any {
	if value == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil
	}
	return parsed
}

func nullableInt(value int) any {
	if value <= 0 {
		return nil
	}
	return value
}

func defaultStringValue(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func cloneStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return map[string]string{}
	}
	output := make(map[string]string, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

func formatNullTime(value sql.NullTime) string {
	if !value.Valid {
		return ""
	}
	return formatTime(value.Time)
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}
