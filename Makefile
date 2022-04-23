REGISTRY?=registry.fly.io

.PHONY: flyauth
flyauth: bin/flyctl
	bin/flyctl auth docker

.PHONY: deploy
deploy: bin/ko flyauth
	$(eval img=$(shell KO_DOCKER_REPO=${REGISTRY} bin/ko publish --sbom none --base-import-paths .))
	bin/flyctl deploy --image ${img}

include Makefile.bindl
