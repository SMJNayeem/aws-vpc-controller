/*
Copyright 2024.

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
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VPCSpec defines the desired state of VPC
// type VPCSpec struct {
// 	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
// 	// Important: Run "make" to regenerate code after modifying this file

// 	// Foo is an example field of VPC. Edit vpc_types.go to remove/update
// 	Foo string `json:"foo,omitempty"`
// }

// // VPCStatus defines the observed state of VPC
// type VPCStatus struct {
// 	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
// 	// Important: Run "make" to regenerate code after modifying this file
// }

// // +kubebuilder:object:root=true
// // +kubebuilder:subresource:status

// // VPC is the Schema for the vpcs API
// type VPC struct {
// 	metav1.TypeMeta   `json:",inline"`
// 	metav1.ObjectMeta `json:"metadata,omitempty"`

// 	Spec   VPCSpec   `json:"spec,omitempty"`
// 	Status VPCStatus `json:"status,omitempty"`
// }

// // +kubebuilder:object:root=true

// VPCList contains a list of VPC
// type VPCList struct {
// 	metav1.TypeMeta `json:",inline"`
// 	metav1.ListMeta `json:"metadata,omitempty"`
// 	Items           []VPC `json:"items"`
// }

// func init() {
// 	SchemeBuilder.Register(&VPC{}, &VPCList{})
// }

// VPCSpec defines the desired state of VPC
type VPCSpec struct {
	// CIDR block for the VPC
	CIDR string `json:"cidr"`

	// Name of the VPC
	Name string `json:"name"`

	// Region where the VPC should be created
	Region string `json:"region"`
}

// VPCStatus defines the observed state of VPC
type VPCStatus struct {
	// ID of the created VPC
	VPCID string `json:"vpcId,omitempty"`

	// Current state of the VPC
	State string `json:"state,omitempty"`

	// Any error message
	Error string `json:"error,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// VPC is the Schema for the vpcs API
type VPC struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VPCSpec   `json:"spec,omitempty"`
	Status VPCStatus `json:"status,omitempty"`
}

type VPCList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VPC `json:"items"`
}

// Ensure VPCList implements runtime.Object
func (v *VPCList) DeepCopyObject() runtime.Object {
	return v.DeepCopy() // Assuming DeepCopy is generated
}

func init() {
	SchemeBuilder.Register(&VPC{}, &VPCList{})
}
