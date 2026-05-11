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
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

// DefaultAddressTypePriority matches metrics-server's default ordering:
// prefer hostname overrides, then internal DNS/IP, then external DNS/IP.
var DefaultAddressTypePriority = []corev1.NodeAddressType{
	corev1.NodeHostName,
	corev1.NodeInternalDNS,
	corev1.NodeInternalIP,
	corev1.NodeExternalDNS,
	corev1.NodeExternalIP,
}

// NodeAddressResolver picks a connection address from a node's status.
type NodeAddressResolver interface {
	NodeAddress(node *corev1.Node) (string, error)
}

type prioNodeAddrResolver struct {
	addrTypePriority []corev1.NodeAddressType
}

func (r *prioNodeAddrResolver) NodeAddress(node *corev1.Node) (string, error) {
	for _, addrType := range r.addrTypePriority {
		for _, addr := range node.Status.Addresses {
			if addr.Type == addrType {
				return addr.Address, nil
			}
		}
	}
	return "", fmt.Errorf("no address matched types %v", r.addrTypePriority)
}

// NewPriorityNodeAddressResolver returns a resolver that walks the priority
// list in order and returns the first matching address.
func NewPriorityNodeAddressResolver(typePriority []corev1.NodeAddressType) NodeAddressResolver {
	return &prioNodeAddrResolver{addrTypePriority: typePriority}
}
