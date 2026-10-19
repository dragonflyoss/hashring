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

// Cross-language consistency vectors shared with the Rust crate
// (tests/consistency.rs). Both test suites assert the same fixed outputs, so a
// change on either side that moves a key to another virtual node fails the
// build.

package hashring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	mockSeedPeers   = []string{"seed-peer-1", "seed-peer-2", "seed-peer-3"}
	mockSeedClients = []string{"default-seed-client-0", "default-seed-client-1", "default-seed-client-2"}
)

func TestSiphash(t *testing.T) {
	tests := []struct {
		name   string
		hash   uint64
		expect func(t *testing.T, hash uint64)
	}{
		{
			name: "first virtual node of seed-peer-1",
			hash: VNode{id: 0, name: "seed-peer-1"}.hash(),
			expect: func(t *testing.T, hash uint64) {
				assert := assert.New(t)
				assert.Equal(uint64(0x1e5a582b8d945969), hash)
			},
		},
		{
			name: "second virtual node of seed-peer-1",
			hash: VNode{id: 1, name: "seed-peer-1"}.hash(),
			expect: func(t *testing.T, hash uint64) {
				assert := assert.New(t)
				assert.Equal(uint64(0xb5db98265419376c), hash)
			},
		},
		{
			name: "first virtual node of seed-peer-2",
			hash: VNode{id: 0, name: "seed-peer-2"}.hash(),
			expect: func(t *testing.T, hash uint64) {
				assert := assert.New(t)
				assert.Equal(uint64(0xf36f748c486b09ef), hash)
			},
		},
		{
			name: "last virtual node of seed-peer-3",
			hash: VNode{id: 511, name: "seed-peer-3"}.hash(),
			expect: func(t *testing.T, hash uint64) {
				assert := assert.New(t)
				assert.Equal(uint64(0xf0eda07426c5cac1), hash)
			},
		},
		{
			name: "task id key",
			hash: hashKey("b2c366cce7e68013d5441c6326d5a3e1b12aeb5ed58564d0fd3fa089bc29cb6e"),
			expect: func(t *testing.T, hash uint64) {
				assert := assert.New(t)
				assert.Equal(uint64(0xaa86f08bc4878a68), hash)
			},
		},
		{
			name: "plain key",
			hash: hashKey("test-task"),
			expect: func(t *testing.T, hash uint64) {
				assert := assert.New(t)
				assert.Equal(uint64(0xa78dd821b4023714), hash)
			},
		},
		{
			name: "empty key",
			hash: hashKey(""),
			expect: func(t *testing.T, hash uint64) {
				assert := assert.New(t)
				assert.Equal(uint64(0x20dd4eb33d9590f2), hash)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.expect(t, tc.hash)
		})
	}
}

