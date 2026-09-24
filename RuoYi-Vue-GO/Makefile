# RuoYi-Vue-GO 工具链（run / test / lint）
# Windows 无 make 时直接执行对应 go 命令（见 README 提交前自查）。

run:
	go run ./cmd/server --env=dev

test:
	go test ./...

lint:
	gofmt -l .
	go vet ./...

fmt:
	gofmt -w .

.PHONY: run test lint fmt
