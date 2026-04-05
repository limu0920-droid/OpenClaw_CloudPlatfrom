# OpenClaw Web

单应用承载两套界面：

- `Portal`：租户用户视角
- `Admin`：平台运维与审计视角

## Development

```bash
npm install
npm run dev
```

默认通过 Vite 代理访问 Go Mock API：

```text
/api -> http://localhost:8080
```

如果需要手工指定后端地址，可配置：

```bash
VITE_API_BASE_URL=http://localhost:8080
```

## Build

```bash
npm run build
```
