# DO NOT EDIT. Generated with:
#
#    devctl@4.2.1
#

.PHONY: lint-chart
## lint-chart: runs ct against the default chart
lint-chart: IMAGE := giantswarm/helm-chart-testing:v3.0.0-rc.1
lint-chart:
	@echo "====> $@"
	rm -rf /tmp/$(APPLICATION)-test
	mkdir -p /tmp/$(APPLICATION)-test/helm
	cp -a ./helm/$(APPLICATION) /tmp/$(APPLICATION)-test/helm/
	architect helm template --dir /tmp/$(APPLICATION)-test/helm/$(APPLICATION)
	docker run -it --rm -v /tmp/$(APPLICATION)-test:/wd --workdir=/wd --name ct $(IMAGE) ct lint --validate-maintainers=false --charts="helm/$(APPLICATION)"
	rm -rf /tmp/$(APPLICATION)-test
