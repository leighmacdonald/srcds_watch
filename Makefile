check: lint_golangci lint_vet lint_imports lint_cyclo lint_golint static

lint_golangci:
	@golangci-lint run --timeout 3m

lint_vet:
	@go vet -tags ci ./...

lint_imports:
	@test -z $(goimports -e -d . | tee /dev/stderr)

lint_cyclo:
	@gocyclo -over 40 .

lint_golint:
	@golint -set_exit_status $(go list -tags ci ./...)

static:
	@staticcheck -go 1.20 ./...

check_deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.51.2
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	go install golang.org/x/lint/golint@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest

docker_run:
	docker build -t leighmacdonald/srcds_watch:latest .
	docker run -v $(pwd)/srcds_watch.yml:/app/srcds_watch.yml -it leighmacdonald/srcds_watch:latest

test:
	go test ./...