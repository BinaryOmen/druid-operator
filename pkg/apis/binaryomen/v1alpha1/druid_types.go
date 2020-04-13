package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DruidSpec represents the druid spec
type DruidSpec struct {
	// HistoricalHot represents the historical node
	Nodes map[string]NodeSpec `json:"nodes"`
	// MidleManager represents the middlemanager node
	MiddleManager NodeSpec `json:"middlemanagers"`
	// TODO: HistoricalCOld
	HistoricalsCold *NodeSpec `json:"historical-cold"`
	// Required: StartScript
	StartScript string `json:"startscript"`
	// Required: Image for cluster
	Image string `json:"image,omitempty"`
}

// NodeSpec specific to all nodes
type NodeSpec struct {
	// NodeType: Can be historical, middlemanager, coordinator, router, overlord
	NodeType string `json:"nodeType"`
	// Required
	Replicas int32 `json:"replicas"`
	// Required
	Port int32 `json:"port"`
	// Required
	CommonNode `json:",inline,omitempty"`
}

// CommonNode Properties for all the processes
type CommonNode struct {
	// Required: MountPath to mount all the runtime.properties, logs and jvm config inside the node as configMap
	MountPath string `json:"mountPath,omitempty"`
	// Required: Runtime Properties for all nodes
	RuntimeProperties string `json:"runtime.properties,omitempty"`
	// Optional
	Volumes []v1.Volume `json:"volumes,omitempty"`
	// Optional
	VolumeMounts []v1.VolumeMount `json:"volumeMounts,omitempty"`
	// Optional
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
