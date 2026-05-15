/*
Copyright 2026 AppsCode Inc.

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

package storage

import (
	"time"

	apitypes "k8s.io/apimachinery/pkg/types"
)

// PVCMetricsPoint is a single observation of one PVC's filesystem stats.
// Values mirror the FsStats returned by kubelet's /stats/summary endpoint
// (which works for any volume kubelet mounts as a filesystem PVC, including
// in-tree, external CSI, and migrated drivers — regardless of whether the
// driver implements CSI NodeGetVolumeStats).
type PVCMetricsPoint struct {
	// Node that reported the stats. Used for deduplication and debugging.
	Node string

	Timestamp time.Time

	CapacityBytes  uint64
	AvailableBytes uint64
	UsedBytes      uint64

	Inodes      uint64
	InodesFree  uint64
	InodesUsed  uint64
	HasInodes   bool
	HasCapacity bool
}

// MetricsBatch is a snapshot of all PVC stats collected during one scrape tick.
type MetricsBatch struct {
	// PVCs is keyed by {namespace, name} of the PersistentVolumeClaim.
	PVCs map[apitypes.NamespacedName]PVCMetricsPoint
}
