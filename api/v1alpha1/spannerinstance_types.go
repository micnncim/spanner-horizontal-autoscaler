/*

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SpannerInstanceSpec defines the desired state of SpannerInstance
type SpannerInstanceSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Format:=string
	InstanceID string `json:"instanceId"`

	// +kubebuilder:validation:Minimum=1
	MinNodes *int32 `json:"minNodes"`

	// +kubebuilder:validation:Minimum=1
	MaxNodes *int32 `json:"maxNodes"`

	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	CPUUtilizationThreshold *int32 `json:"cpuUtilizationThreshold"`
}

// SpannerInstanceStatus defines the observed state of SpannerInstance
type SpannerInstanceStatus struct {
	CPUUtilization *int32 `json:"cpuUtilization"`
	AvailableNodes *int32 `json:"availableNodes"`
}

// +kubebuilder:object:root=true

// SpannerInstance is the Schema for the spannerinstances API
type SpannerInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SpannerInstanceSpec   `json:"spec,omitempty"`
	Status SpannerInstanceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SpannerInstanceList contains a list of SpannerInstance
type SpannerInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SpannerInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SpannerInstance{}, &SpannerInstanceList{})
}
