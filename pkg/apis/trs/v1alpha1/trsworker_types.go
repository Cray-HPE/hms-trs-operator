/*
 * MIT License
 *
 * (C) Copyright [2021] Hewlett Packard Enterprise Development LP
 *
 * Permission is hereby granted, free of charge, to any person obtaining a
 * copy of this software and associated documentation files (the "Software"),
 * to deal in the Software without restriction, including without limitation
 * the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the
 * Software is furnished to do so, subject to the following conditions:
 * The above copyright notice and this permission notice shall be included
 * in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 * THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 * OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 * ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 * OTHER DEALINGS IN THE SOFTWARE.
 */

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

	WorkerImageTag string `json:"worker_image_tag,omitempty"`
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
// +kubebuilder:printcolumn:name="Worker Image Tag",type="string",JSONPath=".spec.worker_image_tag",description="The tag that should be used when pulling the image"
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
