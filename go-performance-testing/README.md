# Go 性能测试示例项目

本项目演示了如何使用 Go 语言的性能测试工具和第三方工具进行性能测试和分析。

## 项目结构

```
go-performance-testing/
├── cmd/
│   ├── profile/        # 使用 pprof 进行性能分析的示例
│   └── server/         # 用于负载测试的 HTTP 服务器
├── internal/
│   └── calculator/     # 包含待测试的计算函数
└── testdata/           # 测试数据
```

## 性能测试工具

### 1. Go 标准测试工具

Go 内置了性能测试工具，可以通过 `go test -bench` 命令运行基准测试。

#### 运行基准测试

```bash
cd go-performance-testing
go test -bench=. ./internal/calculator/
```

#### 查看内存分配

```bash
go test -bench=. -benchmem ./internal/calculator/
```

### 2. pprof 性能分析

pprof 是 Go 内置的性能分析工具，可以分析 CPU、内存、 goroutine 等使用情况。

#### 运行 pprof 服务器

```bash
cd go-performance-testing
go run cmd/profile/main.go
```

然后访问 http://localhost:6060/debug/pprof/ 查看性能数据。

#### 生成 CPU 分析报告

```bash
# 采集 30 秒的 CPU 数据
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# 在 pprof 交互式命令中生成报告
(pprof) top10
(pprof) web  # 生成 SVG 图并在浏览器中打开
```

### 3. Vegeta 负载测试

Vegeta 是一个功能强大的 HTTP 负载测试工具。

#### 安装 Vegeta

```bash
go install github.com/tsenart/vegeta@latest
```

#### 创建负载测试配置

创建 `load-test.txt` 文件：

```
GET http://localhost:8080/fib?n=20
```

#### 运行负载测试

```bash
# 启动服务器
go run cmd/server/main.go

# 另开一个终端运行负载测试
echo "GET http://localhost:8080/fib?n=20" | vegeta attack -duration=30s -rate=10/s | vegeta report
```

### 4. 监控工具

#### Prometheus + Grafana

可以使用 Prometheus 收集指标，Grafana 进行可视化。

1. 安装 Prometheus 和 Grafana
2. 配置 Prometheus 采集 Go 应用的指标
3. 在 Grafana 中创建仪表板

## 示例命令

### 运行所有基准测试

```bash
go test -bench=. ./...
```

### 运行特定基准测试

```bash
go test -bench=BenchmarkFibonacci ./internal/calculator/
```

### 生成 CPU 分析文件

```bash
go test -bench=BenchmarkFibonacci -cpuprofile=cpu.prof ./internal/calculator/
go tool pprof cpu.prof
```

### 生成内存分析文件

```bash
go test -bench=BenchmarkFibonacci -memprofile=mem.prof ./internal/calculator/
go tool pprof mem.prof
```

## 结论

通过本项目，你可以了解如何：
1. 使用 Go 标准测试工具进行基准测试
2. 使用 pprof 进行性能分析
3. 使用 Vegeta 进行负载测试
4. 配置监控工具监控应用性能

这些工具和技术可以帮助你识别和解决 Go 应用中的性能问题，提高应用的响应速度和稳定性。