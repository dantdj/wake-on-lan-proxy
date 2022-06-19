.PHONY: deploy
deploy:
	ssh -i /Users/danielt/.ssh/nginx_pi pi@192.168.2.34 'if pgrep wol-proxy; then pkill wol-proxy; fi && rm wol-proxy'
	GOOS=linux GOARCH=arm GOARM=7 go build -o ./bin/linux_arm/wol-proxy ./cmd/server
	scp -i /Users/danielt/.ssh/nginx_pi ./bin/linux_arm/wol-proxy pi@192.168.2.34:~/wol-proxy
	scp -i /Users/danielt/.ssh/nginx_pi ./.env pi@192.168.2.34:~/.env
	ssh -i /Users/danielt/.ssh/nginx_pi pi@192.168.2.34 './wol-proxy'

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