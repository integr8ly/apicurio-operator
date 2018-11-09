SHELL = /bin/bash
REG = quay.io
ORG = integreatly
IMAGE = apicurio-operator
TAG = latest
KUBE_CMD = oc apply -f
RESOURCES_DIR = ./res
DEPLOY_DIR = deploy
OUT_STATIC_DIR = build/_output
OUTPUT_BIN_NAME = ${IMAGE}
TARGET_BIN = cmd/manager/main.go
NS = apicurio-operator-test
SA = apicurio-operator

travis/setup:
	@echo Installing golang dependencies
	@go get golang.org/x/sys/unix
	@go get golang.org/x/crypto/ssh/terminal
	@echo Installing dep
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
	@echo Installing errcheck
	@go get github.com/kisielk/errcheck
	@echo setup complete run make build deploy to build and deploy the operator to a local cluster
	dep ensure

code/k8s:
	operator-sdk generate k8s

code/fix:
	gofmt -w `find . -type f -name '*.go' -not -path "./vendor/*"`

code/check:
	diff -u <(echo -n) <(gofmt -d `find . -type f -name '*.go' -not -path "./vendor/*"`)

code/compile:
	go build -o ${OUTPUT_BIN_NAME} ${TARGET_BIN}

test/unit:
	go test -v -race -cover ./pkg/...

test/e2e/cluster:
	operator-sdk test cluster --namespace ${NS} --service-account ${SA} ${REG}/${ORG}/${IMAGE}:${TAG}

res/copy:
	mkdir -p ${OUT_STATIC_DIR}/res
	cp -R ${RESOURCES_DIR}/* ${OUT_STATIC_DIR}/res

image/build: res/copy
	operator-sdk build ${REG}/${ORG}/${IMAGE}:${TAG}

image/build-with-tests: res/copy
	operator-sdk build --enable-tests ${REG}/${ORG}/${IMAGE}:${TAG}

image/push:
	docker push ${REG}/${ORG}/${IMAGE}:${TAG}

cluster/prepare:
	${KUBE_CMD} ${DEPLOY_DIR}/role.yaml
	${KUBE_CMD} ${DEPLOY_DIR}/role_binding.yaml
	${KUBE_CMD} ${DEPLOY_DIR}/service_account.yaml
	${KUBE_CMD} ${DEPLOY_DIR}/crds/integreatly_v1alpha1_apicuriodeployment_crd.yaml
	${KUBE_CMD} ${DEPLOY_DIR}/crds/integreatly_v1alpha1_apicuriodeployment_cr.yaml

cluster/deploy:
	${KUBE_CMD} ${DEPLOY_DIR}/operator.yaml

cluster/clean:
	oc delete all -l 'template=apicurio-studio'
