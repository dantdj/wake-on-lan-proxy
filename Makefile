.PHONY: deploy
deploy:
	ssh -i /Users/danielt/.ssh/pi_key pi@192.168.2.36 'if pgrep wol-proxy; then pkill wol-proxy; fi && rm wol-proxy'
	GOOS=linux GOARCH=arm GOARM=7 go build -o ./bin/linux_arm/wol-proxy ./cmd/server
	scp -i /Users/danielt/.ssh/pi_key ./bin/linux_arm/wol-proxy pi@192.168.2.36:~/wol-proxy
	scp -i /Users/danielt/.ssh/pi_key ./.env pi@192.168.2.36:~/.env
	ssh -i /Users/danielt/.ssh/pi_key pi@192.168.2.36 './wol-proxy'

.PHONY: build
build:
	go build -o ./bin/darwin_arm64/wol-proxy ./cmd/server

.PHONY: run
run:
	go run ./...

.PHONY: audit
audit:
	@echo "Tidying and verifying module dependencies..."
	go mod tidy
	go mod verify
	@echo "Formatting code..."
	go fmt ./...
	@echo "Vetting code..."
	go vet ./...
	@echo "Running tests..."
	go test -race -vet=off ./...