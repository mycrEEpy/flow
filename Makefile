test:
	go test -race -v $(go list ./... | grep -v /examples/)
