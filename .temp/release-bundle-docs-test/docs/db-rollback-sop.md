# Platform API DB Rollback SOP

## 适用场景

- migration 执行后出现不可接受的 schema / data 问题
- 发布后 API 虽可启动，但核心业务写入异常
- 必须回退到发布前数据库状态

## 前提

- 发布前已执行数据库备份
- 已记录当前镜像 digest 与上一版本 digest
- 已确认需要人工回滚，而不是仅回滚 Deployment

## 回滚步骤

1. 暂停对外写流量
- 下线入口或临时只读
- 停止会继续写库的后台任务

2. 回滚应用镜像

```powershell
.\deploy\k8s\platform-api\scripts\rollback.ps1
```

3. 评估数据库是否需要恢复
- 仅代码问题：先观察镜像回滚是否恢复
- schema/data 已损坏：进入数据库恢复

4. 恢复数据库

```powershell
.\deploy\k8s\platform-api\scripts\db-restore.ps1 -PostgresPod postgresql-0 -Input .\platform-api-pre-release-backup.sql
```

5. 校验 migration 状态

```powershell
.\deploy\k8s\platform-api\scripts\migration-status.ps1 -DatabaseUrl "postgresql://..."
```

6. 重新执行 smoke

```powershell
.\deploy\k8s\platform-api\scripts\smoke.ps1
```

## 记录项

- 回滚触发时间
- 触发人
- 批准人
- 回滚前镜像 digest
- 回滚后镜像 digest
- 备份文件名 / 存储位置
- 恢复完成时间
- 业务验证结果


