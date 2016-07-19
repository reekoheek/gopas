default: build

build:
	@echo "==> Bootstrap build"
	gopas build

build-with-self: build
	@echo "==> Build self"
	.gopath/bin/gopas build

#install: gopas
#	mv gopas $(GOPATH)/bin

#test:
#	gopas test
