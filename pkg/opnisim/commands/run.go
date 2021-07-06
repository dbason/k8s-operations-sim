package commands

import (
	"os"
	"time"

	"github.com/dbason/k8s-operations-sim/pkg/operations"
	"github.com/rancher/opni/pkg/opnictl/common"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultNamespace = "opni-sim"
	defaultInterval  = "5m"
)

var (
	intervalString string
	namespace      string
)

func BuildRunCmd() *cobra.Command {
	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Run the opni k8s simulation tool",
		RunE:  doK8sOperations,
	}

	runCmd.Flags().StringVar(&intervalString, "interval", defaultInterval, "interval in minutes to run operations")
	runCmd.Flags().StringVar(&namespace, "namespace", defaultNamespace, "namespace to create objects in (will be created)")

	return runCmd
}

func doK8sOperations(cmd *cobra.Command, args []string) error {
	common.Log.Info("Starting k8s operations")
	operationInterval, err := time.ParseDuration(intervalString)
	if err != nil {
		common.Log.Errorf("invalid interval string %v", err)
		os.Exit(1)
	}
	if err := common.K8sClient.Create(cmd.Context(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}); errors.IsAlreadyExists(err) {
		common.Log.Debug(err)
	} else if err != nil {
		return err
	}
	common.Log.Infof("Namespace %s created", namespace)

	shouldDeleteObjects := newShouldDeleteObjects()
	doDelete := func() {
		operations.DeleteK8sApp(cmd.Context())
		operations.DeleteDaemonSet(cmd.Context())
		operations.DeleteCronJob(cmd.Context())
	}
	doCreate := func() {
		operations.CreateK8sApp(namespace, cmd.Context())
		operations.CreateDaemonSet(namespace, cmd.Context())
		operations.CreateCronJob(namespace, cmd.Context(), operationInterval)
		time.Sleep(time.Duration(15) * time.Second)
		operations.DeleteRandomAppPod(namespace, cmd.Context())
		operations.DeleteRandomDaemonSetPod(namespace, cmd.Context())
	}

	doCreate()

	for {
		select {
		case <-time.After(operationInterval):
			if shouldDeleteObjects() {
				doDelete()
			} else {
				doCreate()
			}
		case <-cmd.Context().Done():
			return nil
		}

	}
}

func newShouldDeleteObjects() func() bool {
	deleteObjects := false
	return func() bool {
		deleteObjects = !deleteObjects
		return deleteObjects
	}
}
