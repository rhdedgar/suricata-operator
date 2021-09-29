package k8s

import (
	managedv1alpha1 "github.com/rhdedgar/suricata-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SuricataDaemonSet returns a new daemonset customized for suricata
func SuricataDaemonSet(m *managedv1alpha1.Suricata) *appsv1.DaemonSet {
	var privileged = true
	var runAsUser int64

	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": "suricata",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"name": "suricata",
					},
				},
				Spec: corev1.PodSpec{
					NodeSelector: map[string]string{
						"beta.kubernetes.io/os": "linux",
					},
					//ServiceAccountName: "openshift-scanning-operator",
					Tolerations: []corev1.Toleration{
						{
							Operator: corev1.TolerationOpExists,
						},
					},
					HostNetwork: true,
					Containers: []corev1.Container{{
						Image:           "quay.io/dedgar/suricata:v0.0.7",
						Name:            "suricata",
						ImagePullPolicy: corev1.PullAlways,
						SecurityContext: &corev1.SecurityContext{
							Privileged: &privileged,
							RunAsUser:  &runAsUser,
						},
						Env: []corev1.EnvVar{{
							Name:  "OO_PAUSE_ON_START",
							Value: "false",
						}, {
							Name:  "SURICATA_LOG_FILE",
							Value: "/host/var/log/openshift_managed_suricata.log",
						}},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("50Mi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("900m"),
								corev1.ResourceMemory: resource.MustParse("900Mi"),
							},
						},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "suricata-secrets",
							MountPath: "/secrets",
						}, {
							Name:      "host-filesystem",
							MountPath: "/host/",
						}},
					}},
					Volumes: []corev1.Volume{{
						Name: "suricata-secrets",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: "suricata-secrets",
							},
						},
					}, {
						Name: "host-filesystem",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/",
							},
						},
					}},
				},
			},
		},
	}
	return ds
}
