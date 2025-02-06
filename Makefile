.PHONY: build push

build: shifu-akri-adapter shifu-akri-rtsp-discovery shifu-akri-s7-discovery

shifu-akri-adapter:
	docker buildx build --no-cache -t edgenesis/shifu-akri-adapter:nightly shifu-akri-adapter --load

shifu-akri-rtsp-discovery:
	docker buildx build --no-cache -t edgenesis/shifu-akri-rtsp-discovery:nightly rtsp-discovery --load

shifu-akri-s7-discovery:
	docker buildx build --no-cache -t edgenesis/shifu-akri-s7-discovery:nightly s7-discovery --load

push:
	docker buildx build --platform linux/amd64,linux/arm64 -t edgenesis/shifu-akri-adapter:nightly shifu-akri-adapter --push
	docker buildx build --platform linux/amd64,linux/arm64 -t edgenesis/shifu-akri-rtsp-discovery:nightly rtsp-discovery --push
	docker buildx build --platform linux/amd64,linux/arm64 -t edgenesis/shifu-akri-s7-discovery:nightly s7-discovery --push
