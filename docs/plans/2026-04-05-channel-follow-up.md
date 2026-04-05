# Channel Integration Follow-up Plan

## 当前已完成

- Portal 渠道接入中心
- Admin 渠道列表与详情
- Go Mock API 渠道域模型
- 渠道连接 / 断开 / 健康检查
- 多平台 seed：
  - 飞书
  - 企业微信
  - 钉钉
  - Slack
  - Telegram
  - Discord
  - WhatsApp
  - 自定义 Webhook

## 当前仍未实现

### 1. 真正的外部平台接入

- Slack OAuth 回调与 scope 管理
- Discord Interactions Endpoint 验签
- Telegram Webhook Secret Token 与 update 顺序校验
- WhatsApp Cloud API / Embedded Signup
- 钉钉应用机器人 Stream / Webhook 真接入
- 飞书 / 企业微信 / 钉钉的企业应用安装、免登、回调验签

### 2. 渠道运营能力

- 会话列表
- 会话详情
- 多渠道来源路由
- 渠道级限流与重试
- 富文本 / 卡片 / 模板消息能力矩阵
- 人工接管 / 指派坐席 / SLA

### 3. 渠道安全与审计

- Callback 签名校验
- 密钥轮换
- 渠道级访问控制
- 敏感词与风控策略
- 渠道级审计与导出

### 4. 控制台增强

- 渠道详情页二级 tabs
- 渠道活动流筛选与搜索
- 渠道会话跳转
- 渠道级 cost / usage / rate-limit 视图

## 推荐下一步

1. 先做一个真实渠道打通样板，建议 `Slack` 或 `飞书`
2. 抽象统一 Connector 接口，再接第二个国内渠道与第二个海外渠道
3. 补会话域和消息域，把渠道从“接入状态”推进到“可运营”
