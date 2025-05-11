Start-Process powershell -ArgumentList "go run ./api_gateway/cmd/main.go"
Start-Process powershell -ArgumentList "go run ./playlistService/cmd/main.go"
Start-Process powershell -ArgumentList "go run ./track-service/cmd/main.go"
Start-Process powershell -ArgumentList "go run ./userService/cmd/main.go"
