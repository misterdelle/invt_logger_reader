set GOOS=linux
set GOARCH=amd64
go build -ldflags "-s -w" -o build/invt-logger-reader .
