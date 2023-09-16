fNameSrc = main.go
fNameOut = main.exe

# 64 bit x86 setting
.64Bitx86:
	go env -w GOARCH=amd64

# Windows settings
.windows:
	go env -w GOOS=windows
	$(eval fNameOut = main.exe)

# Linux settings
.linux:
	go env -w GOOS=linux
	$(eval fNameOut = main)

# Mac OS X 10.8 and above
.mac:
	go env -w GOOS=darwin
	$(eval fNameOut = main.app)

# Default value, change for your needs
.default: .64Bitx86 .windows


# Build for 64 bit windows (default)
windows: .default build

# Build for 64 bit linux
linux: .64Bitx86 .linux build .default

# Build for 64 bit mac
mac: .64Bitx86 .mac build .default

docker:
	docker build -t discord-bot .

# Builder command
build: .default
	set CGO_ENABLED=1
	go build -o ./${fNameOut} ./${fNameSrc}

b: build

# Runs go.main
run: .default
	go run ./${fNameSrc}

r: run