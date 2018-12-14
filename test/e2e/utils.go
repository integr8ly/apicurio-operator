package e2e

import (
	goctx "context"
	"github.com/integr8ly/apicurio-operator/pkg/apis/integreatly/v1alpha1"
	appsv1 "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"testing"
	"time"
)

func prepare(t *testing.T) *framework.TestCtx {
	ctx := framework.NewTestCtx(t)
	opt := &framework.CleanupOptions{
		TestContext:   ctx,
		RetryInterval: retryInterval,
		Timeout:       timeout,
	}

	err := ctx.InitializeClusterResources(opt)
	if err != nil {
		t.Fatalf("Failed to initialize test context: %v", err)
	}

	ns, err := ctx.GetNamespace()
	if err != nil {
		t.Fatalf("Failed to get context namespace: %v", err)
	}

	globalVars := framework.Global

	err = e2eutil.WaitForDeployment(t, globalVars.KubeClient, ns, "apicurio-operator", 1, retryInterval, timeout)
	if err != nil {
		t.Fatalf("Operator deployment failed: %v", err)
	}

	return ctx
}

func register() error {
	stuffList := &v1alpha1.ApicurioDeploymentList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ApicurioDeployment",
			APIVersion: "integreatly.org/apicuriodeployment",
		},
	}

	err := framework.AddToFrameworkScheme(v1alpha1.SchemeBuilder.AddToScheme, stuffList)
	if err != nil {
		return err
	}

	return nil
}

func doDeployment(f *framework.Framework, ctx *framework.TestCtx, cr *v1alpha1.ApicurioDeployment) error {
	err := f.Client.Create(goctx.TODO(), cr, cleanupOpts(ctx))
	if err != nil {
		return err
	}

	return nil
}

func validateDeployment(f *framework.Framework, ctx *framework.TestCtx, cr *v1alpha1.ApicurioDeployment) error {
	var err error
	ns, err := ctx.GetNamespace()
	if err != nil {
		return err
	}

	err = waitForDC(f.KubeConfig, ns, "postgresql", retryInterval, timeout)
	if err != nil {
		return err
	}

	err = waitForDC(f.KubeConfig, ns, "apicurio-studio-ws", retryInterval, timeout)
	if err != nil {
		return err
	}

	err = waitForDC(f.KubeConfig, ns, "apicurio-studio-api", retryInterval, timeout)
	if err != nil {
		return err
	}

	err = waitForDC(f.KubeConfig, ns, "apicurio-studio-ui", retryInterval, timeout)
	if err != nil {
		return err
	}

	if cr.Spec.ExternalAuthUrl == "" {
		err = waitForDC(f.KubeConfig, ns, "apicurio-studio-auth", retryInterval, timeout)
		if err != nil {
			return err
		}
	}

	return nil
}

func cleanupOpts(ctx *framework.TestCtx) *framework.CleanupOptions {
	return &framework.CleanupOptions{
		TestContext:   ctx,
		Timeout:       timeout,
		RetryInterval: retryInterval,
	}
}

func waitForPod(t *testing.T, kubeclient kubernetes.Interface, namespace, name string, retryInterval, timeout time.Duration) error {
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		pod, err := kubeclient.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}

		if pod.Status.Phase == v1.PodRunning {
			return true, nil
		}

		return false, nil
	})

	return err
}

func waitForDC(config *rest.Config, namespace, name string, retryInterval, timeout time.Duration) error {
	kubeApps, _ := appsv1.NewForConfig(config)
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		dc, err := kubeApps.DeploymentConfigs(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}

		if dc.Status.Replicas == dc.Status.ReadyReplicas {
			return true, nil
		}

		return false, nil
	})

	return err
}
