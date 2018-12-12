package e2e

import (
	"github.com/integr8ly/apicurio-operator/pkg/apis/integreatly/v1alpha1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"testing"
)

func TestApicurioDeployment(t *testing.T) {
	var err error
	t.Parallel()

	ctx := prepare(t)
	defer ctx.Cleanup()

	err = register()
	if err != nil {
		t.Fatalf("Failed to register crd scheme: %v", err)
	}

	cr := buildCr(ctx)
	err = doDeployment(framework.Global, ctx, cr)
	if err != nil {
		t.Fatalf("Failed to create cr: %v", err)
	}

	err = validateDeployment(framework.Global, ctx, cr)
	if err != nil {
		t.Fatalf("Failed to validate deployment: %v", err)
	}
}

func buildCr(ctx *framework.TestCtx) *v1alpha1.ApicurioDeployment {
	ns, _ := ctx.GetNamespace()

	return &v1alpha1.ApicurioDeployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ApicurioDeployment",
			APIVersion: "integreatly.org/apicuriodeployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "apicurio-standalone",
			Namespace: ns,
		},
		Spec: v1alpha1.ApicurioDeploymentSpec{
			Template:  "apicurio-template.yml",
			Version:   "0.2.18.Final",
			AppDomain: os.Getenv("APICURIO_APPS_HOST"),
			JvmHeap:   [2]string{"768m", "2048m"},
			MemLimit:  [2]string{"800Mi", "4Gi"},
		},
	}
}
