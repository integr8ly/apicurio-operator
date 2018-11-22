package apis

import (
	"github.com/integr8ly/apicurio-operator/pkg/apis/integreatly/v1alpha1"

	apps "github.com/openshift/api/apps/v1"
	authorization "github.com/openshift/api/authorization/v1"
	build "github.com/openshift/api/build/v1"
	image "github.com/openshift/api/image/v1"
	route "github.com/openshift/api/route/v1"
	template "github.com/openshift/api/template/v1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes,
		v1alpha1.SchemeBuilder.AddToScheme,
		apps.AddToScheme,
		apps.AddToSchemeInCoreGroup,
		authorization.AddToScheme,
		authorization.AddToSchemeInCoreGroup,
		build.AddToScheme,
		build.AddToSchemeInCoreGroup,
		image.AddToScheme,
		image.AddToSchemeInCoreGroup,
		route.AddToScheme,
		route.AddToSchemeInCoreGroup,
		template.AddToScheme,
		template.AddToSchemeInCoreGroup,
	)
}
