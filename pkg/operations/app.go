package operations

import (
	"context"
	"math/rand"
	"time"

	"github.com/rancher/opni/pkg/opnictl/common"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	deployment appsv1.Deployment
	service    v1.Service
	ingress    networkingv1beta1.Ingress

	portName           = "http"
	deploymentReplicas = int32(2)
	deploymentLabels   = map[string]string{"app": "sim-deployment"}
	ingressPathType    = networkingv1beta1.PathTypeImplementationSpecific
)

func buildDeployment() appsv1.Deployment {
	return appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "sim-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: deploymentLabels,
			},
			Replicas: &deploymentReplicas,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: deploymentLabels,
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
						},
					},
				},
			},
		},
	}
}

func buildService() v1.Service {
	return v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "sim-service",
		},
		Spec: v1.ServiceSpec{
			Selector: deploymentLabels,
			Ports: []v1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.FromString(portName),
				},
			},
		},
	}
}

func buildIngress() networkingv1beta1.Ingress {
	return networkingv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name: "sim-ingress",
		},
		Spec: networkingv1beta1.IngressSpec{
			Rules: []networkingv1beta1.IngressRule{
				{
					Host: "sim-http.example.com",
					IngressRuleValue: networkingv1beta1.IngressRuleValue{
						HTTP: &networkingv1beta1.HTTPIngressRuleValue{
							Paths: []networkingv1beta1.HTTPIngressPath{
								{
									Backend: networkingv1beta1.IngressBackend{
										ServiceName: service.Name,
										ServicePort: intstr.FromInt(80),
									},
									PathType: &ingressPathType,
								},
							},
						},
					},
				},
			},
		},
	}
}

func CreateK8sApp(ctx context.Context, namespace string) {
	deployment = buildDeployment()
	service = buildService()
	ingress = buildIngress()

	common.Log.Infof("Creating Deployment %s", deployment.Name)
	deployment.SetNamespace(namespace)
	if err := common.K8sClient.Create(ctx, &deployment); err != nil {
		common.Log.Error(err)
	}

	common.Log.Infof("Creating Service %s", service.Name)
	service.SetNamespace(namespace)
	if err := common.K8sClient.Create(ctx, &service); err != nil {
		common.Log.Error(err)
	}

	common.Log.Infof("Creating Ingress %s", ingress.Name)
	ingress.SetNamespace(namespace)
	if err := common.K8sClient.Create(ctx, &ingress); err != nil {
		common.Log.Error(err)
	}
}

func DeleteK8sApp(ctx context.Context) {
	common.Log.Infof("Deleting Deployment %s", deployment.Name)
	if err := common.K8sClient.Delete(ctx, &deployment); err != nil {
		common.Log.Error(err)
	}

	common.Log.Infof("Deleting Service %s", service.Name)
	if err := common.K8sClient.Delete(ctx, &service); err != nil {
		common.Log.Error(err)
	}

	common.Log.Infof("Deleting Ingress %s", ingress.Name)
	if err := common.K8sClient.Delete(ctx, &ingress); err != nil {
		common.Log.Error(err)
	}
}

func DeleteRandomAppPod(ctx context.Context, namespace string) {
	// Get list of pods matching deploymetn selector
	podList := &v1.PodList{}
	selector, err := metav1.LabelSelectorAsSelector(deployment.Spec.Selector)
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
