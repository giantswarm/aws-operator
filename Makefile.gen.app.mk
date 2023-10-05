# DO NOT EDIT. Generated with:
#
#    devctl@6.13.0
#

##@ App

YQ=docker run --rm -u $$(id -u) -v $${PWD}:/workdir mikefarah/yq:4.29.2
HELM_DOCS=docker run --rm -u $$(id -u) -v $${PWD}:/helm-docs jnorwood/helm-docs:v1.11.0

ifdef APPLICATION
DEPS := $(shell find $(APPLICATION)/charts -maxdepth 2 -name "Chart.yaml" -printf "%h\n")
endif

.PHONY: lint-chart check-env update-chart helm-docs update-deps $(DEPS)

lint-chart: IMAGE := giantswarm/helm-chart-testing:v3.0.0-rc.1
lint-chart: check-env ## Runs ct against the default chart.
	@echo "====> $@"
	rm -rf /tmp/$(APPLICATION)-test
	mkdir -p /tmp/$(APPLICATION)-test/helm
	cp -a ./helm/$(APPLICATION) /tmp/$(APPLICATION)-test/helm/
	architect helm template --dir /tmp/$(APPLICATION)-test/helm/$(APPLICATION)
	docker run -it --rm -v /tmp/$(APPLICATION)-test:/wd --workdir=/wd --name ct $(IMAGE) ct lint --validate-maintainers=false --charts="helm/$(APPLICATION)"
	rm -rf /tmp/$(APPLICATION)-test

update-chart: check-env ## Sync chart with upstream repo.
	@echo "====> $@"
	vendir sync
	$(MAKE) update-deps

update-deps: check-env $(DEPS) ## Update Helm dependencies.
	cd $(APPLICATION) && helm dependency update

$(DEPS): check-env ## Update main Chart.yaml with new local dep versions.
	dep_name=$(shell basename $@) && \
	new_version=`$(YQ) .version $(APPLICATION)/charts/$$dep_name/Chart.yaml` && \
	$(YQ) -i e "with(.dependencies[]; select(.name == \"$$dep_name\") | .version = \"$$new_version\")" $(APPLICATION)/Chart.yaml

helm-docs: check-env ## Update $(APPLICATION) README.
	$(HELM_DOCS) -c $(APPLICATION) -g $(APPLICATION)

check-env:
ifndef APPLICATION
	$(error APPLICATION is not defined)
endif
