package apicuriodeployment

import (
	"context"
	"log"

	integreatlyv1alpha1 "github.com/integr8ly/apicurio-operator/pkg/apis/integreatly/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"fmt"
	openshift "github.com/integr8ly/apicurio-operator/pkg/apis/integreatly/openshift/client"
	v1template "github.com/openshift/api/template/v1"
	kuberr "k8s.io/apimachinery/pkg/api/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strings"
	"github.com/openshift/api/image/v1"
	v12 "github.com/openshift/api/apps/v1"

	"github.com/gobuffalo/packr"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// Add creates a new ApicurioDeployment Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileApiCurioDeployment{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		config: mgr.GetConfig(),
		box: packr.NewBox("../../../res"),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	opts := controller.Options{Reconciler: r}
	c, err := controller.New("apicuriodeployment-controller", mgr, opts)
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ApicurioDeployment
	err = c.Watch(&source.Kind{Type: &integreatlyv1alpha1.ApicurioDeployment{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &integreatlyv1alpha1.ApicurioDeployment{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileApiCurioDeployment{}



func (r *ReconcileApiCurioDeployment) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Printf("Reconciling ApicurioDeployment %s/%s\n", request.Namespace, request.Name)

	// Fetch the ApicurioDeployment instance
	instance := &integreatlyv1alpha1.ApicurioDeployment{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}

		return reconcile.Result{}, err
	}

	if instance.GetDeletionTimestamp() != nil {
		err = r.deprovision(instance)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("deprovisioning failed: %v", err)
		}
		return reconcile.Result{}, nil
	}

	ok, err := integreatlyv1alpha1.HasFinalizer(instance, integreatlyv1alpha1.ApicurioFinalizer)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("could not read CR finalizer: %v", err)
	}

	if !ok {
		err = integreatlyv1alpha1.AddFinalizer(instance, integreatlyv1alpha1.ApicurioFinalizer)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("failed to set finalizer in object: %v", err)
		}
		err = r.client.Update(context.TODO(), instance)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("failed update in object: %v", err)
		}
	}

	err = r.bootstrap(request)
	if err != nil {
		return reconcile.Result{}, err
	}

	exts, err := r.processTemplate(instance)
	if err != nil {
		logrus.Errorf("Error while processing template: %v", err)
		return reconcile.Result{}, err
	}

	objs, err := r.getRuntimeObjects(exts)
	if err != nil {
		logrus.Errorf("Error while retrieving runtime objects: %v", err)
		return reconcile.Result{}, err
	}

	err = r.createObjects(objs, request.Namespace, instance)
	if err != nil {
		logrus.Errorf("Error creating runtime objects: %v", err)
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileApiCurioDeployment) bootstrap(request reconcile.Request) error {
	var err error

	if r.tmpl == nil {
		restCfg := r.config
		namespace := request.Namespace
		r.tmpl, err = openshift.NewTemplate(namespace, restCfg, openshift.TemplateDefaultOpts)

		return err
	}

	return nil
}

func (r *ReconcileApiCurioDeployment) processTemplate(cr *integreatlyv1alpha1.ApicurioDeployment) ([]runtime.RawExtension, error) {
	tmplPath := cr.Spec.Template
	if tmplPath == "" {
		return nil, fmt.Errorf("Spec.Template.Path property is not defined")
	}

	yamlData, err := r.box.Find(cr.Spec.Template)
	if err  != nil {
		return nil, err
	}

	jsonData, err := yaml.ToJSON(yamlData)
	if err  != nil {
		return nil, err
	}

	res, err := openshift.LoadKubernetesResource(jsonData)
	if err != nil {
		return nil, err
	}

	params := make(map[string]string)
	for k, v := range routeParams {
		if k == "AUTH_ROUTE" && cr.Spec.ExternalAuthUrl != "" {
			params[k] = cr.Spec.ExternalAuthUrl
			continue
		}
		params[k] = v + "." + cr.Spec.AppDomain
	}

	if cr.Spec.AuthRealm != "" {
		params["KC_REALM"] = cr.Spec.AuthRealm
	}

	if len(cr.Spec.JvmHeap) == 2 {
		params["API_JVM_MIN"] = cr.Spec.JvmHeap[0]
		params["API_JVM_MAX"] = cr.Spec.JvmHeap[1]
	}
	if len(cr.Spec.MemLimit) == 2 {
		params["API_MEM_REQUEST"] = cr.Spec.MemLimit[0]
		params["API_MEM_LIMIT"] = cr.Spec.MemLimit[1]
	}

	tmpl := res.(*v1template.Template)
	r.tmpl.FillParams(tmpl, params)
	ext, err := r.tmpl.Process(tmpl, params, openshift.TemplateDefaultOpts)

	return ext, err
}

