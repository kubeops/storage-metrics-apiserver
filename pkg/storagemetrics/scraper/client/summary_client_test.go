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

package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apitypes "k8s.io/apimachinery/pkg/types"
	stats "k8s.io/kubelet/pkg/apis/stats/v1alpha1"
)

func ptrU64(v uint64) *uint64 { return &v }

func TestSummaryToBatch_PVCsOnly(t *testing.T) {
	now := time.Unix(2000, 0)
	summary := &stats.Summary{
		Pods: []stats.PodStats{{
			VolumeStats: []stats.VolumeStats{
				{
					Name:    "data",
					PVCRef:  &stats.PVCReference{Namespace: "ns1", Name: "data-pvc"},
					FsStats: stats.FsStats{Time: metav1.NewTime(now), CapacityBytes: ptrU64(2048), AvailableBytes: ptrU64(1024), UsedBytes: ptrU64(1024), Inodes: ptrU64(100), InodesFree: ptrU64(70), InodesUsed: ptrU64(30)},
				},
				// emptyDir has no PVCRef and must be skipped.
				{Name: "scratch", FsStats: stats.FsStats{CapacityBytes: ptrU64(8192)}},
				// driver returned empty stats — skip.
				{Name: "broken", PVCRef: &stats.PVCReference{Namespace: "ns1", Name: "broken-pvc"}},
			},
		}},
	}
	batch := summaryToBatch(summary, time.Unix(1000, 0), "node-a")
	require.Len(t, batch.PVCs, 1)
	got, ok := batch.PVCs[apitypes.NamespacedName{Namespace: "ns1", Name: "data-pvc"}]
	require.True(t, ok)
	require.Equal(t, "node-a", got.Node)
	require.Equal(t, uint64(2048), got.CapacityBytes)
	require.True(t, got.HasCapacity)
	require.True(t, got.HasInodes)
	require.Equal(t, now, got.Timestamp)
}

func TestSummaryToBatch_DefaultTimestampOnZeroFsTime(t *testing.T) {
	defaultTime := time.Unix(5000, 0)
	summary := &stats.Summary{
		Pods: []stats.PodStats{{
			VolumeStats: []stats.VolumeStats{{
				PVCRef:  &stats.PVCReference{Namespace: "ns", Name: "pvc"},
				FsStats: stats.FsStats{CapacityBytes: ptrU64(1)},
			}},
		}},
	}
	batch := summaryToBatch(summary, defaultTime, "node-a")
	require.Len(t, batch.PVCs, 1)
	require.Equal(t, defaultTime, batch.PVCs[apitypes.NamespacedName{Namespace: "ns", Name: "pvc"}].Timestamp)
}

func TestSummaryToBatch_KeepsFreshestPerPVC(t *testing.T) {
	older := time.Unix(1000, 0)
	newer := time.Unix(2000, 0)
	summary := &stats.Summary{
		Pods: []stats.PodStats{
			{VolumeStats: []stats.VolumeStats{{
				PVCRef:  &stats.PVCReference{Namespace: "ns", Name: "pvc"},
				FsStats: stats.FsStats{Time: metav1.NewTime(older), CapacityBytes: ptrU64(1), UsedBytes: ptrU64(1)},
			}}},
			{VolumeStats: []stats.VolumeStats{{
				PVCRef:  &stats.PVCReference{Namespace: "ns", Name: "pvc"},
				FsStats: stats.FsStats{Time: metav1.NewTime(newer), CapacityBytes: ptrU64(2), UsedBytes: ptrU64(2)},
			}}},
		},
	}
	batch := summaryToBatch(summary, time.Unix(0, 0), "node-a")
	require.Equal(t, uint64(2), batch.PVCs[apitypes.NamespacedName{Namespace: "ns", Name: "pvc"}].CapacityBytes)
}

func TestSummaryClient_GetMetrics_HTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/stats/summary", r.URL.Path)
		_, _ = w.Write([]byte(`{
            "node": {"nodeName": "n1"},
            "pods": [{
              "podRef": {"name": "p", "namespace": "ns"},
              "volume": [{
                "name": "data",
                "pvcRef": {"name": "pvc1", "namespace": "ns"},
                "capacityBytes": 1024,
                "availableBytes": 512,
                "usedBytes": 512,
                "time": "1970-01-01T00:00:01Z"
              }]
            }]
        }`))
	}))
	defer srv.Close()

	c := newClient(srv.Client(), &fixedAddr{addr: srvHost(srv.URL)},
		srvPort(srv.URL), "http", false)
	node := &corev1.Node{}
	batch, err := c.GetMetrics(context.Background(), node)
	require.NoError(t, err)
	require.Len(t, batch.PVCs, 1)
	got := batch.PVCs[apitypes.NamespacedName{Namespace: "ns", Name: "pvc1"}]
	require.Equal(t, uint64(1024), got.CapacityBytes)
	require.Equal(t, uint64(512), got.UsedBytes)
}

type fixedAddr struct{ addr string }

func (f *fixedAddr) NodeAddress(*corev1.Node) (string, error) { return f.addr, nil }

func srvHost(u string) string { return parseHostPort(u).host }
func srvPort(u string) int    { return parseHostPort(u).port }

type hostPort struct {
	host string
	port int
}

func parseHostPort(rawURL string) hostPort {
	// rawURL is like http://127.0.0.1:54321
	stripped := rawURL[len("http://"):]
	for i := 0; i < len(stripped); i++ {
		if stripped[i] == ':' {
			port := 0
			for j := i + 1; j < len(stripped); j++ {
				port = port*10 + int(stripped[j]-'0')
			}
			return hostPort{host: stripped[:i], port: port}
		}
	}
	return hostPort{host: stripped}
}
