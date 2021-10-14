package operations

import (
	"context"
	"math/rand"
	"time"

	"github.com/rancher/opni/pkg/opnictl/common"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	daemonSet       appsv1.DaemonSet
	daemonSetLabels = map[string]string{"app": "sim-daemonset"}
)

func buildDaemonSet() appsv1.DaemonSet {
	return appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: "sim-daemonset",
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: daemonSetLabels,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: daemonSetLabels,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            "logger",
							Image:           "chentex/random-logger:latest",
							ImagePullPolicy: v1.PullAlways,
							Args: []string{
								"500",
								"1000",
							},
						},
					},
				},
			},
		},
	}
}

func CreateDaemonSet(ctx context.Context, namespace string) {
	daemonSet = buildDaemonSet()

	common.Log.Infof("Creating DaemonSet %s", daemonSet.Name)
	daemonSet.SetNamespace(namespace)
	if err := common.K8sClient.Create(ctx, &daemonSet); err != nil {
		common.Log.Error(err)
	}
}

func DeleteDaemonSet(ctx context.Context) {
	common.Log.Infof("Deleting DaemonSet %s", daemonSet.Name)
	if err := common.K8sClient.Delete(ctx, &daemonSet); err != nil {
		common.Log.Error(err)
	}
}

func DeleteRandomDaemonSetPod(ctx context.Context, namespace string) {
	// Get list of pods matching deploymetn selector
	podList := &v1.PodList{}
	selector, err := metav1.LabelSelectorAsSelector(daemonSet.Spec.Selector)
	if err != nil {
		common.Log.Error(err)
	}
	if err := common.K8sClient.List(ctx, podList, &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: selector,
	}); err != nil {
		common.Log.Error(err)
	}

	// Select pod at random and delete
	randomSource := rand.NewSource(time.Now().UnixNano())
	randomNumber := rand.New(randomSource)
	podIndex := randomNumber.Intn(len(podList.Items))
	common.Log.Infof("Deleting Pod %s", podList.Items[podIndex].Name)
	if err := common.K8sClient.Delete(ctx, &podList.Items[podIndex]); err != nil {
		common.Log.Error(err)
	}
}
