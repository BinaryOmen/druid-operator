package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DruidSpec defines the druid spec
type DruidSpec struct {
	Historicals Historicals `json:"historicals,omitempty"`
	StartScript string      `json:"startscript"`
	Image       string      `json:"image,omitempty"`
}

// Historicals params specific to historical nodes
type Historicals struct {
	Enabled    bool  `json:"enabled"`
	Replicas   int32 `json:"replicas"`
	Port       int32 `json:"port"`
	CommonNode `json:",inline,omitempty"`
}

// CommonNode Properties for all the processes
type CommonNode struct {
	MountPath            string                     `json:"mountPath,omitempty"`
	Volumes              []v1.Volume                `json:"volumes,omitempty"`
	RuntimeProperties    string                     `json:"runtime.properties,omitempty"`
	VolumeClaimTemplates []v1.PersistentVolumeClaim `json:"volumeClaimTemplates,omitempty"`
}

// DruidStatus defines the observed state of Druid
type DruidStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Druid is the Schema for the druids API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=druids,scope=Namespaced
type Druid struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DruidSpec   `json:"spec,omitempty"`
	Status DruidStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DruidList contains a list of Druid
type DruidList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Druid `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Druid{}, &DruidList{})
}
