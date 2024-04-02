package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MyAppResourceSpec struct {
	ReplicaCount int32 `json:"replicaCount"`
	Resources    struct {
		MemoryLimit string `json:"memoryLimit"`
		CPURequest  string `json:"cpuRequest"`
	} `json:"resources"`
	Image struct {
		Repository string `json:"repository"`
		Tag        string `json:"tag"`
	} `json:"image"`
	UI struct {
		Color   string `json:"color"`
		Message string `json:"message"`
	} `json:"ui"`
	Redis struct {
		Enabled bool `json:"enabled"`
	} `json:"redis"`
}

// MyAppResourceStatus defines the observed state of MyAppResource
type MyAppResourceStatus struct {
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MyAppResource is the Schema for the myappresources API
type MyAppResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MyAppResourceSpec   `json:"spec,omitempty"`
	Status MyAppResourceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MyAppResourceList contains a list of MyAppResource
type MyAppResourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MyAppResource `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MyAppResource{}, &MyAppResourceList{})
}
