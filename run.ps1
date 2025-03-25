Get-Content .env | ForEach-Object {
    if ($_ -match "^\s*([^#]\S*)=(\S*)$") {
        [System.Environment]::SetEnvironmentVariable($matches[1], $matches[2])
    }
}
go run cmd/main.go