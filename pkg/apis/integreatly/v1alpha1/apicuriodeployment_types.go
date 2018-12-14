package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ApicurioDeploymentSpec defines the desired state of ApicurioDeployment
const (
	ApicurioFinalizer = "finalizer.org.integreatly.apicurio"
)

type ApicurioDeploymentSpec struct {
	Version         string    `json:"version"`
	AppDomain       string    `json:"app_domain"`
	Template        string    `json:"template"`
	ExternalAuthUrl string    `json:"external_auth_url"`
	AuthRealm       string    `json:"auth_realm"`
	JvmHeap         [2]string `json:"jvm_heap"`
	MemLimit        [2]string `json:"mem_limit"`
}

// ApicurioDeploymentStatus defines the observed state of ApicurioDeployment
type ApicurioDeploymentStatus struct {
	Message string `json:"message"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ApicurioDeployment is the Schema for the apicuriodeployments API
// +k8s:openapi-gen=true
type ApicurioDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApicurioDeploymentSpec   `json:"spec,omitempty"`
	Status ApicurioDeploymentStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ApicurioDeploymentList contains a list of ApicurioDeployment
type ApicurioDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApicurioDeployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ApicurioDeployment{}, &ApicurioDeploymentList{})
}
