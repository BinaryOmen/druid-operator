package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DruidSpec represents the druid spec.
// Scope: Cluster Level
type DruidSpec struct {
	// Nodes represents various druid nodes
	Nodes map[string]NodeSpec `json:"nodes"`
	// Required: JVM Options
	JvmOptions string `json:"jvm.options,omitempty"`
	// Required: StartScript
	StartScript string `json:"startscript"`
	// Required: Image for cluster
	Image string `json:"image,omitempty"`
	// Optional: Tolerations
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
	// Required: Log4jConfig
	Log4jConfig string `json:"log4j.config,omitempty"`
	// Optional: Affinity
	Affinity *v1.Affinity `json:"affinity,omitempty"`
	// Optional: ImagePullSecrets
	ImagePullSecrets []v1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// Required: CommonRuntimeProperties
	CommonRuntimeProperties string `json:"common.runtime.properties"`
	// Optional: SecurityContext
	SecurityContext *v1.PodSecurityContext `json:"securityContext,omitempty"`
	// Optional: Env's
	Env []v1.EnvVar `json:"env,omitempty"`
	// Required: Path to mount commonruntimeproperties
	CommonConfigMountPath string `json:"commonConfigMountPath"`
}

// NodeSpec specific to all nodes
type NodeSpec struct {
	// Required: Name of process
	Name string `json:"name"`
	// NodeType: Can be historical, middlemanager, coordinator, router, overlord
	NodeType string `json:"nodeType"`
	// Required: Replicas
	Replicas int32 `json:"replicas"`
	// Required: MountPath to mount all the runtime.properties, logs and jvm config inside the node as configMap
	MountPath string `json:"mountPath,omitempty"`
	// Required: Runtime Properties for all nodes
	RuntimeProperties string `json:"runtime.properties,omitempty"`
	// Required: Druid Service
	Service DruidService `json:"service"`
	// Optional: Ingress
	Ingress DruidIngress `json:"ingress,omitempty"`
	// Optional: JVM Options
	JvmOptions string `json:"jvm.options,omitempty"`
	// Optional: Log4jConfig
	Log4jConfig string `json:"log4j.config,omitempty"`
	// Optional: Volumes
	Volumes []v1.Volume `json:"volumes,omitempty"`
	// Optional: VolumeMounts
	VolumeMounts []v1.VolumeMount `json:"volumeMounts,omitempty"`
	// Optional: Annotations
	Annotations map[string]string `json:"annotations,omitempty"`
	// Optional: NodeSelector
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Optional: Tolerations
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`
	// Optional: Affinity
	Affinity *v1.Affinity `json:"affinity,omitempty"`
	// Optional: ResourcesS
	Resources v1.ResourceRequirements `json:"resources,omitempty"`
	// Optional: Env's
	Env []v1.EnvVar `json:"env,omitempty"`
	// Optional: SecurityContext
	SecurityContext *v1.PodSecurityContext `json:"securityContext,omitempty"`
	// Optional: VolumeClaimTemplates
	VolumeClaimTemplates []v1.PersistentVolumeClaim `json:"volumeClaimTemplates,omitempty"`
	// Optional: Pod Disruption Budget
	PodDisruptionBudget bool `json:"podDisruptionBudget,omitempty"`
}

type DruidService struct {
	Port       int32          `json:"port"`
	TargetPort int32          `json:"targetPort"`
	Type       v1.ServiceType `json:"type,omitempty"`
}

type DruidIngress struct {
	Annotations   map[string]string `json:"annotations,omitempty"`
	Hostname      string            `json:"hostname,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	Path          string            `json:"path,omitempty"`
	Enabled       bool              `json:"enabled,omitempty"`
	TLSEnabled    bool              `json:"tlsEnabled,omitempty"`
	TLSSecretName string            `json:"tlsSecretName,omitempty"`
	TargetPort    string            `json:"targetPort,omitempty"`
}

// DruidStatus defines the observed state of Druid
type DruidStatus struct {
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
