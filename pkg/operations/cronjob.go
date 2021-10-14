package operations

import (
	"context"
	"time"

	"github.com/dbason/k8s-operations-sim/pkg/util"
	"github.com/rancher/opni/pkg/opnictl/common"

	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	cronJob             batchv1beta1.CronJob
	cronJobLabels       = map[string]string{"app": "sim-cronjob"}
	jobTimeLimitSeconds = int64(60)
)

const (
	desiredIterations = 3
)

func buildCronJob(cronExpression string) batchv1beta1.CronJob {
	return batchv1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name: "sim-cronjob",
		},
		Spec: batchv1beta1.CronJobSpec{
			Schedule: cronExpression,
			JobTemplate: batchv1beta1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					ActiveDeadlineSeconds: &jobTimeLimitSeconds,
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: cronJobLabels,
						},
						Spec: v1.PodSpec{
							RestartPolicy: v1.RestartPolicyNever,
							Containers: []v1.Container{
								{
									Name:            "test-job",
									Image:           "busybox",
									ImagePullPolicy: v1.PullAlways,
									Command: []string{
										"echo",
										"This is a test job",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func CreateCronJob(ctx context.Context, namespace string, interval time.Duration) {
	cronExpression := util.SplitDurationIntoCron(interval, desiredIterations)
	cronJob = buildCronJob(cronExpression)

	common.Log.Infof("Creating CronJob %s", cronJob.Name)
	cronJob.SetNamespace(namespace)
	if err := common.K8sClient.Create(ctx, &cronJob); err != nil {
		common.Log.Error(err)
	}
}

func DeleteCronJob(ctx context.Context) {
	common.Log.Infof("Deleting CronJob %s", cronJob.Name)
	if err := common.K8sClient.Delete(ctx, &cronJob); err != nil {
		common.Log.Error(err)
	}
}
