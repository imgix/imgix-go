test: deps fmt
	go test -cover

deps:
	go get github.com/stretchr/testify golang.org/x/tools/cmd/cover
	go get github.com/joho/godotenv
fmt:
	go fmt
