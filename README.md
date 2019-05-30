# gohst
a lightweight, (almost) portable file hosting and sharing service

## requirements
- `mysql 5.7+` or `mariadb 10.3+`
- `go 1.11+` (if building from source)

## building
```go install github.com/voidiz/gohst```

## quick start
1. `gohst config create` - Creates the configuration file in your home directory
where at least your database credentials and domain have to be filled in before
you proceed.
1. `gohst config setup` - Performs the initial setup (database etc.).
1. `gohst serve` - Runs the server.

## client usage
See [gup](https://github.com/voidiz/gup) for a basic cli that handles both uploading
and deleting.