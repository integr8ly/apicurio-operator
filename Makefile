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

travis/setup:
	@echo Installing golang dependencies
	@go get golang.org/x/sys/unix
	@go get golang.org/x/crypto/ssh/terminal
	@go get -u github.com/gobuffalo/packr/packr
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

code/compile-for-docker:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ${OUT_STATIC_DIR}/bin/${IMAGE} ${TARGET_BIN}

test/unit:
	go test -v -race -cover ./pkg/...

test/e2e/local: image/build-with-tests image/push
	operator-sdk test local ${TEST_FOLDER} --go-test-flags "-v"

test/e2e/cluster: image/build-with-tests image/push
	oc apply -f deploy/test-pod.yaml -n ${NS}
	${SHELL} ./scripts/stream-pod ${TEST_POD_NAME} ${NS}

test/e2e/prepare:
	oc create secret generic apicurio-operator-test-env --from-literal="apicurio-apps-host=${APPS_HOST}" --from-literal="apicurio-kc-host=${KC_HOST}" -n ${NS}

test/e2e/clear:
	oc delete secret/apicurio-operator-test-env -n ${NS}

res/copy:
	packr

image/build: res/copy
	operator-sdk build ${REG}/${ORG}/${IMAGE}:${TAG}

image/docker-build: res/copy code/compile-for-docker
	docker build -t ${REG}/${ORG}/${IMAGE}:${TAG} -f build/Dockerfile .

image/build-with-tests: res/copy
	operator-sdk build --enable-tests ${REG}/${ORG}/${IMAGE}:${TAG}

image/push:
	docker push ${REG}/${ORG}/${IMAGE}:${TAG}

image/docker-build-and-push: image/docker-build image/push


cluster/prepare:
	oc apply -f ${DEPLOY_DIR}/role.yaml -n ${NS}
	oc apply -f ${DEPLOY_DIR}/role_binding.yaml -n ${NS}
	oc apply -f ${DEPLOY_DIR}/service_account.yaml -n ${NS}
	oc apply -f ${DEPLOY_DIR}/crds/integreatly_v1alpha1_apicuriodeployment_crd.yaml -n ${NS}

cluster/deploy:
	oc apply -f ${DEPLOY_DIR}/crds/integreatly_v1alpha1_apicuriodeployment_cr.yaml -n ${NS}
	oc apply -f ${DEPLOY_DIR}/operator.yaml -n ${NS}

cluster/clean:
	oc delete all -l 'template=apicurio-studio'
