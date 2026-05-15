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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	apitypes "k8s.io/apimachinery/pkg/types"
)

func TestStorage_StoreAndGet(t *testing.T) {
	s := NewStorage()
	require.False(t, s.Ready(), "fresh storage should not be ready")

	now := time.Unix(1000, 0)
	batch := &MetricsBatch{
		PVCs: map[apitypes.NamespacedName]PVCMetricsPoint{
			{Namespace: "ns1", Name: "p1"}: {
				Timestamp:     now,
				CapacityBytes: 1024,
				UsedBytes:     512,
				HasCapacity:   true,
			},
		},
	}
	s.Store(batch)
	require.True(t, s.Ready())
	require.Equal(t, 1, s.Count())

	got, ok := s.Get(apitypes.NamespacedName{Namespace: "ns1", Name: "p1"})
	require.True(t, ok)
	require.Equal(t, uint64(1024), got.CapacityBytes)

	// Store nil clears the cache but keeps Ready true.
	s.Store(nil)
	require.True(t, s.Ready())
	require.Equal(t, 0, s.Count())
}

func TestStorage_ListNamespace(t *testing.T) {
	s := NewStorage()
	s.Store(&MetricsBatch{
		PVCs: map[apitypes.NamespacedName]PVCMetricsPoint{
			{Namespace: "a", Name: "p1"}: {CapacityBytes: 1, HasCapacity: true},
			{Namespace: "a", Name: "p2"}: {CapacityBytes: 2, HasCapacity: true},
			{Namespace: "b", Name: "p1"}: {CapacityBytes: 3, HasCapacity: true},
		},
	})
	got := s.ListNamespace("a")
	require.Len(t, got, 2)
	require.Contains(t, got, "p1")
	require.Contains(t, got, "p2")
}