func (r *ReconcileApiCurioDeployment) getRuntimeObjects(exts []runtime.RawExtension) ([]runtime.Object, error) {
	objects := make([]runtime.Object, 0)

	for _, ext := range exts {
		res, err := openshift.LoadKubernetesResource(ext.Raw)
		if err != nil {
			return nil, err
		}
		objects = append(objects, res)
	}

	return objects, nil
}

func (r *ReconcileApiCurioDeployment) transformObject(o runtime.Object) (*unstructured.Unstructured, error) {
	uo, err := openshift.UnstructuredFromRuntimeObject(o)
	if err != nil {
		return nil, err
	}

	return uo, nil
}

func (r *ReconcileApiCurioDeployment) createObject(o runtime.Object, cr *integreatlyv1alpha1.ApicurioDeployment) error {
	err := r.client.Create(context.TODO(), o)
	if err != nil && !kuberr.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func (r *ReconcileApiCurioDeployment) createObjects(objects []runtime.Object, ns string, cr *integreatlyv1alpha1.ApicurioDeployment) error {
	for _, o := range objects {
		uo, err := r.transformObject(o)
		if err != nil {
			return fmt.Errorf("failed to transform object: %v", err)
		}

		uo.SetNamespace(ns)
		err = controllerutil.SetControllerReference(cr, uo, r.scheme)
		if err != nil {
			return fmt.Errorf("failed to set owner in object: %v", err)
		}

		//ignore auth objects if using external keycloak
		isAuthObj := strings.Contains(uo.GetName(), "auth")
		if cr.Spec.ExternalAuthUrl != "" && isAuthObj {
			continue
		}

		//fix image tag based on cr version property
		if uo.GetKind() == "ImageStream" {
			j, _ := uo.MarshalJSON()
			obj, _ := openshift.LoadKubernetesResource(j)
			is := obj.(*v1.ImageStream)
			for i, tag := range is.Spec.Tags {
				tagRef := &is.Spec.Tags[i]
				tagRef.Name = cr.Spec.Version
				if tag.From.Kind == "DockerImage" {
					tagRef.From.Name = strings.Replace(tagRef.From.Name, "latest-release", cr.Spec.Version, -1)
				}
			}
		}

		if uo.GetKind() == "DeploymentConfig" {
			j, _ := uo.MarshalJSON()
			obj, _ := openshift.LoadKubernetesResource(j)
			dc := obj.(*v12.DeploymentConfig)
			container := &dc.Spec.Template.Spec.Containers[0]
			container.Image = strings.Replace(container.Image, "latest-release", cr.Spec.Version, -1)
		}

		err = r.createObject(uo.DeepCopyObject(), cr)
		if err != nil {
			if kuberr.IsAlreadyExists(err) {
				continue
			}
			return fmt.Errorf("failed to create object: %v", err)
		}
	}

	return nil
}

func (r *ReconcileApiCurioDeployment) deprovision(cr *integreatlyv1alpha1.ApicurioDeployment) error {
	ok, err := integreatlyv1alpha1.HasFinalizer(cr, "foregroundDeletion")
	if err != nil {
		return err
	}

	if ok {
		return nil
	}

	_, err = integreatlyv1alpha1.RemoveFinalizer(cr, integreatlyv1alpha1.ApicurioFinalizer)
	if err != nil {
		return err
	}

	err = r.client.Update(context.TODO(), cr)
	if err != nil {
		return fmt.Errorf("failed to update object: %v", err)
	}

	return nil
}
