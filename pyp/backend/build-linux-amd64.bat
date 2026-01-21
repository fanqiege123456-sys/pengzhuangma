@echo off
setlocal
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
go build -o collision-backend-linux
endlocal
