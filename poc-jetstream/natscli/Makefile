.PHONY: build
build:
	GO111MODULE=on CGO_ENABLED=0 go build -o natscli ./nats


.PHONY: image
image:
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -o natscli ./nats
	docker image build -f Dockerfile -t pxsalehi/natscli:latest .

.PHONEY: push-image
push-image:
	docker push pxsalehi/natscli:latest