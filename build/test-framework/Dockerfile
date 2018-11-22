ARG BASEIMAGE
FROM ${BASEIMAGE}
ADD build/_output/bin/apicurio-operator-test /usr/local/bin/apicurio-operator-test
ARG NAMESPACEDMAN
ADD $NAMESPACEDMAN /namespaced.yaml
ADD build/test-framework/go-test.sh /go-test.sh
