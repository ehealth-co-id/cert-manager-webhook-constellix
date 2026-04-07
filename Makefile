IMAGE_NAME := cert-manager-webhook-constellix
IMAGE_TAG := latest
REPO_NAME := bpicio

OUT := $(shell pwd)/_out

$(shell mkdir -p "$(OUT)")

.PHONY: all build buildx buildx-multi tag push helm rendered-manifest.yaml

all: ;

# When Go code changes, we need to update the Docker image
build:
	docker build -t "$(IMAGE_NAME):$(IMAGE_TAG)" .

# Buildx (amd64, loads into local Docker). For multi-arch push use: make buildx-multi REPO_NAME=...
buildx:
	docker buildx build --platform linux/amd64 -t "$(IMAGE_NAME):$(IMAGE_TAG)" --load .

# Multi-arch build and push (no --load; set REPO_NAME/IMAGE_TAG as needed)
buildx-multi:
	docker buildx build --platform linux/amd64,linux/arm64 -t "$(REPO_NAME)/$(IMAGE_NAME):$(IMAGE_TAG)" --push .

tag:
	docker tag "$(IMAGE_NAME):$(IMAGE_TAG)" "$(REPO_NAME)/$(IMAGE_NAME):$(IMAGE_TAG)"

push:
	docker push "$(REPO_NAME)/$(IMAGE_NAME):$(IMAGE_TAG)"


# When helm chart changes, we need to publish to the repo (/docs/):
#
# Ensure version is updated in Chart.yaml
# Run `make helm`
# Check and commit the results, including the tar.gz
helm:
	helm package deploy/$(IMAGE_NAME)/ -d docs/
	helm repo index docs --url https://bpicio.github.io/cert-manager-webhook-constellix --merge docs/index.yaml

rendered-manifest.yaml:
	helm template \
		$(IMAGE_NAME) \
		--set image.repository=$(IMAGE_NAME) \
		--set image.tag=$(IMAGE_TAG) \
		deploy/$(IMAGE_NAME) > "$(OUT)/rendered-manifest.yaml"
