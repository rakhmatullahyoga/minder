tidy:
	go mod tidy

clean:
	rm -f goodies

compile: clean
	go build -o goodies main.go

env:
	cp env.sample .env

bin:
	@mkdir -p bin

bin/golangci-lint: bin
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.43.0

bin/migrate: bin
ifeq ($(UNAME), Linux)
	@curl -sSfL https://github.com/golang-migrate/migrate/releases/download/v4.15.1/migrate.linux-amd64.tar.gz | tar zxf - --directory /tmp && cp /tmp/migrate bin/
else ifeq ($(UNAME), Darwin)
ifeq ($(ARCH), arm64) # for Apple processor macs
	@curl -sSfL https://github.com/golang-migrate/migrate/releases/download/v4.15.1/migrate.darwin-arm64.tar.gz | tar zxf - --directory /tmp && cp /tmp/migrate bin/
else
	@curl -sSfL https://github.com/golang-migrate/migrate/releases/download/v4.15.1/migrate.darwin-amd64.tar.gz | tar zxf - --directory /tmp && cp /tmp/migrate bin/
endif
else
	@echo "Your OS is not supported."
endif
