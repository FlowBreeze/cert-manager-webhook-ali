IMAGE_NAME := docker.rainbow-systems.cloud/rainbow-cicd/cert-manager-webhook-ali
IMAGE_TAG := $(shell cat VERSION)

#test:
#	go test -v .

.PHONY: build-debug
build-debug:
	pack build $(IMAGE_NAME):$(IMAGE_TAG) \
	--path . \
	--builder docker.rainbow-systems.cloud/rainbow-cicd/builder:0.0.2 \
	--pull-policy=if-not-present \
	--env RB_BUILD_WITH_DEBUG=true

.PHONY: build
build:
	pack build $(IMAGE_NAME):$(IMAGE_TAG) \
	--path . \
	--builder docker.rainbow-systems.cloud/rainbow-cicd/builder:0.0.2 \
	--pull-policy=if-not-present

debug: build-debug
	docker run \
    --entrypoint debug \
    -p 2345:2345 \
    --env GROUP_NAME=alidns.certmanager.hook \
    $(IMAGE_NAME):$(IMAGE_TAG) \
    --tls-cert-file=/workspace/testdata/run/ca.crt \
    --tls-private-key-file=/workspace/testdata/run/ca.key \
    --secure-port=6443

run: build
	docker run \
    --env GROUP_NAME=alidns.certmanager.hook \
    $(IMAGE_NAME):$(IMAGE_TAG) \
    --tls-cert-file=/workspace/testdata/run/ca.crt \
    --tls-private-key-file=/workspace/testdata/run/ca.key \
    --secure-port=6443

release: build
	docker push $(IMAGE_NAME):$(IMAGE_TAG)

save:
	docker save $(IMAGE_NAME):$(IMAGE_TAG) > save.tar
