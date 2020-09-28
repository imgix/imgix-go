test: deps fmt
	cd ./v2 && go test -cover
	go mod tidy

deps:
	go get github.com/stretchr/testify golang.org/x/tools/cmd/cover

fmt:
	cd ./v2 go fmt
