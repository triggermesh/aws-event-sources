TARGET=awss3

all: clean fmt vet

clean:
	rm -rf $(TARGET)

fmt:
	go fmt ./...

lint:
	golint ./...

vet:
	go vet ./...