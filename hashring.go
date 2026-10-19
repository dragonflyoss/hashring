/*
 *     Copyright 2026 The Dragonfly Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package hashring provides a consistent hash ring with virtual nodes whose
// placement is bit-for-bit compatible with the dragonfly-hashring crate, so
// the dragonfly scheduler, the dragonfly client and the dragonfly-sdk select
// the same node for the same key.
package hashring

import (
	"cmp"
	"encoding/binary"
	"fmt"
	"slices"

	"github.com/dchest/siphash"
)

// DefaultVNodesPerNode is the default number of virtual nodes of each node on
// the ring, shared by the seed peer selectors of dragonfly and dragonfly-sdk.
const DefaultVNodesPerNode = 512

// VNode is a virtual node on the consistent hash ring. Each physical node is
// represented by multiple virtual nodes to better balance key distribution.
type VNode struct {
	// id is the replica index of this virtual node for its physical node.
	id int

	// name is the physical node name this virtual node represents.
	name string
}

// Name returns the physical node name associated with this virtual node.
func (v VNode) Name() string {
	return v.name
}

// String formats the virtual node as "name|id".
func (v VNode) String() string {
	return fmt.Sprintf("%s|%d", v.name, v.id)
}

// hash returns the position of the virtual node on the ring.
//
// The encoding is the Rust std Hash of the first implementation of the
// virtual node: the replica id as 8 little-endian bytes, the name bytes and
// the 0xff terminator of a string, hashed with SipHash-2-4 and zero keys.
// Changing it moves every key to another node, so it is a wire compatibility
// contract shared with the dragonfly-hashring crate.
func (v VNode) hash() uint64 {
	buf := make([]byte, 0, 8+len(v.name)+1)
	buf = binary.LittleEndian.AppendUint64(buf, uint64(v.id))
	buf = append(buf, v.name...)
	buf = append(buf, 0xff)
	return siphash.Hash(0, 0, buf)
}

// node is a virtual node with its position on the ring.
type node struct {
	// key is the position of the virtual node on the ring.
	key uint64

	// vnode is the virtual node at this position.
	vnode VNode
}

// VNodeHashRing is a consistent hash ring that uses virtual nodes to improve
// key distribution. When a physical node is added, vnodesPerNode virtual nodes
// are inserted into the ring. It is not safe for concurrent use.
type VNodeHashRing struct {
	// vnodesPerNode is the number of virtual nodes of each physical node.
	vnodesPerNode int

	// nodes is the ring sorted by key.
	nodes []node
}

// New returns a new hash ring with vnodesPerNode virtual nodes per node.
func New(vnodesPerNode int) *VNodeHashRing {
	return &VNodeHashRing{vnodesPerNode: vnodesPerNode}
}

// Add adds vnodesPerNode virtual nodes of the named node to the ring.
func (r *VNodeHashRing) Add(name string) {
	for id := range r.vnodesPerNode {
		vnode := VNode{id: id, name: name}
		r.nodes = append(r.nodes, node{key: vnode.hash(), vnode: vnode})
	}

	slices.SortFunc(r.nodes, func(a, b node) int { return cmp.Compare(a.key, b.key) })
}

// Get returns the virtual node responsible for the key, and false when the
// ring is empty.
func (r *VNodeHashRing) Get(key string) (VNode, bool) {
	if len(r.nodes) == 0 {
		return VNode{}, false
	}

	return r.nodes[r.search(hashKey(key))].vnode, true
}

// GetWithReplicas returns the virtual node responsible for the key along with
// the next replicas virtual nodes clockwise on the ring, and nil when the ring
// is empty. When replicas is larger than the length of the ring, the result
// shrinks to the entire ring.
func (r *VNodeHashRing) GetWithReplicas(key string, replicas int) []VNode {
	if len(r.nodes) == 0 {
		return nil
	}

	take := replicas + 1
	if replicas > len(r.nodes) {
		take = len(r.nodes)
	}

	start := r.search(hashKey(key))
	vnodes := make([]VNode, 0, take)
	for i := range take {
		vnodes = append(vnodes, r.nodes[(start+i)%len(r.nodes)].vnode)
	}

	return vnodes
}

// Len returns the number of virtual nodes on the ring.
func (r *VNodeHashRing) Len() int {
	return len(r.nodes)
}

// IsEmpty reports whether the ring has no virtual nodes.
func (r *VNodeHashRing) IsEmpty() bool {
	return len(r.nodes) == 0
}

// search returns the index of the first node whose key is not less than the
// key, wrapping to the first node past the end of the ring.
func (r *VNodeHashRing) search(key uint64) int {
	n, _ := slices.BinarySearchFunc(r.nodes, key, func(n node, key uint64) int { return cmp.Compare(n.key, key) })
	return n % len(r.nodes)
}

// hashKey returns the position of the key on the ring, hashed with the Rust
// std Hash encoding of a string.
func hashKey(key string) uint64 {
	buf := make([]byte, 0, len(key)+1)
	buf = append(buf, key...)
	buf = append(buf, 0xff)
	return siphash.Hash(0, 0, buf)
}
