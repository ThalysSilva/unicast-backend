Get-Content .env | ForEach-Object {
    if ($_ -match "^\s*([^#]\S*)=(\S*)$") {
        [System.Environment]::SetEnvironmentVariable($matches[1], $matches[2])
    }
}
swag init -g cmd/main.go
go run cmd/main.go