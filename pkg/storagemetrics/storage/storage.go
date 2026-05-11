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
	"sync"

	apitypes "k8s.io/apimachinery/pkg/types"
)

// Storage holds the most recently observed batch of PVC metrics and serves
// it via a read-only RWMutex. Each Store call atomically replaces the map —
// readers never see partial writes.
//
// The implementation matches the metrics-server pattern: there is no
// time-window aggregation because PVC FS stats are gauges, not counters.
type Storage struct {
	mu    sync.RWMutex
	pvcs  map[apitypes.NamespacedName]PVCMetricsPoint
	ready bool
}

func NewStorage() *Storage {
	return &Storage{
		pvcs: map[apitypes.NamespacedName]PVCMetricsPoint{},
	}
}

// Store atomically replaces the cached batch with the new one.
// A nil batch is treated as an empty batch.
func (s *Storage) Store(batch *MetricsBatch) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if batch == nil || batch.PVCs == nil {
		s.pvcs = map[apitypes.NamespacedName]PVCMetricsPoint{}
	} else {
		s.pvcs = batch.PVCs
	}
	s.ready = true
}

// Ready reports whether at least one scrape tick has completed.
func (s *Storage) Ready() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ready
}

// Get returns the most recent point for a single PVC, or false if none is cached.
func (s *Storage) Get(key apitypes.NamespacedName) (PVCMetricsPoint, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.pvcs[key]
	return p, ok
}

// ListNamespace returns all cached points in a namespace. Caller must not mutate.
func (s *Storage) ListNamespace(namespace string) map[string]PVCMetricsPoint {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]PVCMetricsPoint)
	for k, v := range s.pvcs {
		if k.Namespace == namespace {
			out[k.Name] = v
		}
	}
	return out
}

// Count returns the number of cached PVC points (for metrics / debugging).
func (s *Storage) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.pvcs)
}
