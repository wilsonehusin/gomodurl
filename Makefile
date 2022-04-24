REGISTRY?=ghcr.io/wilsonehusin

.PHONY: deploy
deploy: bin/flyctl
	ifndef IMAGE
		$(error IMAGE is not set)
	endif
	bin/flyctl deploy --image ${IMAGE}

.PHONY: koauth
koauth: bin/ko
	ifndef PASSWORD
		$(error PASSWORD is not set)
	endif
	@ko login ${REGISTRY} --username "gomodurl" --password ${PASSWORD}

.PHONY: publish-container
publish-container: bin/ko
	KO_DOCKER_REPO=${REGISTRY} ko publish --base-import-paths ./cmd/gomodurl

.PHONY: gh/lint
gh/lint: bin/golangci-lint
	bin/golangci-lint run --out-format github-actions

include Makefile.bindl
