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

use siphasher::sip::SipHasher24;
use std::fmt;
use std::hash::Hasher;

/// The default number of virtual nodes of each node on the ring, shared by the
/// seed peer selectors of dragonfly and dragonfly-sdk.
pub const DEFAULT_VNODES_PER_NODE: usize = 512;

/// A virtual node on the consistent hash ring. Each physical node is
/// represented by multiple virtual nodes to better balance key distribution.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct VNode {
    /// The replica index of this virtual node for its physical node.
    id: usize,

    /// The physical node name this virtual node represents.
    name: String,
}

/// Implements the virtual node.
impl VNode {
    /// Creates a new virtual node with the given replica id and physical name.
    fn new(id: usize, name: String) -> Self {
        VNode { id, name }
    }

    /// Returns the physical node name associated with this virtual node.
    pub fn name(&self) -> &str {
        &self.name
    }

    /// Returns the position of the virtual node on the ring.
    ///
    /// The encoding is the Rust std `Hash` of the first implementation of the
    /// virtual node: the replica id as 8 little-endian bytes, the name bytes
    /// and the 0xff terminator of a string, hashed with SipHash-2-4 and zero
    /// keys. Changing it moves every key to another node, so it is a wire
    /// compatibility contract shared with the Go module.
    fn hash(&self) -> u64 {
        let mut hasher = SipHasher24::new();
        hasher.write(&(self.id as u64).to_le_bytes());
        hasher.write(self.name.as_bytes());
        hasher.write_u8(0xff);
        hasher.finish()
    }
}

/// Implements the Display trait to format the virtual node as "name|id".
impl fmt::Display for VNode {
    /// Formats the virtual node.
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}|{}", self.name, self.id)
    }
}

/// A virtual node with its position on the ring.
#[derive(Debug, Clone)]
struct Node {
    /// The position of the virtual node on the ring.
    key: u64,

    /// The virtual node at this position.
    vnode: VNode,
}

/// A consistent hash ring that uses virtual nodes to improve key distribution.
/// When a physical node is added, `vnodes_per_node` virtual nodes are inserted
/// into the ring.
#[derive(Debug, Clone)]
pub struct VNodeHashRing {
    /// The number of virtual nodes of each physical node.
    vnodes_per_node: usize,

    /// The ring sorted by key.
    nodes: Vec<Node>,
}

/// Implements the Default trait with [DEFAULT_VNODES_PER_NODE] virtual nodes per node.
impl Default for VNodeHashRing {
    /// Creates a new hash ring with the default virtual nodes per node.
    fn default() -> Self {
        Self::with_vnodes_per_node(DEFAULT_VNODES_PER_NODE)
    }
}

/// Implements the consistent hash ring.
impl VNodeHashRing {
    /// Creates a new hash ring with [DEFAULT_VNODES_PER_NODE] virtual nodes per node.
    pub fn new() -> Self {
        Self::default()
    }

    /// Creates a new hash ring with `vnodes_per_node` virtual nodes per node.
    pub fn with_vnodes_per_node(vnodes_per_node: usize) -> Self {
        VNodeHashRing {
            vnodes_per_node,
            nodes: Vec::new(),
        }
    }

    /// Adds `vnodes_per_node` virtual nodes of the named node to the ring.
    pub fn add(&mut self, name: String) {
        for id in 0..self.vnodes_per_node {
            let vnode = VNode::new(id, name.clone());
            self.nodes.push(Node {
                key: vnode.hash(),
                vnode,
            });
        }

        self.nodes.sort_by_key(|node| node.key);
    }

    /// Returns the virtual node responsible for the key, and `None` when the
    /// ring is empty.
    pub fn get(&self, key: &str) -> Option<&VNode> {
        if self.nodes.is_empty() {
            return None;
        }

        Some(&self.nodes[self.search(hash_key(key))].vnode)
    }

    /// Returns the virtual node responsible for the key along with the next
    /// `replicas` virtual nodes clockwise on the ring, and `None` when the ring
    /// is empty. When `replicas` is larger than the length of the ring, the
    /// result shrinks to the entire ring.
    pub fn get_with_replicas(&self, key: &str, replicas: usize) -> Option<Vec<VNode>> {
        if self.nodes.is_empty() {
            return None;
        }

        let take = if replicas > self.nodes.len() {
            self.nodes.len()
        } else {
            replicas + 1
        };

        let start = self.search(hash_key(key));
        Some(
            self.nodes
                .iter()
                .cycle()
                .skip(start)
                .take(take)
                .map(|node| node.vnode.clone())
                .collect(),
        )
    }

