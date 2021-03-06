# gohst
a lightweight, (almost) portable file hosting and sharing service

## requirements
tested with
- `mysql 5.7+` or `mariadb 10.1+`
- `go 1.11+` (if building from source)

## building and installing
```go get github.com/voidiz/gohst```

## prebuilt binaries
https://github.com/voidiz/gohst/releases

Substitute `gohst` with the path to the executable and run the 
[quick start](#quick-start) commands in cmd/PowerShell. 
Alternatively, add the folder where the executable is located to your 
PATH environment variable and proceed as usual.

## quick start
1. `gohst config create` - Creates the configuration file in your home directory
where at least your database credentials and domain have to be filled in before
you proceed.
1. `gohst config setup` - Performs the initial setup (database etc.).
1. `gohst account create <account_name>` - Creates an account and
generates a random password.
1. `gohst serve` - Runs the server.

## client usage
See [gup](https://github.com/voidiz/gup) for a basic cli that handles both uploading
and deleting.
