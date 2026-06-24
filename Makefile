FNAME_SRC := main.go

APP_NAME := dootBot
CGO_ENABLED := 0

export GOOS
export GOARCH
export CGO_ENABLED

.PHONY: windows linux rpi mac docker build run b r clean

# Windows 64-bit
windows: GOOS := windows
windows: GOARCH := amd64
windows: FNAME_OUT := $(APP_NAME).exe
windows: build

# Raspberry Pi 64-bit -- Change arm64 to amd64 for other linux systems
linux: GOOS := linux
linux: GOARCH := arm64
linux: FNAME_OUT := $(APP_NAME)
linux: build

build:
	$(info Building $(FNAME_OUT) for $(GOOS)/$(GOARCH))
	go build -o ./$(FNAME_OUT) ./$(FNAME_SRC)

docker:
	docker build -t discord-bot .

run:
	go run ./$(FNAME_SRC)

b: windows

r: run

clean:
	del /Q $(APP_NAME).exe $(APP_NAME) 2>NUL || exit 0