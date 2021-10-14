package operations

import (
	"context"

	"github.com/rancher/opni/pkg/opnictl/common"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	statefulset appsv1.StatefulSet

	statefulsetLabels = map[string]string{"app": "sim-statefulset"}
)

func buildStatefulset() appsv1.StatefulSet {
	return appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: "sim-statefulset",
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: statefulsetLabels,
			},
			Replicas: pointer.Int32(1),
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: statefulsetLabels,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            "httpbin",
							Image:           "kennethreitz/httpbin",
							ImagePullPolicy: v1.PullAlways,
							Ports: []v1.ContainerPort{
								{
									Name:          portName,
									Protocol:      v1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "sim-pvc",
									MountPath: "/opt/sim-vol",
								},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: []v1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "sim-pvc",
					},
					Spec: v1.PersistentVolumeClaimSpec{
						AccessModes: []v1.PersistentVolumeAccessMode{
							v1.ReadWriteOnce,
						},
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								"storage": *resource.NewQuantity(1024*1024*1024, resource.BinarySI),
							},
						},
					},
				},
			},
		},
	}
}

func CreateStatefulset(ctx context.Context, namespace string) {
	statefulset = buildStatefulset()
	common.Log.Infof("Creating Statefulset %s", statefulset.Name)
	statefulset.SetNamespace(namespace)
	if err := common.K8sClient.Create(ctx, &statefulset); err != nil {
		common.Log.Error(err)
	}
}

func DeleteStatefulset(ctx context.Context, namespace string) {
	common.Log.Infof("Deleting Statefulset %s", &statefulset.Name)
	if err := common.K8sClient.Delete(ctx, &statefulset); err != nil {
		common.Log.Error(err)
	}

	common.Log.Info("Deleting all persistent volumes")
	if err := common.K8sClient.DeleteAllOf(ctx, &v1.PersistentVolume{}, client.InNamespace(namespace)); err != nil {
		common.Log.Error(err)
	}
}
