package apicuriodeployment

import (
	"k8s.io/client-go/rest"
	openshift "github.com/integr8ly/apicurio-operator/pkg/apis/integreatly/openshift/client"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"k8s.io/apimachinery/pkg/runtime"
	"github.com/gobuffalo/packr"
)

// ReconcileApiCurioDeployment reconciles a ApicurioDeployment object
type ReconcileApiCurioDeployment struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	config *rest.Config
	scheme *runtime.Scheme
	tmpl *openshift.Template
	box packr.Box
}

var routeParams = map[string]string{
	"UI_ROUTE": "apicurio-studio",
	"WS_ROUTE": "apicurio-studio-ws",
	"API_ROUTE": "apicurio-studio-api",
	"AUTH_ROUTE": "apicurio-studio-auth",
}
