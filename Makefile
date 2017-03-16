.PHONY: verify-glide-installation install-vendor update-vendor all install check

all:
	go generate github.com/giantswarm/aws-operator/bindata
	go build github.com/giantswarm/aws-operator

verify-glide-installation:
	@which glide || go get github.com/Masterminds/glide
	@which glide-vc || go get github.com/sgotti/glide-vc

install-vendor: verify-glide-installation
	glide install --strip-vendor
	glide-vc --use-lock-file

update-vendor: verify-glide-installation
	glide update --strip-vendor
	glide-vc --use-lock-file

verify-go-bindata-installation:
	@which go-bindata || go get -u github.com/jteeuwen/go-bindata/...

check:
	go test `glide novendor`
