package resources

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	shulkermciov1alpha1 "github.com/iamblueslime/shulker/api/v1alpha1"
)

const proxyConfigDir = "/config"
const proxyServerDir = "/server"
const proxyHealthCommand = "/health.sh"

type ProxyResourcePodBuilder struct {
	*ProxyResourceBuilder
}

func (b *ProxyResourceBuilder) ProxyPod() *ProxyResourcePodBuilder {
	return &ProxyResourcePodBuilder{b}
}

func (b *ProxyResourcePodBuilder) Build() (client.Object, error) {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.GetPodName(),
			Namespace: b.Instance.Namespace,
			Labels:    b.getLabels(),
		},
	}, nil
}

func (b *ProxyResourcePodBuilder) Update(object client.Object) error {
	pod := object.(*corev1.Pod)

	pod.Spec = corev1.PodSpec{
		InitContainers: []corev1.Container{
			{
				Image:   "busybox:stable",
				Name:    "init-fs",
				Command: []string{"ash", fmt.Sprintf("%s/init-fs.sh", proxyConfigDir)},
				Env: []corev1.EnvVar{
					{
						Name:  "SHULKER_CONFIG_DIR",
						Value: proxyConfigDir,
					},
					{
						Name:  "SHULKER_DATA_DIR",
						Value: proxyServerDir,
					},
				},
				SecurityContext: b.getSecurityContext(),
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "proxy-server",
						MountPath: proxyServerDir,
					},
					{
						Name:      "proxy-config",
						MountPath: proxyConfigDir,
						ReadOnly:  true,
					},
				},
			},
		},
		Containers: []corev1.Container{{
			Image: "itzg/bungeecord:latest",
			Name:  "proxy",
			Ports: []corev1.ContainerPort{{
				Name:          "minecraft",
				ContainerPort: 25577,
			}},
			Env: b.getEnv(),
			// LivenessProbe: &corev1.Probe{
			// 	ProbeHandler: corev1.ProbeHandler{
			// 		Exec: &corev1.ExecAction{
			// 			Command: []string{proxyHealthCommand},
			// 		},
			// 	},
			// 	InitialDelaySeconds: 10,
			// 	PeriodSeconds:       10,
			// },
			// ReadinessProbe: &corev1.Probe{
			// 	ProbeHandler: corev1.ProbeHandler{
			// 		Exec: &corev1.ExecAction{
			// 			Command: []string{proxyHealthCommand},
			// 		},
			// 	},
			// 	InitialDelaySeconds: 10,
			// 	PeriodSeconds:       10,
			// },
			SecurityContext: b.getSecurityContext(),
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "proxy-server",
					MountPath: proxyServerDir,
				},
				{
					Name:      "proxy-tmp",
					MountPath: "/tmp",
				},
			},
		}},
		RestartPolicy: corev1.RestartPolicyOnFailure,
		Volumes: []corev1.Volume{
			{
				Name: "proxy-server",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
			{
				Name: "proxy-config",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: b.GetConfigMapName(),
						},
					},
				},
			},
			{
				Name: "proxy-tmp",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
		},
	}

	if b.Instance.Spec.PodOverrides != nil {
		if b.Instance.Spec.PodOverrides.Affinity != nil {
			pod.Spec.Affinity = b.Instance.Spec.PodOverrides.Affinity
		}

		if b.Instance.Spec.PodOverrides.ServiceAccountName != "" {
			pod.Spec.ServiceAccountName = b.Instance.Spec.PodOverrides.ServiceAccountName
		}
	}

	if err := controllerutil.SetControllerReference(b.Instance, pod, b.Scheme); err != nil {
		return fmt.Errorf("failed setting controller reference for Pod: %v", err)
	}

	return nil
}

func (b *ProxyResourcePodBuilder) CanBeUpdated() bool {
	return false
}

func getTypeFromVersionChannel(channel shulkermciov1alpha1.ProxyDeploymentVersionChannel) string {
	switch channel {
	case shulkermciov1alpha1.ProxyDeploymentVersionBungeeCord:
		return "BUNGEECORD"
	case shulkermciov1alpha1.ProxyDeploymentVersionWaterfall:
		return "WATERFALL"
	case shulkermciov1alpha1.ProxyDeploymentVersionVelocity:
		return "VELOCITY"
	}

	return ""
}

func getVersionEnvFromVersionChannel(channel shulkermciov1alpha1.ProxyDeploymentVersionChannel) string {
	switch channel {
	case shulkermciov1alpha1.ProxyDeploymentVersionBungeeCord:
		return "BUNGEE_JOB_ID"
	case shulkermciov1alpha1.ProxyDeploymentVersionWaterfall:
		return "WATERFALL_BUILD_ID"
	case shulkermciov1alpha1.ProxyDeploymentVersionVelocity:
		return "VELOCITY_BUILD_ID"
	}

	return ""
}

func (b *ProxyResourcePodBuilder) getEnv() []corev1.EnvVar {
	env := []corev1.EnvVar{
		{
			Name: "SERVER_ID",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
		{
			Name:  "TYPE",
			Value: getTypeFromVersionChannel(b.Instance.Spec.Version.Channel),
		},
		{
			Name:  getVersionEnvFromVersionChannel(b.Instance.Spec.Version.Channel),
			Value: b.Instance.Spec.Version.Name,
		},
	}

	return env
}

func (b *ProxyResourcePodBuilder) getSecurityContext() *corev1.SecurityContext {
	securityEscalation := false
	readOnlyFs := true
	runAsNonRoot := true
	userUid := int64(1000)

	return &corev1.SecurityContext{
		AllowPrivilegeEscalation: &securityEscalation,
		ReadOnlyRootFilesystem:   &readOnlyFs,
		RunAsNonRoot:             &runAsNonRoot,
		RunAsUser:                &userUid,
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{"ALL"},
		},
	}
}
