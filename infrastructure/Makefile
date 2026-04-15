ENV         ?= dev
LAYER       ?= 20-app
PROJECT     ?= $(shell gcloud config get-value project 2>/dev/null)
REGION      ?= us-central1
STATE_BUCKET ?= $(PROJECT)-tf-state-$(ENV)

TF_DIR      = terraform/envs/$(ENV)/$(LAYER)
TF_FLAGS    = -var="project_id=$(PROJECT)" -var="region=$(REGION)"
TF_BACKEND  = -backend-config="bucket=$(STATE_BUCKET)"

.PHONY: init plan apply destroy fmt validate addons

init:
	terraform -chdir=$(TF_DIR) init $(TF_BACKEND)

plan: init
	terraform -chdir=$(TF_DIR) plan $(TF_FLAGS)

apply: init
	terraform -chdir=$(TF_DIR) apply $(TF_FLAGS)

destroy: init
	terraform -chdir=$(TF_DIR) destroy $(TF_FLAGS)

fmt:
	terraform fmt -recursive terraform/

validate:
	@for dir in terraform/modules/*/; do \
		echo "validating $$dir"; \
		terraform -chdir=$$dir validate; \
	done

addons:
	./cluster-addons/install.sh $(ENV)
