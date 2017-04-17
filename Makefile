up:
	minikube start
	wait 5
	helm init
	wait 5
	helm install -n rk-test redkeep-api

down:
	helm delete --purge rk-test
	helm reset
	minikube stop

open:
	minikube service rk-test-redkeep-api

new:
	REDKEEP_HOST=$(shell minikube service --url rk-test-redkeep-api) go run cmd/redkeep.go new

build: build-base build-api

build-base:
	docker build -f Dockerfile-base -t nicr9/redkeep-base:latest .

build-api:
	docker build -f Dockerfile-api -t nicr9/redkeep-api:latest .
