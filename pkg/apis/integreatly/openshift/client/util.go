package client

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"fmt"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/yaml"
	cgoscheme "k8s.io/client-go/kubernetes/scheme"
	"strings"

	apps "github.com/openshift/api/apps/v1"
	authorization "github.com/openshift/api/authorization/v1"
	build "github.com/openshift/api/build/v1"
	image "github.com/openshift/api/image/v1"
	route "github.com/openshift/api/route/v1"
	template "github.com/openshift/api/template/v1"

	"encoding/json"
)

var (
	scheme      = runtime.NewScheme()
	codecs      = serializer.NewCodecFactory(scheme)
	decoderFunc = decoder
)

func init() {
	metav1.AddToGroupVersion(scheme, schema.GroupVersion{Version: "v1"})
	cgoscheme.AddToScheme(scheme)

	//add openshift types
	apps.AddToScheme(scheme)
	authorization.AddToScheme(scheme)
	build.AddToScheme(scheme)
	image.AddToScheme(scheme)
	route.AddToScheme(scheme)
	template.AddToScheme(scheme)

	//legacy openshift types
	apps.AddToSchemeInCoreGroup(scheme)
	authorization.AddToSchemeInCoreGroup(scheme)
	build.AddToSchemeInCoreGroup(scheme)
	image.AddToSchemeInCoreGroup(scheme)
	route.AddToSchemeInCoreGroup(scheme)
	template.AddToSchemeInCoreGroup(scheme)
}

func decoder(gv schema.GroupVersion, codecs serializer.CodecFactory) runtime.Decoder {
	codec := codecs.UniversalDecoder(gv)
	return codec
}

func runtimeObjectFromUnstructured(u *unstructured.Unstructured) (runtime.Object, error) {
	gvk := u.GroupVersionKind()
	decoder := decoderFunc(gvk.GroupVersion(), codecs)

	b, err := u.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("error running MarshalJSON on unstructured object: %v", err)
	}

	ro, _, err := decoder.Decode(b, &gvk, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decode json data with gvk(%v): %v", gvk.String(), err)
	}

	return ro, nil
}

func LoadKubernetesResourceFromFile(path string) (runtime.Object, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	data, err = JsonIfYaml(data, path)
	if err != nil {
		return nil, err
	}

	return LoadKubernetesResource(data)
}

func LoadKubernetesResource(jsonData []byte) (runtime.Object, error) {
	u := unstructured.Unstructured{}

	err := u.UnmarshalJSON(jsonData)
	if err != nil {
		return nil, err
	}

	return runtimeObjectFromUnstructured(&u)
}

func JsonIfYaml(source []byte, filename string) ([]byte, error) {
	if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
		return yaml.ToJSON(source)
	}

	return source, nil
}

func UnstructuredFromRuntimeObject(ro runtime.Object) (*unstructured.Unstructured, error) {
	b, err := json.Marshal(ro)
	if err != nil {
		return nil, fmt.Errorf("error running MarshalJSON on runtime object: %v", err)
	}
	var u unstructured.Unstructured
	if err := json.Unmarshal(b, &u.Object); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json into unstructured object: %v", err)
	}
	return &u, nil
}
