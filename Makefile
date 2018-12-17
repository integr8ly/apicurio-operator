SHELL = /bin/bash
REG = quay.io
ORG = integreatly
IMAGE = apicurio-operator
TAG = latest
RESOURCES_DIR = ./res
DEPLOY_DIR = deploy
OUT_STATIC_DIR = build/_output
OUTPUT_BIN_NAME = ${IMAGE}
TARGET_BIN = cmd/manager/main.go
NS = apicurio-operator-test
TEST_FOLDER = ./test/e2e
TEST_POD_NAME = apicurio-operator-test
KC_HOST =
APPS_HOST =

.PHONY: setup/dep
setup/dep:
	@echo Installing golang dependencies
	@go get golang.org/x/sys/unix
	@go get golang.org/x/crypto/ssh/terminal
	@go get -u github.com/gobuffalo/packr/packr
	@echo Installing dep
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
	@echo setup complete

.PHONY: setup/travis
setup/travis:
	@echo Installing Operator SDK
	@curl -Lo operator-sdk https://github.com/operator-framework/operator-sdk/releases/download/v0.1.0/operator-sdk-v0.1.0-x86_64-linux-gnu && chmod +x operator-sdk && sudo mv operator-sdk /usr/local/bin/

.PHONY: code/run
code/run:
	@operator-sdk up local --namespace=${NAMESPACE}

.PHONY: code/compile
code/compile:
	@packr
	@go build -o ${OUTPUT_BIN_NAME} ${TARGET_BIN}
	@packr clean

.PHONY: code/gen
code/gen:
	@operator-sdk generate k8s

.PHONY: code/check
code/check:
	@diff -u <(echo -n) <(gofmt -d `find . -type f -name '*.go' -not -path "./vendor/*"`)

.PHONY: code/fix
code/fix:
	@gofmt -w `find . -type f -name '*.go' -not -path "./vendor/*"`

.PHONY: image/build
image/build: code/compile
	@packr
	@operator-sdk build ${REG}/${ORG}/${IMAGE}:${TAG}
	@packr clean

.PHONY: image/push
image/push:
	docker push ${REG}/${ORG}/${IMAGE}:${TAG}

.PHONY: image/build/push
image/build/push: image/build
	@docker push ${REG}/${ORG}/${IMAGE}:${TAG}

.PHONY: test/unit
test/unit:
	@go test -v -race -cover ./pkg/...


.PHONY: test/e2e/prepare
test/e2e/prepare:
	oc create secret generic apicurio-operator-test-env --from-literal="apicurio-apps-host=${APPS_HOST}" --from-literal="apicurio-kc-host=${KC_HOST}" -n ${NS}

.PHONY: test/e2e
test/e2e: image/build/test image/push
	operator-sdk test local ${TEST_FOLDER} --go-test-flags "-v"

.PHONY: test/e2e/clear
test/e2e/clear:
	oc delete secret/apicurio-operator-test-env -n ${NS}

.PHONY: test/e2e/cluster
test/e2e/cluster: image/build/test image/push
	oc apply -f deploy/test-pod.yaml -n ${NS}
	${SHELL} ./scripts/stream-pod ${TEST_POD_NAME} ${NS}

.PHONY: image/build/test
image/build/test:
	@packr
	operator-sdk build --enable-tests ${REG}/${ORG}/${IMAGE}:${TAG}
	@packr clean

.PHONY: cluster/prepare
cluster/prepare:
	oc create namespace apicurio-operator-test
	oc apply -f ${DEPLOY_DIR}/role.yaml -n ${NS}
	oc apply -f ${DEPLOY_DIR}/role_binding.yaml -n ${NS}
	oc apply -f ${DEPLOY_DIR}/service_account.yaml -n ${NS}
	oc apply -f ${DEPLOY_DIR}/crds/integreatly_v1alpha1_apicuriodeployment_crd.yaml -n ${NS}

.PHONY: cluster/deploy
cluster/deploy:
	oc apply -f ${DEPLOY_DIR}/crds/integreatly_v1alpha1_apicuriodeployment_cr.yaml -n ${NS}
	oc apply -f ${DEPLOY_DIR}/operator.yaml -n ${NS}

.PHONY: cluster/clean
cluster/clean:
	oc delete namespace apicurio-operator-test
	oc delete all -l 'template=apicurio-studio'