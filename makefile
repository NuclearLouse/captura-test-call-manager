.PHONY: build
build:
	go build -v -o CaptTestCallsSrvc.exe 

.PHONY: linux
linux:
	SET GOOS=linux 
	SET GOARCH=amd64 
	go build -o CaptTestCallsSrvc -v ./
	SET GOOS=windows


DEFAULT_GOAL := build