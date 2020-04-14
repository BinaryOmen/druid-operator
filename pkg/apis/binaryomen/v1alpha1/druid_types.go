package v1alpha1

import (
	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DruidSpec represents the druid spec.
// Scope: Cluster Level
type DruidSpec struct {
	// HistoricalHot represents the historical node
	Nodes map[string]NodeSpec `json:"nodes"`
	// JVM Options
	JvmOptions string `json:"jvm.options,omitempty"`
	// Required: StartScript
	StartScript string `json:"startscript"`
	// Required: Image for cluster
	Image string `json:"image,omitempty"`

	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
	//
	Log4jConfig string `json:"log4j.config,omitempty"`
	//
	Affinity *v1.Affinity `json:"affinity,omitempty"`
	//
	ImagePullSecrets []v1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// Required: CommonRuntimeProperties
	CommonRuntimeProperties string `json:"common.runtime.properties"`
	//
	SecurityContext *v1.PodSecurityContext `json:"securityContext,omitempty"`
	//
	Env []v1.EnvVar `json:"env,omitempty"`
	// Required: Path to mount commonruntimeproperties
	CommonConfigMountPath string `json:"commonConfigMountPath"`
}

// NodeSpec specific to all nodes
type NodeSpec struct {
	//
	Name string `json:"name"`
	// NodeType: Can be historical, middlemanager, coordinator, router, overlord
	NodeType string `json:"nodeType"`
	// Required
	Replicas int32 `json:"replicas"`
	// Required
	Port int32 `json:"port"`
	// Required
	Labels map[string]string `json:"labels"`
	// Optional: JVM Options
	JvmOptions string `json:"jvm.options,omitempty"`
	// Log4jConfig
	Log4jConfig string `json:"log4j.config,omitempty"`
	// Required: MountPath to mount all the runtime.properties, logs and jvm config inside the node as configMap
	MountPath string `json:"mountPath,omitempty"`
	// Required: Runtime Properties for all nodes
	RuntimeProperties string `json:"runtime.properties,omitempty"`
	// Optional
	Volumes []v1.Volume `json:"volumes,omitempty"`
	// Optional
	VolumeMounts []v1.VolumeMount `json:"volumeMounts,omitempty"`
	//
	Annotations map[string]string `json:"annotations,omitempty"`
	//
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	//
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
	//
	Affinity *v1.Affinity `json:"affinity,omitempty"`
	//
	Resources v1.ResourceRequirements `json:"resources,omitempty"`
	// Optional
	Env []v1.EnvVar `json:"env,omitempty"`
	//
	SecurityContext *v1.PodSecurityContext `json:"securityContext,omitempty"`
	//
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
