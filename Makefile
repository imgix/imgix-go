test: deps fmt
	cd ./v2 && go test -cover
	go mod tidy

fmt:
	cd ./v2 go fmt
