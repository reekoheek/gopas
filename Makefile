default: build

build:
	gopas build

install: gopas
	mv gopas $(GOPATH)/bin

test:
	gopas test
