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

// EksPodEipAssociationSpec defines the desired state of EksPodEipAssociation
type EksPodEipAssociationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	EipAllocationId string `json:"eipAllocationId"`
	PodNamespace    string `json:"podNamespace"`
	PodName         string `json:"podName"`
	PrivateIP       string `json:"privateIP"`
}

// EksPodEipAssociationStatus defines the observed state of EksPodEipAssociation
type EksPodEipAssociationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Associated bool   `json:"associated"`
	ElasticIP  string `json:"elasticIP"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:subresource:status

// EksPodEipAssociation is the Schema for the EksPodEipAssociations API
type EksPodEipAssociation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EksPodEipAssociationSpec   `json:"spec,omitempty"`
	Status EksPodEipAssociationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Cluster

// EksPodEipAssociationList contains a list of EksPodEipAssociation
type EksPodEipAssociationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EksPodEipAssociation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EksPodEipAssociation{}, &EksPodEipAssociationList{})
}
