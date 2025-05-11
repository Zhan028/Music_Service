Start-Job { Set-Location -Path "./api_gateway"; go run ./cmd }
Start-Job { Set-Location -Path "./playlistService"; go run ./cmd }
Start-Job { Set-Location -Path "./track-service"; go run ./cmd }
Start-Job { Set-Location -Path "./userService"; go run ./cmd }
