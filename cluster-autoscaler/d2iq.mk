# D2iQ Release build to push a mesosphere image for the
KUBERNETES_VERSION?=1.16
D2IQ_TAG_VERSION?=v$(KUBERNETES_VERSION)-0.1.0

PKG = k8s.io/autoscaler/cluster-autoscaler

# DOCKER_IMG_TAG is the tag of the builder image based on the go.sum and Dockerfile and additional build-args that are passed, that otherwise don't change the sha of the Dockerfile
GO_MOD_SHASUM := $$(shasum go.sum | awk '{ print $$1 }' | cut -c1-3)
GO_VERSION = 1.13.8
KUBERNETES_VERSION_SHASUM = $$(echo $(KUBERNETES_VERSION) | shasum | awk '{ print $$1 }' | cut -c1-3)
GO_VERSION_SHASUM = $$(echo $(GO_VERSION) | shasum | awk '{ print $$1 }' | cut -c1-3)
DOCKERFILE_SHASUM := $$(shasum ./Dockerfile.ci | awk '{ print $$1 }' | cut -c1-3)
DOCKER_IMG_TAG := $(shell echo $(DOCKERFILE_SHASUM)$(GO_MOD_SHASUM)$(GO_VERSION_SHASUM)$(KUBERNETES_VERSION_SHASUM))
DOCKER_CI_IMG := mesosphere/cluster-autoscaler:$(DOCKER_IMG_TAG)

.PHONY: d2iq-docker-release
d2iq-docker-release:
	TAG=$(D2IQ_TAG_VERSION) GOOS=linux REGISTRY=mesosphere PROVIDER= $(MAKE) build build-binary execute-release


.PHONY: docker-ci.builder
docker-ci.builder:
ifeq ($(shell docker image list -q $(DOCKER_CI_IMG)),)
	docker build                                              \
	    --build-arg KUBERNETES_VERSION=v$(KUBERNETES_VERSION) \
	    -f Dockerfile.ci -t $(DOCKER_CI_IMG) .
endif


.PHONY: gitauth
gitauth:
ifeq ($(CI),yes)
ifdef GITHUB_TOKEN
	@git config --global url."https://$(GITHUB_TOKEN):x-oauth-basic@github.com/mesosphere".insteadOf "https://github.com/mesosphere"
endif
endif

.PHONY: docker-ci.test-unit
docker-ci.test-unit: docker-ci.builder
	echo Running test in a container;                       \
	docker run                                          	  \
			--rm                                                \
			--net=host                                          \
			-e ENVVAR=$(ENVVAR)                                 \
			-e GITHUB_TOKEN=$(GITHUB_TOKEN)                     \
			-e CI="$(CI)"                                       \
			-v "$(shell pwd)":"$(HOME)/src/$(PKG)"              \
			-v "/tmp/":"/tmp"                                   \
	    -w $(HOME)/src/$(PKG)                               \
	    $(DOCKER_CI_IMG)                                     \
	    make gitauth test-unit
