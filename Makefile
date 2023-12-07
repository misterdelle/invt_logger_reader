build-arm:
	env GOOS=linux GOARCH=arm GOARM=5 go build -o invt-arm

build:
	go build -o invt-arm