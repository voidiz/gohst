# gohst
a lightweight, (almost) portable file server

## requirements
- `mysql 5.7+` or `mariadb 10.3+`
- `go 1.11` or higher (if building from source)

## building
```go install github.com/voidiz/gohst```

## quick start
1. `gohst config create` - Creates the configuration file in your home directory where at least your database credentials and domain have to be filled in.
1. `gohst config setup` - Performs the initial setup (database etc.).
1. `gohst serve` - Run the server.