# RuoYi-Vue-GO

RuoYi 管理系统 Go 版服务端（Gin + GORM），复刻 Java 版接口契约，前端 RuoYi-Vue3 零改动可切换。规范与任务台账见 [AGENTS.md](./AGENTS.md) 与 [specs/README.md](./specs/README.md)。

## 快速开始

```bash
go run ./cmd/server --env=dev    # 监听 8080，配置见 configs/.env.dev
```

## 提交前自查（必做）

```bash
gofmt -l .          # 输出必须为空（先 gofmt -w . 修正）
go vet ./...        # 零告警
go test ./...       # 全绿
```

或直接执行 `make lint`（macOS/Linux）。

## lint 工具决策

初期以 `gofmt + go vet` 为门禁，暂不引入 golangci-lint（避免额外工具链安装负担）；当模块增多后再评估引入，届时在 `specs/deviations.md` 记录。
