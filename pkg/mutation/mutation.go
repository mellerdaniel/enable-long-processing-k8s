package mutation

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

// Mutator is a container for mutation
type Mutator struct {
	Logger *logrus.Entry
}

// NewMutator returns an initialised instance of Mutator
func NewMutator(logger *logrus.Entry) *Mutator {
	return &Mutator{Logger: logger}
}

// podMutators is an interface used to group functions mutating pods
type podMutator interface {
	Mutate(*corev1.Pod) (*corev1.Pod, error)
	Name() string
}

// MutatePodPatch returns a json patch containing all the mutations needed for
// a given pod
func (m *Mutator) MutatePodPatch(pod *corev1.Pod) ([]byte, error) {
	var podName string
	if pod.Name != "" {
		podName = pod.Name
	} else {
		if pod.ObjectMeta.GenerateName != "" {
			podName = pod.ObjectMeta.GenerateName
		}
	}

	log := logrus.WithField("pod_name", podName)
	mpod := pod.DeepCopy()
	p := []map[string]string{}
	for i := range mpod.Spec.Containers {
		log.Info(fmt.Sprintf("currently working on container number %d", i))
		patchMap := map[string]string{
			"op":    "replace",
			"path":  fmt.Sprintf("/spec/containers/%d/image", i),
			"value": "debian",
		}
		p = append(p, patchMap)
	}
	patchb, err := json.Marshal(p)
	if err != nil {
		log.Error("unable to marshal", err)
		return nil, err
	}
	return patchb, nil
}
