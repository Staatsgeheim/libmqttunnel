.PHONY: build

build:
	go build -C lib/libmqttunnel -o ../../libmqttunnel -buildmode=c-shared
