# OpenClaw 渠道接入中心对标调研

## 研究问题

- 当前主流 AI 平台 / Bot 平台在“第三方聊天平台接入”上，控制台通常提供哪些能力？
- 这些能力中，哪些适合直接落到 OpenClaw 一期的 `Channels / 渠道接入中心`？

## 结论先行

- 市面上成熟的平台几乎都不会只给一个“Webhook 输入框”，而是会把渠道接入做成独立能力域。
- 一期最值得直接照着做的，不是“支持全部平台所有深水区能力”，而是先把下面这组共性做齐：
  - 渠道目录 / 搜索 / 已连接状态
  - 接入方式区分：`OAuth`、`Webhook`、`Stream`、`Bot Token`、`QR`、`Manual`
  - 一键连接或半自动配置向导
  - Callback / Signing Secret / Verify Token / Secret Token / Redirect URL 等关键字段
  - 连接状态、健康检查、最近活动、最近错误
  - 能力矩阵：是否支持收发消息、群聊、单聊、线程、卡片、文件、富文本、Slash Command
  - 会话入口与来源标识
- 对 OpenClaw 一期来说，最合理的渠道范围是：
  - 海外：`Slack`、`Telegram`、`Discord`、`WhatsApp`
  - 国内：`飞书`、`企业微信`、`钉钉`

## 官方资料事实

### 1. Botpress 把“外部消息渠道”当成独立集成域

来源：

- https://www.botpress.com/docs/integrations/sdk/integration/messaging
- https://www.botpress.com/docs/learn/reference/integrations
- https://botpress.com/docs/integrations/integration-guides/slack

事实：

- Botpress 明确说明：如果没有 Integration，机器人无法发送或接收外部消息。
- 外部渠道接入流程被抽象成固定链路：
  1. 外部服务收消息
  2. 外部服务把消息转发到平台的 Webhook URL
  3. 平台把消息交给运行时
  4. 运行时返回回复
  5. Integration 再把回复发回外部平台
- Botpress 文档明确提到，不同平台的接入方式不同：
  - Telegram 可以用 API 设置 webhook
  - Slack 需要在 Slack App 设置中配置 webhook URL
- Botpress 的 Integrations Hub 支持：
  - 搜索集成
  - 安装集成
  - 展示 prerequisites / installation instructions
  - 标记 verified integrations
- 官方 Slack 集成还提供：
  - 自动配置与手动配置
  - 填写 `Client ID`、`Client Secret`、`Signing Secret`
  - 复制平台生成的 webhook URL 并填回 Slack App
  - 自定义 Bot 昵称 / 头像
  - 限制说明，如 rate limit 和富文本兼容问题

推断：

- `Channels Hub + Install/Connect + Config + Limitations + Activity` 这套结构，已经是被验证过的标准形态。

### 2. Slack 官方已经把接入能力分成 OAuth、事件回调、App Home、企业级安装

来源：

- https://docs.slack.dev/authentication/installing-with-oauth/
- https://docs.slack.dev/apis/events-api/
- https://api.slack.com/surfaces/tabs/using
- https://api.slack.com/enterprise/apps/changes-install

事实：

- Slack App 安装使用 OAuth v2 流程：
  - `/oauth/v2/authorize`
  - `redirect_uri`
  - `oauth.v2.access`
- Events API 支持两种事件接收方式：
  - `Socket Mode`
  - `HTTP endpoint`
- 事件订阅与 OAuth scope 绑定。
- App Home 是独立能力面，需要订阅 `app_home_opened`。
- Enterprise 安装场景下，会出现 `is_enterprise_install`，还涉及工作区分发。

推断：

- OpenClaw 控制台里，Slack 渠道页至少应该显式展示：
  - OAuth Redirect URL
  - Events Endpoint / Socket Mode
  - Scopes / Install Status
  - Workspace / Org 级安装状态

### 3. Discord 官方把“交互式应用”做成命令、组件、模态框、交互端点

来源：

- https://docs.discord.com/developers/interactions/overview
- https://docs.discord.com/developers/interactions/receiving-and-responding

事实：

- Discord 支持：
  - Slash commands
  - Message commands
  - User commands
  - Message components
  - Modals
- 交互可通过：
  - Gateway connection
  - HTTP outgoing webhook
- 配置 HTTP 交互端点时，必须：
  - 正确回应 `PING`
  - 校验 `X-Signature-Ed25519` 与 `X-Signature-Timestamp`

