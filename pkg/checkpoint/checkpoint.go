package checkpoint

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

const CHECKPOINT_CYCLE = time.Second * 10

type CheckpointManager struct {
	podsCheckpointMap sync.Map
	logger            *logrus.Entry
	podsMap           sync.Map //maps between pod name to pod latest checkpoint path
}

var singleton *CheckpointManager
var once sync.Once

func GetCheckpointManager(logger *logrus.Entry) *CheckpointManager {
	once.Do(func() {
		singleton = &CheckpointManager{logger: logger}
		go singleton.runPeriodicCheckpoint()
	})
	return singleton
}
func (cm *CheckpointManager) PodCreated(podName string) {
	v, ok := cm.podsMap.Load(podName)
	if ok {
		cm.logger.Info(fmt.Sprintf("Pod was already running, mutate the pod CHECKPOINT using the latest checkpoint, Pod Name: %v, checkpointName:%v", podName, v))
	} else {
		cm.logger.Info(fmt.Sprintf("First time pod running - add it to CHECKPOINT handling, Pod Name:%v", podName))
		now := time.Now()
		cm.podsMap.Store(podName, "")
		cm.podsCheckpointMap.Store(podName, &now)
	}
}

func (cm *CheckpointManager) PodDeleted(podName string) {
	cm.logger.Info(fmt.Sprintf("checkpoint manager deleting %v", podName))
	if _, ok := cm.podsMap.Load(podName); ok {
		cm.logger.Info(fmt.Sprintf("Removing pod from checkpoint cycle , Pod Name: %v", podName))
		cm.podsCheckpointMap.Delete(podName)
	} else {
		cm.logger.Info(fmt.Sprintf("couldnt find pod, Pod Name: %v", podName))
	}
}

func (cm *CheckpointManager) runPeriodicCheckpoint() {
	for {
		cm.podsCheckpointMap.Range(func(k, v interface{}) bool {
			if k == "" {
				return true
			}
			now := time.Now()
			timeValue := v.(*time.Time)
			if now.Sub(*timeValue) >= CHECKPOINT_CYCLE {
				cm.logger.Info(fmt.Sprintf("checkpointing %s", k))
				//run checkpoint
				cm.podsCheckpointMap.Store(k, &now)
			}
			return true
		})
		time.Sleep(time.Second * 5)
	}
}
