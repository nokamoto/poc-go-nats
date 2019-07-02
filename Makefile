fmt:
	gofmt -d .
	gofmt -w .

test: fmt
	go vet .
	go test .
	go mod tidy
	docker-compose down
	docker-compose build

mod:
	rm -f go.mod go.sum
	go mod init github.com/nokamoto/poc-go-nats
	go install .
	go mod tidy