    /// Returns the number of virtual nodes on the ring.
    pub fn len(&self) -> usize {
        self.nodes.len()
    }

    /// Returns whether the ring has no virtual nodes.
    pub fn is_empty(&self) -> bool {
        self.nodes.is_empty()
    }

    /// Returns the index of the first node whose key is not less than the key,
    /// wrapping to the first node past the end of the ring.
    fn search(&self, key: u64) -> usize {
        self.nodes.partition_point(|node| node.key < key) % self.nodes.len()
    }
}

/// Returns the position of the key on the ring, hashed with the Rust std
/// `Hash` encoding of a string.
fn hash_key(key: &str) -> u64 {
    let mut hasher = SipHasher24::new();
    hasher.write(key.as_bytes());
    hasher.write_u8(0xff);
    hasher.finish()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn vnode_formats_as_name_and_id() {
        let vnode = VNode::new(1, "default-seed-client-1".to_string());
        assert_eq!(vnode.name(), "default-seed-client-1");
        assert_eq!(vnode.to_string(), "default-seed-client-1|1");
    }

    #[test]
    fn add_inserts_vnodes_per_node_for_each_node() {
        let test_cases = vec![
            (3, vec![], 0),
            (2, vec!["default-seed-client-1"], 2),
            (2, vec!["default-seed-client-1", "default-seed-client-2"], 4),
        ];

        for (vnodes_per_node, nodes, expected) in test_cases {
            let mut ring = VNodeHashRing::with_vnodes_per_node(vnodes_per_node);
            for name in nodes {
                ring.add(name.to_string());
            }

            assert_eq!(ring.len(), expected);
            assert_eq!(ring.is_empty(), expected == 0);
            assert!(ring.nodes.is_sorted_by_key(|node| node.key));
        }
    }

    #[test]
    fn new_uses_default_vnodes_per_node() {
        let mut ring = VNodeHashRing::new();
        ring.add("default-seed-client-1".to_string());
        assert_eq!(ring.len(), DEFAULT_VNODES_PER_NODE);
        assert_eq!(ring.len(), VNodeHashRing::default().vnodes_per_node);
    }

    #[test]
    fn get_returns_none_for_empty_ring() {
        let ring = VNodeHashRing::with_vnodes_per_node(2);
        assert!(ring.get("task").is_none());
        assert!(ring.get_with_replicas("task", 2).is_none());
    }

    #[test]
    fn get_maps_every_key_to_the_single_node() {
        let mut ring = VNodeHashRing::with_vnodes_per_node(2);
        ring.add("default-seed-client-1".to_string());

        for key in ["", "task-a", "task-b"] {
            assert_eq!(ring.get(key).unwrap().name(), "default-seed-client-1");
        }
    }

    #[test]
    fn get_with_replicas_clamps_to_the_ring() {
        let mut ring = VNodeHashRing::with_vnodes_per_node(2);
        ring.add("default-seed-client-1".to_string());
        ring.add("default-seed-client-2".to_string());

        let test_cases = vec![(2, 3), (4, 5), (100, 4)];
        for (replicas, expected) in test_cases {
            assert_eq!(
                ring.get_with_replicas("task", replicas).unwrap().len(),
                expected
            );
        }
    }

    #[test]
    fn add_order_does_not_affect_get() {
        let mut ring_a = VNodeHashRing::new();
        for name in [
            "default-seed-client-1",
            "default-seed-client-2",
            "default-seed-client-3",
        ] {
            ring_a.add(name.to_string());
        }

        let mut ring_b = VNodeHashRing::new();
        for name in [
            "default-seed-client-3",
            "default-seed-client-1",
            "default-seed-client-2",
        ] {
            ring_b.add(name.to_string());
        }

        for key in ["task-a", "task-b", "task-c", "task-d", "task-e"] {
            assert_eq!(ring_a.get(key), ring_b.get(key));
        }
    }
}
