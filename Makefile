.PHONY: \
  app-build app-test app-fmt app-vet app-tidy \
  console-build console-install \
  kind-up kind-down deploy-dev tilt \
  infra-plan infra-apply infra-fmt \
  lint

# ── Backend (monitoring-app) ───────────────────────────────────────────────────

app-build:
	$(MAKE) -C monitoring-app build

app-test:
	$(MAKE) -C monitoring-app test

app-fmt:
	$(MAKE) -C monitoring-app fmt

app-vet:
	$(MAKE) -C monitoring-app vet

app-tidy:
	$(MAKE) -C monitoring-app tidy

# ── Console (monitoring-console) ──────────────────────────────────────────────

console-install:
	cd monitoring-console && npm ci

console-build:
	cd monitoring-console && npm run build

# ── Local dev cluster ─────────────────────────────────────────────────────────

kind-up:
	$(MAKE) -C monitoring-app kind-up

kind-down:
	$(MAKE) -C monitoring-app kind-down

deploy-dev:
	$(MAKE) -C monitoring-app deploy-dev

tilt:
	cd monitoring-app && tilt up

# ── Infrastructure ────────────────────────────────────────────────────────────

ENV   ?= dev
LAYER ?= 20-app

infra-plan:
	$(MAKE) -C infrastructure plan ENV=$(ENV) LAYER=$(LAYER)

infra-apply:
	$(MAKE) -C infrastructure apply ENV=$(ENV) LAYER=$(LAYER)

infra-fmt:
	$(MAKE) -C infrastructure fmt

# ── All checks (run in CI locally) ────────────────────────────────────────────

lint: app-vet console-build infra-fmt
