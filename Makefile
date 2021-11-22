serve:
	go run cmd/server/main.go

run:
	go run cmd/client/main.go -token=XXXXXXX

test:
	go test ./...