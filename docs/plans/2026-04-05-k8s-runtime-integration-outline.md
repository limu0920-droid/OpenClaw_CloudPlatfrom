# K8s Runtime Integration Outline

## 目标

把当前 Mock 控制面逐步替换成真实的 Kubernetes 控制面接入。

## 后端建议结构

### 1. Runtime Adapter

- 统一接口：
  - `createInstance`
  - `startInstance`
  - `stopInstance`
  - `restartInstance`
  - `deleteInstance`
  - `getRuntimeUnits`
  - `getRuntimeMetrics`
  - `getRuntimeLogs`
  - `openTerminalSession`

### 2. K8s 适配层

- 调用对象：
  - `Namespace`
  - `Deployment`
  - `Service`
  - `Ingress` / `Gateway`
  - `ConfigMap`
  - `Secret`
  - `PVC`
- 权限边界：
  - 普通前端不直接接触 K8s
  - 平台后端通过受控 `ServiceAccount`
  - 管理员高风险动作需要审计与审批

### 3. 指标与日志

- `Prometheus`
  - CPU
  - Memory
  - Disk
  - Network
  - Request / Error / Latency
- `OpenSearch`
  - 审计
  - 渠道活动
  - 运行日志
  - 工单上下文

## 控制台映射

### Portal

- 用户只能看到自己的实例资源与运行态
- 可执行：
  - 启停
  - 重启
  - 发布配置
  - 备份

### Admin

- 管理员可看到：
  - 集群
  - 节点
  - Pod
  - Deployment
  - Service
  - 资源趋势
  - 运行日志

## 建议落地顺序

1. 先接 `Prometheus` 指标
2. 再接 `OpenSearch` 日志
3. 再接 `Headlamp` 或自定义 K8s 资源视图
4. 最后把启停/重启等动作从 Mock API 切到真实 Runtime Adapter
