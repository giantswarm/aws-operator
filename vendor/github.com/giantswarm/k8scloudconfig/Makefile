
.PHONY: verify-glide-installation install-vendor update-vendor all install check

all:
	go build github.com/giantswarm/k8scloudconfig

verify-glide-installation:
	@which glide || go get github.com/Masterminds/glide
	@which glide-vc || go get github.com/sgotti/glide-vc

install-vendor: verify-glide-installation
	glide install --strip-vendor
	glide-vc --use-lock-file

update-vendor: verify-glide-installation
	glide update --strip-vendor
	glide-vc --use-lock-file

check:
	go test `glide novendor`
