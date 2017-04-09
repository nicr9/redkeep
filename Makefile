install:
	rm redkeep-api-*.tgz
	helm package redkeep-api
	helm install redkeep-api-*.tgz

build-base:
	docker build -f Dockerfile-base -t nicr9/redkeep-base:latest .

build-api:
	docker build -f Dockerfile-api -t nicr9/redkeep-api:latest .
