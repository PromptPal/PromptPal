release: export GOOS=linux
release: export GOARCH=amd64

release:
	go build -ldflags="-linkmode external -extldflags '-static' -s -w -X main.GitCommit=${GITHUB_SHA}" -tags release,musl -o up-pp-api main.go