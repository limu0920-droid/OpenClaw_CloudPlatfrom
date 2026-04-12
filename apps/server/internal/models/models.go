package models

type Tenant struct {
	ID        int    `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Plan      string `json:"plan"`
	ExpiredAt string `json:"expiredAt,omitempty"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type UserProfile struct {
	ID             int    `json:"id"`
	TenantID       int    `json:"tenantId"`
	LoginName      string `json:"loginName"`
	Status         string `json:"status"` // active, disabled, locked
	LockReason     string `json:"lockReason,omitempty"`
	DisplayName    string `json:"displayName"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	AvatarURL      string `json:"avatarUrl,omitempty"`
	Locale         string `json:"locale"`
	Timezone       string `json:"timezone"`
	Department     string `json:"department,omitempty"`
	Title          string `json:"title,omitempty"`
	Bio            string `json:"bio,omitempty"`
	PasswordMasked string `json:"passwordMasked"`
	UpdatedAt      string `json:"updatedAt"`
}

type AuthIdentity struct {
	ID           int    `json:"id"`
	UserID       int    `json:"userId"`
	TenantID     int    `json:"tenantId"`
	Provider     string `json:"provider"` // keycloak, wechat
	IsPrimary    bool   `json:"isPrimary"`
	Status       string `json:"status"` // active, disabled
	StatusReason string `json:"statusReason,omitempty"`
	Subject      string `json:"subject"`
	Email        string `json:"email,omitempty"`
	OpenID       string `json:"openId,omitempty"`
	UnionID      string `json:"unionId,omitempty"`
	ExternalName string `json:"externalName,omitempty"`
	LastLoginAt  string `json:"lastLoginAt,omitempty"`
	UpdatedAt    string `json:"updatedAt"`
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
	InstanceID  int            `json:"instanceId"`
	Version     int            `json:"version"`
	Hash        string         `json:"hash"`
	PublishedAt string         `json:"publishedAt"`
	UpdatedBy   string         `json:"updatedBy"`
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
	ID             int                 `json:"id"`
	Code           string              `json:"code"`
	Name           string              `json:"name"`
	Platform       string              `json:"platform"`      // e.g. feishu, dingding, wecom, slack, telegram, whatsapp, discord
	Status         string              `json:"status"`        // disconnected, connecting, connected, degraded, error
	ConnectMethod  string              `json:"connectMethod"` // oauth, webhook, token, qrcode
	AuthURL        string              `json:"authUrl,omitempty"`
	WebhookURL     string              `json:"webhookUrl,omitempty"`
	QrCodeURL      string              `json:"qrCodeUrl,omitempty"`
	TokenMasked    string              `json:"tokenMasked,omitempty"`
	CallbackSecret string              `json:"callbackSecret,omitempty"`
	Health         ChannelHealth       `json:"health"`
	Stats          ChannelStats        `json:"stats"`
	LastError      string              `json:"lastError,omitempty"`
	Settings       map[string]any      `json:"settings,omitempty"`
	EntryPoints    []ChannelEntryPoint `json:"entryPoints,omitempty"`
	Notes          string              `json:"notes,omitempty"`
	UpdatedAt      string              `json:"updatedAt"`
	CreatedAt      string              `json:"createdAt"`
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
	InstanceID         int    `json:"instanceId"`
	PowerState         string `json:"powerState"` // running, stopped, restarting
	CPUUsagePercent    int    `json:"cpuUsagePercent"`
	MemoryUsagePercent int    `json:"memoryUsagePercent"`
	DiskUsagePercent   int    `json:"diskUsagePercent"`
	APIRequests24h     int    `json:"apiRequests24h"`
	APITokens24h       int    `json:"apiTokens24h"`
	LastSeenAt         string `json:"lastSeenAt"`
}

type InstanceCredential struct {
	InstanceID     int    `json:"instanceId"`
	AdminUser      string `json:"adminUser"`
	PasswordMasked string `json:"passwordMasked"`
	LastRotatedAt  string `json:"lastRotatedAt"`
	RequiresReset  bool   `json:"requiresReset"`
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
	Currency   string `json:"currency"`
	OrderNo    string `json:"orderNo"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

type Subscription struct {
	ID                 int    `json:"id"`
	SubscriptionNo     string `json:"subscriptionNo"`
	TenantID           int    `json:"tenantId"`
	InstanceID         int    `json:"instanceId,omitempty"`
	ProductCode        string `json:"productCode"`
	PlanCode           string `json:"planCode"`
	Status             string `json:"status"`    // pending, active, expired, cancelled
	RenewMode          string `json:"renewMode"` // manual, auto
	CurrentPeriodStart string `json:"currentPeriodStart"`
	CurrentPeriodEnd   string `json:"currentPeriodEnd"`
	ExpiredAt          string `json:"expiredAt,omitempty"`
	CreatedAt          string `json:"createdAt"`
	UpdatedAt          string `json:"updatedAt"`
}

type PaymentTransaction struct {
	ID             int            `json:"id"`
	OrderID        int            `json:"orderId"`
	Channel        string         `json:"channel"` // wechatpay, alipay, stripe
	PayMode        string         `json:"payMode"` // native, h5, jsapi
	TradeNo        string         `json:"tradeNo"`
	ChannelOrderNo string         `json:"channelOrderNo,omitempty"`
	Amount         int            `json:"amount"`
	Currency       string         `json:"currency"`
	Status         string         `json:"status"` // created, paying, paid, failed, refunded, closed
	PayURL         string         `json:"payUrl,omitempty"`
	CodeURL        string         `json:"codeUrl,omitempty"`
	PrepayID       string         `json:"prepayId,omitempty"`
	AppID          string         `json:"appId,omitempty"`
	MchID          string         `json:"mchId,omitempty"`
	PaidAt         string         `json:"paidAt,omitempty"`
	CreatedAt      string         `json:"createdAt"`
	UpdatedAt      string         `json:"updatedAt"`
	Raw            map[string]any `json:"raw,omitempty"`
}

type RefundRecord struct {
	ID              int    `json:"id"`
	OrderID         int    `json:"orderId"`
	PaymentID       int    `json:"paymentId"`
	RefundNo        string `json:"refundNo"`
	ChannelRefundNo string `json:"channelRefundNo,omitempty"`
	Status          string `json:"status"` // pending, success, failed
	Amount          int    `json:"amount"`
	Reason          string `json:"reason"`
	NotifyURL       string `json:"notifyUrl,omitempty"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}

type InvoiceRecord struct {
	ID          int    `json:"id"`
	TenantID    int    `json:"tenantId"`
	OrderID     int    `json:"orderId"`
	InvoiceType string `json:"invoiceType"`
	Status      string `json:"status"`
	Amount      int    `json:"amount"`
	Title       string `json:"title"`
	TaxNo       string `json:"taxNo,omitempty"`
	Email       string `json:"email,omitempty"`
	InvoiceNo   string `json:"invoiceNo,omitempty"`
	PDFURL      string `json:"pdfUrl,omitempty"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type PaymentCallbackEvent struct {
	ID              int    `json:"id"`
	Channel         string `json:"channel"`
	EventType       string `json:"eventType"`
	OutTradeNo      string `json:"outTradeNo,omitempty"`
	OutRefundNo     string `json:"outRefundNo,omitempty"`
	SignatureStatus string `json:"signatureStatus"` // skipped, verified, failed
	DecryptStatus   string `json:"decryptStatus"`   // skipped, success, failed
	ProcessStatus   string `json:"processStatus"`   // accepted, succeeded, failed
	RequestSerial   string `json:"requestSerial,omitempty"`
	CreatedAt       string `json:"createdAt"`
	RawBody         string `json:"rawBody"`
}

type AccountSettings struct {
	TenantID               int    `json:"tenantId"`
	PrimaryEmail           string `json:"primaryEmail"`
	BillingEmail           string `json:"billingEmail"`
	AlertEmail             string `json:"alertEmail"`
	PreferredLocale        string `json:"preferredLocale"`
	SecondaryLocale        string `json:"secondaryLocale"`
	Timezone               string `json:"timezone"`
	EmailVerified          bool   `json:"emailVerified"`
	MarketingOptIn         bool   `json:"marketingOptIn"`
	NotifyOnAlert          bool   `json:"notifyOnAlert"`
	NotifyOnPayment        bool   `json:"notifyOnPayment"`
	NotifyOnExpiry         bool   `json:"notifyOnExpiry"`
	NotifyChannelEmail     bool   `json:"notifyChannelEmail"`
	NotifyChannelWebhook   bool   `json:"notifyChannelWebhook"`
	NotifyChannelInApp     bool   `json:"notifyChannelInApp"`
	NotificationWebhookURL string `json:"notificationWebhookUrl,omitempty"`
	PortalHeadline         string `json:"portalHeadline,omitempty"`
	PortalSubtitle         string `json:"portalSubtitle,omitempty"`
	WorkspaceCallout       string `json:"workspaceCallout,omitempty"`
	ExperimentBadge        string `json:"experimentBadge,omitempty"`
	UpdatedAt              string `json:"updatedAt"`
}

type WalletBalance struct {
	TenantID         int    `json:"tenantId"`
	Currency         string `json:"currency"`
	AvailableAmount  int    `json:"availableAmount"`
	FrozenAmount     int    `json:"frozenAmount"`
	CreditLimit      int    `json:"creditLimit"`
	AutoRecharge     bool   `json:"autoRecharge"`
	LastSettlementAt string `json:"lastSettlementAt,omitempty"`
	UpdatedAt        string `json:"updatedAt"`
}

type BillingStatement struct {
	ID             int    `json:"id"`
	TenantID       int    `json:"tenantId"`
	StatementNo    string `json:"statementNo"`
	BillingMonth   string `json:"billingMonth"`
	Status         string `json:"status"` // open, settled, overdue
	Currency       string `json:"currency"`
	OpeningBalance int    `json:"openingBalance"`
	ChargeAmount   int    `json:"chargeAmount"`
	RefundAmount   int    `json:"refundAmount"`
	ClosingBalance int    `json:"closingBalance"`
	PaidAmount     int    `json:"paidAmount"`
	DueAt          string `json:"dueAt"`
	CreatedAt      string `json:"createdAt"`
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
	ID           int      `json:"id"`
	Code         string   `json:"code"`
	Name         string   `json:"name"`
	Status       string   `json:"status"` // active, inactive, draft
	LogoURL      string   `json:"logoUrl"`
	FaviconURL   string   `json:"faviconUrl"`
	SupportEmail string   `json:"supportEmail"`
	SupportURL   string   `json:"supportUrl"`
	Domains      []string `json:"domains"`
	UpdatedAt    string   `json:"updatedAt"`
	CreatedAt    string   `json:"createdAt"`
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
	BrandID               int  `json:"brandId"`
	PortalEnabled         bool `json:"portalEnabled"`
	AdminEnabled          bool `json:"adminEnabled"`
	ChannelsEnabled       bool `json:"channelsEnabled"`
	TicketsEnabled        bool `json:"ticketsEnabled"`
	PurchaseEnabled       bool `json:"purchaseEnabled"`
	RuntimeControlEnabled bool `json:"runtimeControlEnabled"`
	AuditEnabled          bool `json:"auditEnabled"`
	SSOEnabled            bool `json:"ssoEnabled"`
}

type TenantBrandBinding struct {
	TenantID    int    `json:"tenantId"`
	BrandID     int    `json:"brandId"`
	BindingMode string `json:"bindingMode"` // dedicated, shared
	UpdatedAt   string `json:"updatedAt"`
}

type RuntimeBinding struct {
	InstanceID   int    `json:"instanceId"`
	ClusterID    string `json:"clusterId"`
	Namespace    string `json:"namespace"`
	WorkloadID   string `json:"workloadId"`
	WorkloadName string `json:"workloadName"`
}

type ApprovalRecord struct {
	ID              int               `json:"id"`
	ApprovalNo      string            `json:"approvalNo"`
	TenantID        int               `json:"tenantId"`
	InstanceID      int               `json:"instanceId,omitempty"`
	ApprovalType    string            `json:"approvalType"`
	TargetType      string            `json:"targetType"`
	TargetID        int               `json:"targetId,omitempty"`
	ApplicantID     int               `json:"applicantId"`
	ApplicantName   string            `json:"applicantName,omitempty"`
	ApproverID      int               `json:"approverId,omitempty"`
	ApproverName    string            `json:"approverName,omitempty"`
	ExecutorID      int               `json:"executorId,omitempty"`
	ExecutorName    string            `json:"executorName,omitempty"`
	Status          string            `json:"status"` // pending, approved, rejected, executing, executed, cancelled, expired
	RiskLevel       string            `json:"riskLevel"`
	Reason          string            `json:"reason,omitempty"`
	ApprovalComment string            `json:"approvalComment,omitempty"`
	RejectReason    string            `json:"rejectReason,omitempty"`
	ApprovedAt      string            `json:"approvedAt,omitempty"`
	ExecutedAt      string            `json:"executedAt,omitempty"`
	ExpiredAt       string            `json:"expiredAt,omitempty"`
	CreatedAt       string            `json:"createdAt"`
	UpdatedAt       string            `json:"updatedAt"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

type ApprovalAction struct {
	ID         int               `json:"id"`
	ApprovalID int               `json:"approvalId"`
	ActorID    int               `json:"actorId,omitempty"`
	ActorName  string            `json:"actorName"`
	Action     string            `json:"action"` // submitted, approved, rejected, executed, cancelled
	Comment    string            `json:"comment,omitempty"`
	CreatedAt  string            `json:"createdAt"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type DiagnosticSession struct {
	ID             int    `json:"id"`
	SessionNo      string `json:"sessionNo"`
	TenantID       int    `json:"tenantId"`
	InstanceID     int    `json:"instanceId"`
	ClusterID      string `json:"clusterId,omitempty"`
	Namespace      string `json:"namespace,omitempty"`
	WorkloadID     string `json:"workloadId,omitempty"`
	WorkloadName   string `json:"workloadName,omitempty"`
	PodName        string `json:"podName"`
	ContainerName  string `json:"containerName,omitempty"`
	AccessMode     string `json:"accessMode"` // readonly, whitelist
	Status         string `json:"status"`     // active, closed, expired, failed
	ApprovalTicket string `json:"approvalTicket,omitempty"`
	ApprovedBy     string `json:"approvedBy,omitempty"`
	Operator       string `json:"operator"`
	OperatorUserID int    `json:"operatorUserId,omitempty"`
	Reason         string `json:"reason,omitempty"`
	CloseReason    string `json:"closeReason,omitempty"`
	ExpiresAt      string `json:"expiresAt,omitempty"`
	LastCommandAt  string `json:"lastCommandAt,omitempty"`
	StartedAt      string `json:"startedAt"`
	EndedAt        string `json:"endedAt,omitempty"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}

type DiagnosticCommandRecord struct {
	ID              int    `json:"id"`
	SessionID       int    `json:"sessionId"`
	TenantID        int    `json:"tenantId"`
	InstanceID      int    `json:"instanceId"`
	CommandKey      string `json:"commandKey,omitempty"`
	CommandText     string `json:"commandText"`
	Status          string `json:"status"` // succeeded, failed, blocked, timeout
	ExitCode        int    `json:"exitCode"`
	DurationMs      int    `json:"durationMs"`
	Output          string `json:"output,omitempty"`
	ErrorOutput     string `json:"errorOutput,omitempty"`
	OutputTruncated bool   `json:"outputTruncated,omitempty"`
	ExecutedAt      string `json:"executedAt"`
}

type WorkspaceSession struct {
	ID              int    `json:"id"`
	SessionNo       string `json:"sessionNo"`
	TenantID        int    `json:"tenantId"`
	InstanceID      int    `json:"instanceId"`
	Title           string `json:"title"`
	Status          string `json:"status"` // active, archived
	WorkspaceURL    string `json:"workspaceUrl"`
	ProtocolVersion string `json:"protocolVersion,omitempty"`
	LastOpenedAt    string `json:"lastOpenedAt,omitempty"`
	LastArtifactAt  string `json:"lastArtifactAt,omitempty"`
	LastSyncedAt    string `json:"lastSyncedAt,omitempty"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}

type WorkspaceArtifact struct {
	ID             int    `json:"id"`
	SessionID      int    `json:"sessionId"`
	TenantID       int    `json:"tenantId"`
	InstanceID     int    `json:"instanceId"`
	MessageID      int    `json:"messageId,omitempty"`
	Title          string `json:"title"`
	Kind           string `json:"kind"` // web, pdf, pptx, docx, xlsx, image, video, audio, text, unknown
	ExternalID     string `json:"externalId,omitempty"`
	Origin         string `json:"origin,omitempty"`
	SourceURL      string `json:"sourceUrl"`
	PreviewURL     string `json:"previewUrl,omitempty"`
	ArchiveStatus  string `json:"archiveStatus"`
	ContentType    string `json:"contentType,omitempty"`
	SizeBytes      int64  `json:"sizeBytes"`
	StorageBucket  string `json:"storageBucket,omitempty"`
	StorageKey     string `json:"storageKey,omitempty"`
	Filename       string `json:"filename,omitempty"`
	ChecksumSHA256 string `json:"checksumSha256,omitempty"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}

type WorkspaceMessage struct {
	ID              int    `json:"id"`
	SessionID       int    `json:"sessionId"`
	TenantID        int    `json:"tenantId"`
	InstanceID      int    `json:"instanceId"`
	ParentMessageID int    `json:"parentMessageId,omitempty"`
	Role            string `json:"role"`   // user, assistant, system, note
	Status          string `json:"status"` // recorded, sent, delivered, failed
	ExternalID      string `json:"externalId,omitempty"`
	Origin          string `json:"origin,omitempty"`
	TraceID         string `json:"traceId,omitempty"`
	ErrorCode       string `json:"errorCode,omitempty"`
	ErrorMessage    string `json:"errorMessage,omitempty"`
	DeliveryAttempt int    `json:"deliveryAttempt,omitempty"`
	Content         string `json:"content"`
	DeliveredAt     string `json:"deliveredAt,omitempty"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}

type WorkspaceMessageEvent struct {
	ID          int    `json:"id"`
	SessionID   int    `json:"sessionId"`
	MessageID   int    `json:"messageId,omitempty"`
	TenantID    int    `json:"tenantId"`
	InstanceID  int    `json:"instanceId"`
	EventType   string `json:"eventType"`
	ExternalID  string `json:"externalId,omitempty"`
	Origin      string `json:"origin,omitempty"`
	TraceID     string `json:"traceId,omitempty"`
	PayloadJSON string `json:"payloadJson,omitempty"`
	CreatedAt   string `json:"createdAt"`
}

type WorkspaceArtifactAccessLog struct {
	ID         int    `json:"id"`
	ArtifactID int    `json:"artifactId"`
	SessionID  int    `json:"sessionId"`
	TenantID   int    `json:"tenantId"`
	InstanceID int    `json:"instanceId"`
	Action     string `json:"action"` // detail, view, download
	Scope      string `json:"scope"`  // portal, admin
	Actor      string `json:"actor"`
	RemoteAddr string `json:"remoteAddr,omitempty"`
	UserAgent  string `json:"userAgent,omitempty"`
	CreatedAt  string `json:"createdAt"`
}

type WorkspaceArtifactFavorite struct {
	ID         int    `json:"id"`
	ArtifactID int    `json:"artifactId"`
	SessionID  int    `json:"sessionId"`
	TenantID   int    `json:"tenantId"`
	InstanceID int    `json:"instanceId"`
	UserID     int    `json:"userId,omitempty"`
	Actor      string `json:"actor"`
	CreatedAt  string `json:"createdAt"`
}

type WorkspaceArtifactShare struct {
	ID              int    `json:"id"`
	ArtifactID      int    `json:"artifactId"`
	SessionID       int    `json:"sessionId"`
	TenantID        int    `json:"tenantId"`
	InstanceID      int    `json:"instanceId"`
	Scope           string `json:"scope"`
	Token           string `json:"token"`
	Note            string `json:"note,omitempty"`
	CreatedBy       string `json:"createdBy"`
	CreatedByUserID int    `json:"createdByUserId,omitempty"`
	UseCount        int    `json:"useCount"`
	ExpiresAt       string `json:"expiresAt,omitempty"`
	LastOpenedAt    string `json:"lastOpenedAt,omitempty"`
	RevokedAt       string `json:"revokedAt,omitempty"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}

type Data struct {
	Tenants                    []Tenant
	Users                      []UserProfile
	AuthIdentities             []AuthIdentity
	Clusters                   []Cluster
	Instances                  []Instance
	Accesses                   []InstanceAccess
	Configs                    []InstanceConfig
	Backups                    []BackupRecord
	Jobs                       []Job
	Alerts                     []Alert
	Audits                     []AuditEvent
	Channels                   []Channel
	Activities                 []ChannelActivity
	Runtimes                   []InstanceRuntime
	Credentials                []InstanceCredential
	PlanOffers                 []PlanOffer
	Orders                     []Order
	Subscriptions              []Subscription
	Payments                   []PaymentTransaction
	Refunds                    []RefundRecord
	Invoices                   []InvoiceRecord
	PaymentCallbackEvents      []PaymentCallbackEvent
	AccountSettings            []AccountSettings
	Wallets                    []WalletBalance
	BillingStatements          []BillingStatement
	Tickets                    []Ticket
	Brands                     []OEMBrand
	BrandThemes                []OEMTheme
	BrandFeatures              []OEMFeatureFlags
	BrandBindings              []TenantBrandBinding
	RuntimeBindings            []RuntimeBinding
	Approvals                  []ApprovalRecord
	ApprovalActions            []ApprovalAction
	DiagnosticSessions         []DiagnosticSession
	DiagnosticCommandRecords   []DiagnosticCommandRecord
	WorkspaceSessions          []WorkspaceSession
	WorkspaceArtifacts         []WorkspaceArtifact
	WorkspaceMessages          []WorkspaceMessage
	WorkspaceMessageEvents     []WorkspaceMessageEvent
	WorkspaceArtifactLogs      []WorkspaceArtifactAccessLog
	WorkspaceArtifactFavorites []WorkspaceArtifactFavorite
	WorkspaceArtifactShares    []WorkspaceArtifactShare
}
