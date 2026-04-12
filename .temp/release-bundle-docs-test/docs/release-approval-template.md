# Platform API Release Approval Template

## 变更基本信息

- 版本 / Tag:
- 镜像 Digest:
- 变更负责人:
- 审批人:
- 发布时间窗口:

## 变更范围

- 是否包含 migration:
- 是否包含 bootstrap:
- 是否包含外部配置变更:
- 是否包含密钥轮换:

## 预检结果

- `go test ./...`:
- `smoke-migrations.ps1`:
- `smoke-persistence.ps1 -Build`:
- `preflight.ps1 -Overlay prod -Strict`:
- `validate.ps1 -Overlay prod`:
- `validate.ps1 -Overlay prod` for bootstrap:

## 回滚准备

- 已完成数据库备份:
- 备份文件位置:
- 上一版本镜像 digest:
- 回滚负责人:

## 风险评估

- 风险等级:
- 主要风险:
- 缓解措施:

## 审批结论

- 批准 / 驳回:
- 审批备注:

