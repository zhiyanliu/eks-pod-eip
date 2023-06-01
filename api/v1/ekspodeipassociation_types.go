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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// EKSPodEIPAssociationSpec defines the desired state of EKSPodEIPAssociation
type EKSPodEIPAssociationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	EIPAllocationID string `json:"eipAllocationID"`
	PodNamespace    string `json:"podNamespace"`
	PodName         string `json:"podName"`
	PrivateIP       string `json:"privateIP"`
}

// EKSPodEIPAssociationStatus defines the observed state of EKSPodEIPAssociation
type EKSPodEIPAssociationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Associated bool `json:"associated"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// EKSPodEIPAssociation is the Schema for the ekspodeipassociations API
type EKSPodEIPAssociation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EKSPodEIPAssociationSpec   `json:"spec,omitempty"`
	Status EKSPodEIPAssociationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// EKSPodEIPAssociationList contains a list of EKSPodEIPAssociation
type EKSPodEIPAssociationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EKSPodEIPAssociation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EKSPodEIPAssociation{}, &EKSPodEIPAssociationList{})
}
