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

package hashring

import (
	"cmp"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	mockNodes = []string{"default-seed-client-1", "default-seed-client-2", "default-seed-client-3"}
)

func TestVNode_String(t *testing.T) {
	tests := []struct {
		name   string
		vnode  VNode
		expect func(t *testing.T, vnode VNode)
	}{
		{
			name:  "format virtual node as name and id",
			vnode: VNode{id: 1, name: "default-seed-client-1"},
			expect: func(t *testing.T, vnode VNode) {
				assert := assert.New(t)
				assert.Equal("default-seed-client-1|1", vnode.String())
				assert.Equal("default-seed-client-1", vnode.Name())
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.expect(t, tc.vnode)
		})
	}
}

func TestVNodeHashRing_Add(t *testing.T) {
	tests := []struct {
		name          string
		vnodesPerNode int
		nodes         []string
		expect        func(t *testing.T, ring *VNodeHashRing)
	}{
		{
			name:          "empty ring",
			vnodesPerNode: 3,
			nodes:         []string{},
			expect: func(t *testing.T, ring *VNodeHashRing) {
				assert := assert.New(t)
				assert.True(ring.IsEmpty())
				assert.Equal(0, ring.Len())
			},
		},
		{
			name:  "default vnodes per node",
			nodes: []string{"default-seed-client-1"},
			expect: func(t *testing.T, ring *VNodeHashRing) {
				assert := assert.New(t)
				assert.Equal(DefaultVNodesPerNode, ring.Len())
			},
		},
		{
			name:          "add one node",
			vnodesPerNode: 2,
			nodes:         []string{"default-seed-client-1"},
			expect: func(t *testing.T, ring *VNodeHashRing) {
				assert := assert.New(t)
				assert.False(ring.IsEmpty())
				assert.Equal(2, ring.Len())
			},
		},
		{
			name:          "add two nodes",
			vnodesPerNode: 2,
			nodes:         []string{"default-seed-client-1", "default-seed-client-2"},
			expect: func(t *testing.T, ring *VNodeHashRing) {
				assert := assert.New(t)
				assert.Equal(4, ring.Len())
				assert.True(slices.IsSortedFunc(ring.nodes, func(a, b node) int { return cmp.Compare(a.key, b.key) }))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var options []Option
			if tc.vnodesPerNode != 0 {
				options = append(options, WithVNodesPerNode(tc.vnodesPerNode))
			}

			ring := New(options...)
			for _, name := range tc.nodes {
				ring.Add(name)
			}

			tc.expect(t, ring)
		})
	}
}

func TestVNodeHashRing_Get(t *testing.T) {
	tests := []struct {
		name   string
		nodes  []string
		expect func(t *testing.T, ring *VNodeHashRing)
	}{
		{
			name:  "empty ring",
			nodes: []string{},
			expect: func(t *testing.T, ring *VNodeHashRing) {
				assert := assert.New(t)
				_, ok := ring.Get("task")
				assert.False(ok)
			},
		},
		{
			name:  "single node owns every key",
			nodes: []string{"default-seed-client-1"},
			expect: func(t *testing.T, ring *VNodeHashRing) {
				assert := assert.New(t)
				for _, key := range []string{"", "task-a", "task-b"} {
					vnode, ok := ring.Get(key)
					assert.True(ok)
					assert.Equal("default-seed-client-1", vnode.Name())
				}
			},
		},
		{
			name:  "same key maps to the same virtual node",
			nodes: mockNodes,
			expect: func(t *testing.T, ring *VNodeHashRing) {
				assert := assert.New(t)
				first, _ := ring.Get("task")
				second, _ := ring.Get("task")
				assert.Equal(first, second)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ring := New(WithVNodesPerNode(2))
			for _, name := range tc.nodes {
				ring.Add(name)
			}

			tc.expect(t, ring)
		})
	}
}

func TestVNodeHashRing_GetWithReplicas(t *testing.T) {
	tests := []struct {
		name     string
		nodes    []string
		replicas int
		expect   func(t *testing.T, vnodes []VNode)
	}{
		{
			name:     "empty ring",
			nodes:    []string{},
			replicas: 2,
			expect: func(t *testing.T, vnodes []VNode) {
				assert := assert.New(t)
				assert.Nil(vnodes)
			},
		},
		{
			name:     "replicas within the ring",
			nodes:    []string{"default-seed-client-1", "default-seed-client-2"},
			replicas: 2,
			expect: func(t *testing.T, vnodes []VNode) {
				assert := assert.New(t)
				assert.Len(vnodes, 3)
			},
		},
		{
			name:     "replicas larger than the ring",
			nodes:    []string{"default-seed-client-1", "default-seed-client-2"},
			replicas: 100,
			expect: func(t *testing.T, vnodes []VNode) {
				assert := assert.New(t)
				assert.Len(vnodes, 4)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ring := New(WithVNodesPerNode(2))
			for _, name := range tc.nodes {
				ring.Add(name)
			}

			tc.expect(t, ring.GetWithReplicas("task", tc.replicas))
		})
	}
}

func TestVNodeHashRing_AddOrder(t *testing.T) {
	tests := []struct {
		name   string
		keys   []string
		expect func(t *testing.T, ringA, ringB *VNodeHashRing, key string)
	}{
		{
			name: "add order does not affect get",
			keys: []string{"task-a", "task-b", "task-c", "task-d", "task-e"},
			expect: func(t *testing.T, ringA, ringB *VNodeHashRing, key string) {
				assert := assert.New(t)
				vnodeA, _ := ringA.Get(key)
				vnodeB, _ := ringB.Get(key)
				assert.Equal(vnodeA, vnodeB)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ringA := New()
			for _, name := range mockNodes {
				ringA.Add(name)
			}

			ringB := New()
			for _, name := range slices.Backward(mockNodes) {
				ringB.Add(name)
			}

			for _, key := range tc.keys {
				tc.expect(t, ringA, ringB, key)
			}
		})
	}
}
