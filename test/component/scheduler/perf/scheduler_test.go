/*
Copyright 2015 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package benchmark

import (
	"fmt"
	"k8s.io/kubernetes/pkg/api"
	"testing"
	"time"
)

var nNode int = 500
var nPod int = 20
var maxFailedPod int = 10

// TestSchedule100Node3KPods schedules 3k pods on 100 nodes.
func TestSchedule100Node3KPods(t *testing.T) {
	//schedulePods(100, 3000)
	schedulePods(nNode, nPod)
}

// TestSchedule1000Node30KPods schedules 30k pods on 1000 nodes.
//func TestSchedule1000Node30KPods(t *testing.T) {
//	schedulePods(1000, 30000)
//}

// schedulePods schedules specific number of pods on specific number of nodes.
// This is used to learn the scheduling throughput on various
// sizes of cluster and changes as more and more pods are scheduled.
// It won't stop until all pods are scheduled.
func schedulePods(numNodes, numPods int) {
	schedulerConfigFactory, destroyFunc := mustSetupScheduler()
	defer destroyFunc()
	c := schedulerConfigFactory.Client

	fmt.Printf("====== will create %d nodes =======\n", numNodes)
	makeNodes(c, numNodes)

	// makePods(c, numPods)

	prev := 0
	prev_fail := 0
	//start := time.Now()
	for {
		// This can potentially affect performance of scheduler, since List() is done under mutex.
		// Listing 10000 pods is an expensive operation, so running it frequently may impact scheduler.
		// TODO: Setup watch on apiserver and wait until all pods scheduled.

		// try to create 100 pods

		makePods(c, numPods)

		time.Sleep(5 * time.Second)

		scheduled := schedulerConfigFactory.ScheduledPodLister.Store.List()

		scheduled_num := len(scheduled)

		//fmt.Printf("%ds\trate: %d\ttotal: %d\n", time.Since(start)/time.Second, len(scheduled)-prev, len(scheduled))
		//		if len(scheduled) >= numPods {
		//			return
		//		}

		fmt.Printf("======= total pods scheduled %d, this time scheduled:%d ===========\n", len(scheduled), scheduled_num-prev)

		fmt.Printf("failed this time:%d, total failed:%d.\n", numPods-(scheduled_num-prev), prev_fail + numPods-(scheduled_num-prev))

		if prev_fail + numPods - (scheduled_num-prev) >= maxFailedPod {

			fmt.Printf("can not create pods , will statics resource utility.\n")

			var totalCPU int64
			var totalMemory int64

			totalCPU = 0
			totalMemory = 0

			for _, v := range scheduled {

				pod := v.(*api.Pod)
				//				fmt.Printf("cpu of pod:%d,memory of pod:%d.assigned to node:%s.\n",
				//					pod.Spec.Containers[0].Resources.Requests.Cpu().MilliValue(),
				//					pod.Spec.Containers[0].Resources.Requests.Memory().Value()/(1024*1024),
				//					pod.Spec.NodeName)
				totalCPU = totalCPU + pod.Spec.Containers[0].Resources.Requests.Cpu().MilliValue()
				totalMemory = totalMemory + pod.Spec.Containers[0].Resources.Requests.Memory().Value()
			}

			fmt.Printf("\n>>>>\nTest: %v, totalPod=%v, NodeNum=%v, PodPerRound=%v\n", time.Now(), len(scheduled), nNode, nPod)
			fmt.Printf("toalCPU:%d,totalMemory:%d.\n", totalCPU/1000, totalMemory/(1024*1024))
			fmt.Printf("cpu utility:%f,memory:%f.\n", float32(totalCPU)/1000.0/(float32(numNodes)*4),
				float32(totalMemory)/(1024*1024*1024)/(float32(numNodes)*32))

			//			// print nodes info
			//			nodes := schedulerConfigFactory.NodeLister.Store.List()
			//			for _, nd := range nodes {
			//				n := nd.(*api.Node)
			//				fmt.Printf("nodeName: %v\n", n.Name)
			//			}

			return
		}

		prev_fail = prev_fail + numPods - (scheduled_num - prev)
		prev = scheduled_num
		//time.Sleep(1 * time.Second)
	}
}
