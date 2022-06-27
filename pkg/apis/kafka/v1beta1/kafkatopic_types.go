/*
 * MIT License
 *
 * (C) Copyright [2021-2022] Hewlett Packard Enterprise Development LP
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

package v1beta1

import (
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Unfortunately this file has to exist because the Strimzi operator is written entirely in Java.

// KafkaTopicSpec defines the desired state of KafkaTopic
type KafkaTopicSpec struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	Config map[string]IntOrString `json:"config,omitempty"`

	// +kubebuilder:validation:Minimum=1
	Partitions int `json:"partitions"`

	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=32767
	Replicas int `json:"replicas"`

	TopicName string `json:"topicName,omitempty"`
}

func (i KafkaTopicSpec) DeepCopy() KafkaTopicSpec {
	o := i
	o.Config = make(map[string]IntOrString)
	for key, value := range i.Config {
		o.Config[key] = value.DeepCopy()
	}
	return o
}

type IntOrString struct {
	// The json marshaling and unmarshaling is handled by the custom
	// MarshalJSON and UnmarshalJSON functions.
	// A json annotation is still required by the operator-sdk tool.
	Value interface{} `json:"-"`
}

func (i IntOrString) DeepCopy() IntOrString {
	o := i
	return o
}

func (si IntOrString) MarshalJSON() ([]byte, error) {
	return json.Marshal(si.Value)
}

func (si *IntOrString) UnmarshalJSON(data []byte) error {
	var value interface{}
	err := json.Unmarshal(data, &value)
	if err != nil {
		return err
	}
	si.Value = value
	return nil
}

func (t IntOrString) String() string {
	return fmt.Sprintf("%v", t.Value)
}

type ConditionsStruct struct {
	LastTransitionTime string `json:"lastTransitionTime,omitempty"`
	Message            string `json:"message,omitempty"`
	Reason             string `json:"reason,omitempty"`
	Status             string `json:"status,omitempty"`
	Type               string `json:"type,omitempty"`
}

// KafkaTopicStatus defines the observed state of KafkaTopic
type KafkaTopicStatus struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Conditions         []ConditionsStruct `json:"conditions,omitempty"`
	ObservedGeneration int                `json:"observedGeneration,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KafkaTopic is the Schema for the kafkatopics API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=kafkatopics,scope=Namespaced
type KafkaTopic struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KafkaTopicSpec   `json:"spec,omitempty"`
	Status KafkaTopicStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KafkaTopicList contains a list of KafkaTopic
type KafkaTopicList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KafkaTopic `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KafkaTopic{}, &KafkaTopicList{})
}
