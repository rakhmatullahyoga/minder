# Minder
Minder is an application for match-making and dating purpose. This microservice runs on RESTful API protocols.

## Dependencies
- Go 1.20
- MySQL 8.0
- Redis
- Docker (optional)
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [golangci-lint](https://github.com/golangci/golangci-lint)

## Project structure

```
root
|- auth # go package for register, login, and authorization of the service
|  |- error.go # errors definition
|  |- handler.go # http handler
|  |- middleware.go # http middleware for core business process APIs
|  |- model.go # data structure definition
|  |- repo.go # data repository functions
|  |- service.go # business logic
|- bin # additional tools such as db migrate, linter
|- candidate # go package for minder's core business process such as get candidate, swipe and subscription
|  |- error.go # errors definition
|  |- handler.go # http handler
|  |- model.go # data structure definition
|  |- repo.go # data repository functions
|  |- service.go # business logic
|- db # db related scripts
|  |- migrations # migration sql files
|     |- x.down.sql
|     |- x.up.sql
|- docker-compose.yml # system dependencies: database, cache
|- main.go # application entrypoint
```

## How to run
1. Clone this repository
```bash
git clone git@github.com:rakhmatullahyoga/minder.git
```
2. Setup dependencies via Docker (optional)
```bash
docker-compose up -d
```
3. Run database migration to setup database schema. To run the migration, please follow the instruction on the [developer page](https://pkg.go.dev/github.com/golang-migrate/migrate/cli#section-readme)
4. Setup environment variables. You can easily setup environment variables using `.env` file by copying from the `env.sample` file and modifying the `.env` file
```bash
make env
```
5. Run unit test
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o cover.html
```
6. Run linter
```bash
make lint
```
7. Compile the project
```bash
make compile
```
8. Run the application
```bash
./minder
```

## Test the application
You can now test the application by making http request to minder service as described in the attached [Postman collection](minder.postman_collection.json)