UNAME := $(shell uname)

# Default mysql migration settings
export MYSQL_USERNAME ?= root
export MYSQL_PASSWORD ?= rootpw
export MYSQL_HOST     ?= localhost
export MYSQL_PORT     ?= 3306
export MYSQL_DATABASE ?= minder

tidy:
	go mod tidy

clean:
	rm -f minder

compile: clean
	go build -o minder main.go

env:
	cp env.sample .env

bin:
	@mkdir -p bin

bin/golangci-lint: bin
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.55.2

lint: bin/golangci-lint
	./bin/golangci-lint run -v

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

# If the first argument is "migrate"...
ifeq (migrate,$(firstword $(MAKECMDGOALS)))
  # use the rest as arguments for "migrate"
  MIGRATE_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(MIGRATE_ARGS):;@:)
endif

migrate: bin/migrate
	./bin/migrate -source file://db/migrations -database "mysql://$(MYSQL_USERNAME):$(MYSQL_PASSWORD)@tcp($(MYSQL_HOST):$(MYSQL_PORT))/$(MYSQL_DATABASE)" $(MIGRATE_ARGS)
