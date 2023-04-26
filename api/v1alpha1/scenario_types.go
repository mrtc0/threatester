/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ScenarioSpec defines the desired state of Scenario
type ScenarioSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Templates    []Template    `json:"templates"`
	Expectations []Expectation `json:"expectations,omitempty"`
}

// ScenarioStatus defines the observed state of Scenario
type ScenarioStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Scenario is the Schema for the scenarios API
type Scenario struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScenarioSpec   `json:"spec,omitempty"`
	Status ScenarioStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ScenarioList contains a list of Scenario
type ScenarioList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Scenario `json:"items"`
}

type Template struct {
	Name      string            `json:"name,omitempty"`
	Container *corev1.Container `json:"container,omitempty"`
}

type Expectation struct {
	Timeout string              `json:"timeout,omitempty"`
	Datadog *DatadogExpectation `json:"datadog,omitempty"`
}

type DatadogExpectation struct {
	Monitor *DatadogMonitor `json:"monitor,omitempty"`
}

type DatadogMonitor struct {
	ID     string `json:"id,omitempty"`
	Status string `json:"status,omitempty"`
}

func init() {
	SchemeBuilder.Register(&Scenario{}, &ScenarioList{})
}
