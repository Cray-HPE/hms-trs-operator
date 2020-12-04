package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TRSWorkerSpec defines the desired state of TRSWorker
type TRSWorkerSpec struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// +kubebuilder:validation:Enum=http
	WorkerType string `json:"worker_type"`

	// +kubebuilder:validation:Enum=v1
	WorkerVersion string `json:"worker_version"`
}

// TRSWorkerStatus defines the observed state of TRSWorker
type TRSWorkerStatus struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TRSWorker is the Schema for the trsworkers API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=trsworkers,scope=Namespaced
// +kubebuilder:printcolumn:name="Worker Type",type="string",JSONPath=".spec.worker_type",description="The type of worker this deployment is using"
// +kubebuilder:printcolumn:name="Worker Version",type="string",JSONPath=".spec.worker_version",description="The version of worker this deployment is using"
type TRSWorker struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TRSWorkerSpec   `json:"spec,omitempty"`
	Status TRSWorkerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TRSWorkerList contains a list of TRSWorker
type TRSWorkerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TRSWorker `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TRSWorker{}, &TRSWorkerList{})
}
