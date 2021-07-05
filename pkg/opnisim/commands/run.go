package commands

import (
	"time"

	"github.com/dbason/k8s-operations-sim/pkg/operations"
	"github.com/rancher/opni/pkg/opnictl/common"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	simNamespace = "opni-sim"
)

var (
	deleteObjects = true
)

func BuildRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Run the opni k8s simulation tool",
		RunE:  doK8sOperations,
	}
}

func doK8sOperations(cmd *cobra.Command, args []string) error {
	common.Log.Info("Starting k8s operations")
	operationInterval := time.Duration(1) * time.Minute
	if err := common.K8sClient.Create(cmd.Context(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: simNamespace,
		},
	}); errors.IsAlreadyExists(err) {
		common.Log.Debug(err)
	} else if err != nil {
		return err
	}
	common.Log.Infof("Namespace %s created", simNamespace)

	shouldDeleteObjects := newShouldDeleteObjects()
	doDelete := func() {
		operations.DeleteK8sApp(simNamespace, cmd.Context())
	}
	doCreate := func() {
		operations.CreateK8sApp(simNamespace, cmd.Context())
		time.Sleep(time.Duration(30) * time.Second)
		operations.DeleteRandomAppPod(simNamespace, cmd.Context())
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
