PKG := ./...
SERVICES := collector ingestor api
REGISTRY ?= localhost:5001
TAG ?= dev
CLUSTER := monitoring-dev

.PHONY: build test fmt vet tidy clean \
        docker-build kind-up kind-down kind-load \
        helm-deps deploy-dev uninstall-dev tilt

build:
	@for svc in $(SERVICES); do \
		go build -o bin/$$svc ./cmd/$$svc; \
	done

test:
	go test $(PKG)

fmt:
	go fmt $(PKG)

vet:
	go vet $(PKG)

tidy:
	go mod tidy

clean:
	rm -rf bin

docker-build:
	@for svc in $(SERVICES); do \
		docker build -t $(REGISTRY)/$$svc:$(TAG) -f deploy/docker/$$svc.Dockerfile .; \
	done

kind-up:
	./scripts/kind-up.sh $(CLUSTER)

kind-down:
	./scripts/kind-down.sh $(CLUSTER)

kind-load: docker-build
	@for svc in $(SERVICES); do \
		kind load docker-image $(REGISTRY)/$$svc:$(TAG) --name $(CLUSTER); \
	done

helm-deps:
	helm dependency update deploy/helm/charts/monitoring-app-dev

deploy-dev: kind-load helm-deps
	helm upgrade --install dev deploy/helm/charts/monitoring-app-dev \
		-f deploy/helm/values/dev.yaml \
		--set collector.image.repository=$(REGISTRY)/collector \
		--set collector.image.tag=$(TAG) \
		--set api.image.repository=$(REGISTRY)/api \
		--set api.image.tag=$(TAG) \
		--create-namespace --namespace monitoring

uninstall-dev:
	helm uninstall dev -n monitoring || true

tilt:
	tilt up
