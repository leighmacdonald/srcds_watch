check: fmt lint_golangci static

lint_golangci:
	@golangci-lint run --timeout 3m

static:
	@staticcheck -go 1.20 ./...

check_deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.62.2
	go install honnef.co/go/tools/cmd/staticcheck@latest

docker_build:
	docker build -t leighmacdonald/srcds_watch:latest .

fmt:
	gci write . --skip-generated -s standard -s default
	gofumpt -l -w .

test:
	go test ./...

update:
	go get -u ./...

