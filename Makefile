REGISTRY?=registry.fly.io

.PHONY: flyauth
flyauth: bin/flyctl
	bin/flyctl auth docker

.PHONY: deploy
deploy: bin/ko flyauth
	$(eval img=$(shell KO_DOCKER_REPO=${REGISTRY} bin/ko publish --sbom none --base-import-paths ./cmd/gomodurl))
	bin/flyctl deploy --image ${img}

.PHONY: gh/lint
gh/lint: bin/golangci-lint
	bin/golangci-lint run --out-format github-actions

include Makefile.bindl
