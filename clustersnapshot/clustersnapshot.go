package clustersnapshot

import (
	"errors"

	apiv1 "k8s.io/api/core/v1"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// ClusterSnapshot is abstraction of cluster state used for predicate simulations.
// It exposes mutation methods and can be viewed as scheduler's SharedLister.
type ClusterSnapshot interface {
	schedulerframework.SharedLister
	// AddNode adds node to the snapshot.
	AddNode(node *apiv1.Node) error
	// AddNodes adds nodes to the snapshot.
	AddNodes(nodes []*apiv1.Node) error
	// RemoveNode removes nodes (and pods scheduled to it) from the snapshot.
	RemoveNode(nodeName string) error
	// AddPod adds pod to the snapshot and schedules it to given node.
	AddPod(pod *apiv1.Pod, nodeName string) error
	// RemovePod removes pod from the snapshot.
	RemovePod(namespace string, podName string, nodeName string) error
	// AddNodeWithPods adds a node and set of pods to be scheduled to this node to the snapshot.
	AddNodeWithPods(node *apiv1.Node, pods []*apiv1.Pod) error
	// IsPVCUsedByPods returns if the pvc is used by any pod, key = <namespace>/<pvc_name>
	IsPVCUsedByPods(key string) bool

	// Fork creates a fork of snapshot state. All modifications can later be reverted to moment of forking via Revert().
	// Use WithForkedSnapshot() helper function instead if possible.
	Fork()
	// Revert reverts snapshot state to moment of forking.
	Revert()
	// Commit commits changes done after forking.
	Commit() error
	// Clear reset cluster snapshot to empty, unforked state.
	Clear()
}

// ErrNodeNotFound means that a node wasn't found in the snapshot.
var ErrNodeNotFound = errors.New("node not found")

func InitializeClusterSnapshotOrDie(
	snapshot ClusterSnapshot,
	nodes []apiv1.Node,
	pods []apiv1.Pod) {
	var err error

	snapshot.Clear()

	for i := 0; i < len(nodes); i++ {
		node := nodes[i]
		err = snapshot.AddNode(&node)
	}

	for i := 0; i < len(pods); i++ {
		pod := pods[i]
		if pod.Spec.NodeName != "" {
			err = snapshot.AddPod(&pod, pod.Spec.NodeName)
		} else if pod.Status.NominatedNodeName != "" {
			err = snapshot.AddPod(&pod, pod.Status.NominatedNodeName)
		} else {
			panic(err)
		}
	}
}
