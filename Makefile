DOCKER_REPO = ghcr.io/dogmatiq/proclaim
DOCKER_PLATFORMS += linux/amd64
DOCKER_PLATFORMS += linux/arm64

-include .makefiles/Makefile
-include .makefiles/pkg/go/v1/Makefile
-include .makefiles/pkg/go/v1/with-ferrite.mk
-include .makefiles/pkg/docker/v1/Makefile
-include .makefiles/pkg/helm/v1/Makefile

.PHONY: run
run: $(GO_DEBUG_DIR)/proclaim
	DEBUG=true $<

.makefiles/%:
	@curl -sfL https://makefiles.dev/v1 | bash /dev/stdin "$@"