func TestHashring(t *testing.T) {
	strings := func(vnodes []VNode) []string {
		names := make([]string, 0, len(vnodes))
		for _, vnode := range vnodes {
			names = append(names, vnode.String())
		}

		return names
	}

	tests := []struct {
		name          string
		vnodesPerNode int
		nodes         []string
		key           string
		replicas      int
		expect        func(t *testing.T, vnodes []VNode)
	}{
		{
			name:          "task-a",
			vnodesPerNode: 3,
			nodes:         mockSeedPeers,
			key:           "task-a",
			replicas:      2,
			expect: func(t *testing.T, vnodes []VNode) {
				assert := assert.New(t)
				assert.Equal([]string{"seed-peer-3|2", "seed-peer-1|2", "seed-peer-3|0"}, strings(vnodes))
			},
		},
		{
			name:          "task-b",
			vnodesPerNode: 3,
			nodes:         mockSeedPeers,
			key:           "task-b",
			replicas:      2,
			expect: func(t *testing.T, vnodes []VNode) {
				assert := assert.New(t)
				assert.Equal([]string{"seed-peer-2|0", "seed-peer-2|1", "seed-peer-1|0"}, strings(vnodes))
			},
		},
		{
			name:          "task-c",
			vnodesPerNode: 3,
			nodes:         mockSeedPeers,
			key:           "task-c",
			replicas:      2,
			expect: func(t *testing.T, vnodes []VNode) {
				assert := assert.New(t)
				assert.Equal([]string{"seed-peer-2|1", "seed-peer-1|0", "seed-peer-3|2"}, strings(vnodes))
			},
		},
		{
			name:          "task-a with 100 replicas",
			vnodesPerNode: 3,
			nodes:         mockSeedPeers,
			key:           "task-a",
			replicas:      100,
			expect: func(t *testing.T, vnodes []VNode) {
				assert := assert.New(t)
				assert.Equal([]string{
					"seed-peer-3|2",
					"seed-peer-1|2",
					"seed-peer-3|0",
					"seed-peer-1|1",
					"seed-peer-2|2",
					"seed-peer-3|1",
					"seed-peer-2|0",
					"seed-peer-2|1",
					"seed-peer-1|0",
				}, strings(vnodes))
			},
		},
		{
			name:          "task 0000 with default vnodes per node",
			vnodesPerNode: DefaultVNodesPerNode,
			nodes:         mockSeedClients,
			key:           "0000000000000000000000000000000000000000000000000000000000000000",
			replicas:      2,
			expect: func(t *testing.T, vnodes []VNode) {
				assert := assert.New(t)
				assert.Equal([]string{
					"default-seed-client-1|260",
					"default-seed-client-0|383",
					"default-seed-client-0|281",
				}, strings(vnodes))
			},
		},
		{
			name:          "task 1eef with default vnodes per node",
			vnodesPerNode: DefaultVNodesPerNode,
			nodes:         mockSeedClients,
			key:           "0000000000000000000000000000000000000000000000000000000000001eef",
			replicas:      2,
			expect: func(t *testing.T, vnodes []VNode) {
				assert := assert.New(t)
				assert.Equal([]string{
					"default-seed-client-1|375",
					"default-seed-client-2|109",
					"default-seed-client-0|483",
				}, strings(vnodes))
			},
		},
		{
			name:          "task 3dde with default vnodes per node",
			vnodesPerNode: DefaultVNodesPerNode,
			nodes:         mockSeedClients,
			key:           "0000000000000000000000000000000000000000000000000000000000003dde",
			replicas:      2,
			expect: func(t *testing.T, vnodes []VNode) {
				assert := assert.New(t)
				assert.Equal([]string{
					"default-seed-client-2|366",
					"default-seed-client-0|374",
					"default-seed-client-1|30",
				}, strings(vnodes))
			},
		},
		{
			name:          "task 5ccd with default vnodes per node",
			vnodesPerNode: DefaultVNodesPerNode,
			nodes:         mockSeedClients,
			key:           "0000000000000000000000000000000000000000000000000000000000005ccd",
			replicas:      2,
			expect: func(t *testing.T, vnodes []VNode) {
				assert := assert.New(t)
				assert.Equal([]string{
					"default-seed-client-0|69",
					"default-seed-client-0|121",
					"default-seed-client-2|243",
				}, strings(vnodes))
			},
		},
		{
			name:          "task 7bbc with default vnodes per node",
			vnodesPerNode: DefaultVNodesPerNode,
			nodes:         mockSeedClients,
			key:           "0000000000000000000000000000000000000000000000000000000000007bbc",
			replicas:      2,
			expect: func(t *testing.T, vnodes []VNode) {
				assert := assert.New(t)
				assert.Equal([]string{
					"default-seed-client-0|239",
					"default-seed-client-1|156",
					"default-seed-client-1|157",
				}, strings(vnodes))
			},
		},
		{
			name:          "task 9aab with default vnodes per node",
			vnodesPerNode: DefaultVNodesPerNode,
			nodes:         mockSeedClients,
			key:           "0000000000000000000000000000000000000000000000000000000000009aab",
			replicas:      2,
			expect: func(t *testing.T, vnodes []VNode) {
				assert := assert.New(t)
				assert.Equal([]string{
					"default-seed-client-1|501",
					"default-seed-client-0|435",
					"default-seed-client-1|308",
				}, strings(vnodes))
			},
		},
		{
			name:          "task b99a with default vnodes per node",
			vnodesPerNode: DefaultVNodesPerNode,
			nodes:         mockSeedClients,
			key:           "000000000000000000000000000000000000000000000000000000000000b99a",
			replicas:      2,
			expect: func(t *testing.T, vnodes []VNode) {
				assert := assert.New(t)
				assert.Equal([]string{
					"default-seed-client-0|223",
					"default-seed-client-0|9",
					"default-seed-client-2|83",
				}, strings(vnodes))
			},
		},
		{
			name:          "task d889 with default vnodes per node",
			vnodesPerNode: DefaultVNodesPerNode,
			nodes:         mockSeedClients,
			key:           "000000000000000000000000000000000000000000000000000000000000d889",
			replicas:      2,
			expect: func(t *testing.T, vnodes []VNode) {
				assert := assert.New(t)
				assert.Equal([]string{
					"default-seed-client-1|429",
					"default-seed-client-2|220",
					"default-seed-client-1|220",
				}, strings(vnodes))
			},
		},
		{
			name:          "task f778 with default vnodes per node",
			vnodesPerNode: DefaultVNodesPerNode,
			nodes:         mockSeedClients,
			key:           "000000000000000000000000000000000000000000000000000000000000f778",
			replicas:      2,
			expect: func(t *testing.T, vnodes []VNode) {
				assert := assert.New(t)
				assert.Equal([]string{
					"default-seed-client-2|326",
					"default-seed-client-1|155",
					"default-seed-client-0|473",
				}, strings(vnodes))
			},
		},
		{
			name:          "task 1667 with default vnodes per node",
			vnodesPerNode: DefaultVNodesPerNode,
			nodes:         mockSeedClients,
			key:           "0000000000000000000000000000000000000000000000000000000000011667",
			replicas:      2,
			expect: func(t *testing.T, vnodes []VNode) {
				assert := assert.New(t)
				assert.Equal([]string{
					"default-seed-client-0|293",
					"default-seed-client-1|334",
					"default-seed-client-2|199",
				}, strings(vnodes))
			},
		},
		{
			name:          "task 3556 with default vnodes per node",
			vnodesPerNode: DefaultVNodesPerNode,
			nodes:         mockSeedClients,
			key:           "0000000000000000000000000000000000000000000000000000000000013556",
			replicas:      2,
			expect: func(t *testing.T, vnodes []VNode) {
				assert := assert.New(t)
				assert.Equal([]string{
					"default-seed-client-2|246",
					"default-seed-client-2|65",
					"default-seed-client-0|369",
				}, strings(vnodes))
			},
		},
		{
			name:          "task 5445 with default vnodes per node",
			vnodesPerNode: DefaultVNodesPerNode,
			nodes:         mockSeedClients,
			key:           "0000000000000000000000000000000000000000000000000000000000015445",
			replicas:      2,
			expect: func(t *testing.T, vnodes []VNode) {
				assert := assert.New(t)
				assert.Equal([]string{
					"default-seed-client-2|480",
					"default-seed-client-1|24",
					"default-seed-client-1|147",
				}, strings(vnodes))
			},
		},
		{
			name:          "empty key with default vnodes per node",
			vnodesPerNode: DefaultVNodesPerNode,
			nodes:         mockSeedClients,
			key:           "",
			replicas:      2,
			expect: func(t *testing.T, vnodes []VNode) {
				assert := assert.New(t)
				assert.Equal([]string{
					"default-seed-client-1|163",
					"default-seed-client-2|24",
					"default-seed-client-1|252",
				}, strings(vnodes))
			},
		},
		{
			name:          "key a with default vnodes per node",
			vnodesPerNode: DefaultVNodesPerNode,
			nodes:         mockSeedClients,
			key:           "a",
			replicas:      2,
			expect: func(t *testing.T, vnodes []VNode) {
				assert := assert.New(t)
				assert.Equal([]string{
					"default-seed-client-0|392",
					"default-seed-client-1|40",
					"default-seed-client-0|359",
				}, strings(vnodes))
			},
		},
		{
			name:          "key 00000 with default vnodes per node",
			vnodesPerNode: DefaultVNodesPerNode,
			nodes:         mockSeedClients,
			key:           "00000",
			replicas:      2,
			expect: func(t *testing.T, vnodes []VNode) {
				assert := assert.New(t)
				assert.Equal([]string{
					"default-seed-client-2|115",
					"default-seed-client-2|286",
					"default-seed-client-1|355",
				}, strings(vnodes))
			},
		},
		{
			name:          "key dragonfly with default vnodes per node",
			vnodesPerNode: DefaultVNodesPerNode,
			nodes:         mockSeedClients,
			key:           "dragonfly",
			replicas:      2,
			expect: func(t *testing.T, vnodes []VNode) {
				assert := assert.New(t)
				assert.Equal([]string{
					"default-seed-client-1|397",
					"default-seed-client-0|422",
					"default-seed-client-2|153",
				}, strings(vnodes))
			},
		},
		{
			name:          "key dragonfly-seed-client-0 with default vnodes per node",
			vnodesPerNode: DefaultVNodesPerNode,
			nodes:         mockSeedClients,
			key:           "dragonfly-seed-client-0",
			replicas:      2,
			expect: func(t *testing.T, vnodes []VNode) {
				assert := assert.New(t)
				assert.Equal([]string{
					"default-seed-client-2|477",
					"default-seed-client-0|157",
					"default-seed-client-0|199",
				}, strings(vnodes))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ring := New(WithVNodesPerNode(tc.vnodesPerNode))
			for _, name := range tc.nodes {
				ring.Add(name)
			}

			vnode, ok := ring.Get(tc.key)
			assert.True(t, ok)

			vnodes := ring.GetWithReplicas(tc.key, tc.replicas)
			assert.Equal(t, vnode, vnodes[0])
			tc.expect(t, vnodes)
		})
	}
}
