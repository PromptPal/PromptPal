release: export GOOS=linux
release: export GOARCH=amd64

release:
	go build -ldflags="-s -w -X main.GitCommit=${GITHUB_SHA}" -tags release -o up-pp-api main.go