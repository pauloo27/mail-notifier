BINARY_NAME = mail-notifier

build:
	go build -o $(BINARY_NAME) -v ./gui

dev:
	fiber dev -r . -t ./gui

run: build
	./$(BINARY_NAME) 

install: build
	sudo cp ./$(BINARY_NAME) /usr/bin/

tidy:
	go mod tidy

# (build but with a smaller binary)
dist:
	go build -ldflags="-w -s" -gcflags=all=-l -v

# (even smaller binary)
pack: dist
	upx ./$(BINARY_NAME)
