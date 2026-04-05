package models

type Tenant struct {
	ID         int    `json:"id"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Plan       string `json:"plan"`
	ExpiredAt  string `json:"expiredAt,omitempty"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

type Cluster struct {
	ID        int    `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	Region    string `json:"region"`
	Status    string `json:"status"`
	NodeCount int    `json:"nodeCount"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type Instance struct {
	ID          int               `json:"id"`
	TenantID    int               `json:"tenantId"`
	ClusterID   int               `json:"clusterId"`
	Code        string            `json:"code"`
	Name        string            `json:"name"`
	Status      string            `json:"status"`
	Version     string            `json:"version"`
	Plan        string            `json:"plan"`
	RuntimeType string            `json:"runtimeType"`
	Region      string            `json:"region"`
	Spec        map[string]string `json:"spec"`
	ActivatedAt string            `json:"activatedAt,omitempty"`
	ExpiredAt   string            `json:"expiredAt,omitempty"`
	CreatedAt   string            `json:"createdAt"`
	UpdatedAt   string            `json:"updatedAt"`
}

type InstanceAccess struct {
	InstanceID int    `json:"instanceId"`
	EntryType  string `json:"entryType"`
	URL        string `json:"url"`
	Domain     string `json:"domain,omitempty"`
	AccessMode string `json:"accessMode"`
	IsPrimary  bool   `json:"isPrimary"`
}

type InstanceConfig struct {
	InstanceID  int    `json:"instanceId"`
	Version     int    `json:"version"`
	Hash        string `json:"hash"`
	PublishedAt string `json:"publishedAt"`
	UpdatedBy   string `json:"updatedBy"`
	Settings    ConfigSettings `json:"settings"`
}

type ConfigSettings struct {
	Model          string `json:"model"`
	AllowedOrigins string `json:"allowedOrigins"`
	BackupPolicy   string `json:"backupPolicy"`
}

type BackupRecord struct {
	ID         int    `json:"id"`
	InstanceID int    `json:"instanceId"`
	BackupNo   string `json:"backupNo"`
	Type       string `json:"type"`
	Status     string `json:"status"`
	SizeBytes  int64  `json:"sizeBytes"`
	StartedAt  string `json:"startedAt"`
	FinishedAt string `json:"finishedAt"`
}

type Job struct {
	ID         int    `json:"id"`
	JobNo      string `json:"jobNo"`
	Type       string `json:"type"`
	TargetType string `json:"targetType"`
	TargetID   int    `json:"targetId"`
	Status     string `json:"status"`
	Summary    string `json:"summary"`
	StartedAt  string `json:"startedAt"`
	FinishedAt string `json:"finishedAt,omitempty"`
}

type Alert struct {
	ID          int    `json:"id"`
	InstanceID  int    `json:"instanceId"`
	Severity    string `json:"severity"`
	Status      string `json:"status"`
	MetricKey   string `json:"metricKey"`
	Summary     string `json:"summary"`
	TriggeredAt string `json:"triggeredAt"`
}

type AuditEvent struct {
	ID        int               `json:"id"`
	TenantID  int               `json:"tenantId"`
	Actor     string            `json:"actor"`
	Action    string            `json:"action"`
	Target    string            `json:"target"`
	TargetID  int               `json:"targetId"`
	Result    string            `json:"result"`
	CreatedAt string            `json:"createdAt"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type Channel struct {
	ID             int            `json:"id"`
	Code           string         `json:"code"`
	Name           string         `json:"name"`
	Platform       string         `json:"platform"` // e.g. feishu, dingding, wecom, slack, telegram, whatsapp, discord
	Status         string         `json:"status"`   // disconnected, connecting, connected, degraded, error
	ConnectMethod  string         `json:"connectMethod"` // oauth, webhook, token, qrcode
	AuthURL        string         `json:"authUrl,omitempty"`
	WebhookURL     string         `json:"webhookUrl,omitempty"`
	QrCodeURL      string         `json:"qrCodeUrl,omitempty"`
	TokenMasked    string         `json:"tokenMasked,omitempty"`
	CallbackSecret string         `json:"callbackSecret,omitempty"`
	Health         ChannelHealth  `json:"health"`
	Stats          ChannelStats   `json:"stats"`
	LastError      string         `json:"lastError,omitempty"`
	Settings       map[string]any `json:"settings,omitempty"`
	EntryPoints    []ChannelEntryPoint `json:"entryPoints,omitempty"`
	Notes          string         `json:"notes,omitempty"`
	UpdatedAt      string         `json:"updatedAt"`
	CreatedAt      string         `json:"createdAt"`
}

type ChannelEntryPoint struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

type ChannelHealth struct {
	LastChecked string `json:"lastChecked"`
	Status      string `json:"status"` // healthy, warning, critical
	LatencyMs   int    `json:"latencyMs"`
}

type ChannelStats struct {
	Messages24h   int     `json:"messages24h"`
	UsersActive24 int     `json:"usersActive24"`
	SuccessRate   float64 `json:"successRate"`
}

type ChannelActivity struct {
	ID        int    `json:"id"`
	ChannelID int    `json:"channelId"`
	Type      string `json:"type"` // message_in, message_out, webhook, auth
	Title     string `json:"title"`
	Summary   string `json:"summary"`
	CreatedAt string `json:"createdAt"`
}

type InstanceRuntime struct {
	InstanceID       int    `json:"instanceId"`
	PowerState       string `json:"powerState"` // running, stopped, restarting
	CPUUsagePercent  int    `json:"cpuUsagePercent"`
	MemoryUsagePercent int  `json:"memoryUsagePercent"`
	DiskUsagePercent int    `json:"diskUsagePercent"`
	APIRequests24h   int    `json:"apiRequests24h"`
	APITokens24h     int    `json:"apiTokens24h"`
	LastSeenAt       string `json:"lastSeenAt"`
}

type InstanceCredential struct {
	InstanceID      int    `json:"instanceId"`
	AdminUser       string `json:"adminUser"`
	PasswordMasked  string `json:"passwordMasked"`
	LastRotatedAt   string `json:"lastRotatedAt"`
	RequiresReset   bool   `json:"requiresReset"`
}

type PlanOffer struct {
	ID           int      `json:"id"`
	Code         string   `json:"code"`
	Name         string   `json:"name"`
	MonthlyPrice int      `json:"monthlyPrice"`
	CPU          string   `json:"cpu"`
	Memory       string   `json:"memory"`
	Storage      string   `json:"storage"`
	Highlight    string   `json:"highlight"`
	Features     []string `json:"features"`
}

type Order struct {
	ID         int    `json:"id"`
	TenantID   int    `json:"tenantId"`
	InstanceID int    `json:"instanceId,omitempty"`
	PlanCode   string `json:"planCode"`
	Action     string `json:"action"` // buy, renew, upgrade
	Status     string `json:"status"` // pending, paid, active
	Amount     int    `json:"amount"`
	CreatedAt  string `json:"createdAt"`
}

type Ticket struct {
	ID          int    `json:"id"`
	TicketNo    string `json:"ticketNo"`
	TenantID    int    `json:"tenantId"`
	InstanceID  int    `json:"instanceId,omitempty"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	Severity    string `json:"severity"`
	Status      string `json:"status"` // open, in_progress, resolved, closed
	Reporter    string `json:"reporter"`
	Assignee    string `json:"assignee,omitempty"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type OEMBrand struct {
	ID          int      `json:"id"`
	Code        string   `json:"code"`
	Name        string   `json:"name"`
	Status      string   `json:"status"` // active, inactive, draft
	LogoURL     string   `json:"logoUrl"`
	FaviconURL  string   `json:"faviconUrl"`
	SupportEmail string  `json:"supportEmail"`
	SupportURL  string   `json:"supportUrl"`
	Domains     []string `json:"domains"`
	UpdatedAt   string   `json:"updatedAt"`
	CreatedAt   string   `json:"createdAt"`
}

type OEMTheme struct {
	BrandID        int    `json:"brandId"`
	PrimaryColor   string `json:"primaryColor"`
	SecondaryColor string `json:"secondaryColor"`
	AccentColor    string `json:"accentColor"`
	SurfaceMode    string `json:"surfaceMode"` // light, dark, hybrid
	FontFamily     string `json:"fontFamily"`
	Radius         string `json:"radius"`
}

type OEMFeatureFlags struct {
	BrandID             int  `json:"brandId"`
	PortalEnabled       bool `json:"portalEnabled"`
	AdminEnabled        bool `json:"adminEnabled"`
	ChannelsEnabled     bool `json:"channelsEnabled"`
	TicketsEnabled      bool `json:"ticketsEnabled"`
	PurchaseEnabled     bool `json:"purchaseEnabled"`
	RuntimeControlEnabled bool `json:"runtimeControlEnabled"`
	AuditEnabled        bool `json:"auditEnabled"`
	SSOEnabled          bool `json:"ssoEnabled"`
}

type TenantBrandBinding struct {
	TenantID   int    `json:"tenantId"`
	BrandID    int    `json:"brandId"`
	BindingMode string `json:"bindingMode"` // dedicated, shared
	UpdatedAt  string `json:"updatedAt"`
}

type Data struct {
	Tenants   []Tenant
	Clusters  []Cluster
	Instances []Instance
	Accesses  []InstanceAccess
	Configs   []InstanceConfig
	Backups   []BackupRecord
	Jobs      []Job
	Alerts    []Alert
	Audits    []AuditEvent
	Channels  []Channel
	Activities []ChannelActivity
	Runtimes  []InstanceRuntime
	Credentials []InstanceCredential
	PlanOffers []PlanOffer
	Orders    []Order
	Tickets   []Ticket
	Brands    []OEMBrand
	BrandThemes []OEMTheme
	BrandFeatures []OEMFeatureFlags
	BrandBindings []TenantBrandBinding
}