推断：

- OpenClaw 的 Discord 渠道页不能只展示“机器人 token”，还应展示：
  - Interactions Endpoint URL
  - Signature 验证状态
  - Commands / Components / Modal 支持能力

### 4. Telegram 官方重点是 Bot Token、Webhook 与 Secret Token

来源：

- https://core.telegram.org/bots/api
- https://core.telegram.org/bots/faq

事实：

- Telegram Bot API 支持 webhook。
- `update_id` 可用于保证更新顺序与去重。
- `secret_token` 会放在 `X-Telegram-Bot-Api-Secret-Token` 请求头中。
- 长轮询与 webhook 不能同时使用。

推断：

- OpenClaw 的 Telegram 渠道页应明确：
  - 当前接收方式：Webhook 或 Polling
  - Webhook URL
  - Secret Token
  - 最近 update 顺序与递送状态

### 5. 钉钉官方区分应用机器人、群 Webhook 机器人，以及 Stream / Webhook 两种接收模式

来源：

- https://opensource.dingtalk.com/developerpedia/docs/learn/bot/overview/
- https://opensource.dingtalk.com/developerpedia/docs/learn/bot/appbot/receive/
- https://open.dingtalk.com/tutorial/

事实：

- 钉钉聊天机器人分为：
  - 应用机器人
  - 群自定义 Webhook 机器人
- 官方推荐应用机器人。
- 应用机器人支持：
  - 完整收发消息
  - 单聊和群聊
  - Stream Mode
  - 以应用身份调用 OpenAPI
  - 互动卡片消息
- 机器人接收消息支持：
  - `Stream` 模式
  - `Webhook` 模式
- 钉钉官方教程还把：
  - 扫码登录第三方网站
  - 钉钉内免登第三方网站
  - 使用钉钉账号登录第三方网站
 作为标准接入能力提供。

推断：

- 对 OpenClaw 来说，钉钉不是只有“群消息转发”这么简单。
- 一期控制台中，钉钉渠道卡片至少要把这些能力显式分类：
  - 聊天机器人
  - Webhook 群机器人
  - Stream 接收
  - 扫码 / 免登 / OAuth 类登录接入

### 6. Dify / Marketplace 方向说明“集成中心”需要安装、权限与发布视角

来源：

- https://docs.dify.ai/en/use-dify/publish/webapp/web-app-access
- https://docs.dify.ai/en/develop-plugin/publishing/marketplace-listing/release-to-dify-marketplace

事实：

- Dify 把 Web App 的访问控制当成发布配置的一部分。
- Dify 也把插件/扩展的发布、市场分发和权限控制当成一等公民。

推断：

- OpenClaw 后续不仅要有“接入中心”，还应有：
  - 权限可见性
  - 谁能启用某个渠道
  - 哪些渠道属于租户级，哪些属于平台级

## 一期建议直接实现的能力

### Portal 侧

- 渠道接入中心列表
- 一键连接卡片
- 支持 `OAuth / Webhook / Stream / Token / QR / Manual` 方式标识
- 渠道详情页
  - 连接状态
  - 接入方式
  - 必填字段
  - 回调 URL / 重定向 URL / Secret
  - 最近活动
  - 最近错误
  - 最近会话入口
- 连接 / 断开 / 健康检查

### Admin 侧

- 平台渠道总览
- 按租户 / 渠道 / 状态 / 健康筛选
- 渠道详情页
  - 能力矩阵
  - 最近活动
  - 最近错误
  - 会话与来源分布
  - 风险提示

## 一期不建议硬做的能力

- 真实 OAuth 联邦登录与回调验签
- WhatsApp Embedded Signup 全流程
- Discord 命令自动注册
- Slack Enterprise Grid 真正的组织级审批流
- 钉钉 / 飞书 / 企业微信 的真实企业安装流程
- 多地区多租户的消息限流、回退、灰度发布

## 对当前仓库的落地建议

1. 增加 `channels` 域模型
2. 增加 `portal/channels` 与 `admin/channels` 路由
3. 为每个渠道定义：
   - `connectMode`
   - `status`
   - `health`
   - `capabilities`
   - `credentials`
   - `recentActivities`
4. 用 Go Mock API 先模拟：
   - 连接
   - 断开
   - 健康检查
   - 最近事件
5. 把“其他平台有的功能”拆成能力矩阵，而不是一次性全做成真实联通
