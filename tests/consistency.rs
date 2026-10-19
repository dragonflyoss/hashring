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

//! Cross-language consistency vectors shared with the Go module
//! (consistency_test.go). Both test suites assert the same fixed outputs, so a
//! change on either side that moves a key to another virtual node fails the
//! build.

use dragonfly_hashring::{VNodeHashRing, DEFAULT_VNODES_PER_NODE};
use siphasher::sip::SipHasher24;
use std::hash::Hasher;

const SEED_PEERS: [&str; 3] = ["seed-peer-1", "seed-peer-2", "seed-peer-3"];

const SEED_CLIENTS: [&str; 3] = [
    "default-seed-client-0",
    "default-seed-client-1",
    "default-seed-client-2",
];

fn vnode_hash(id: u64, name: &str) -> u64 {
    let mut hasher = SipHasher24::new();
    hasher.write(&id.to_le_bytes());
    hasher.write(name.as_bytes());
    hasher.write_u8(0xff);
    hasher.finish()
}

fn key_hash(key: &str) -> u64 {
    let mut hasher = SipHasher24::new();
    hasher.write(key.as_bytes());
    hasher.write_u8(0xff);
    hasher.finish()
}

#[test]
fn siphash_matches_go() {
    let test_cases = vec![
        (0, "seed-peer-1", 0x1e5a582b8d945969),
        (1, "seed-peer-1", 0xb5db98265419376c),
        (0, "seed-peer-2", 0xf36f748c486b09ef),
        (511, "seed-peer-3", 0xf0eda07426c5cac1),
    ];

    for (id, name, expected) in test_cases {
        assert_eq!(vnode_hash(id, name), expected, "id: {id}, name: {name}");
    }

    let test_cases = vec![
        (
            "b2c366cce7e68013d5441c6326d5a3e1b12aeb5ed58564d0fd3fa089bc29cb6e",
            0xaa86f08bc4878a68,
        ),
        ("test-task", 0xa78dd821b4023714),
        ("", 0x20dd4eb33d9590f2),
    ];

    for (key, expected) in test_cases {
        assert_eq!(key_hash(key), expected, "key: {key}");
    }
}

#[test]
fn hashring_matches_go() {
    let test_cases = vec![
        (
            3,
            SEED_PEERS,
            "task-a",
            2,
            vec!["seed-peer-3|2", "seed-peer-1|2", "seed-peer-3|0"],
        ),
        (
            3,
            SEED_PEERS,
            "task-b",
            2,
            vec!["seed-peer-2|0", "seed-peer-2|1", "seed-peer-1|0"],
        ),
        (
            3,
            SEED_PEERS,
            "task-c",
            2,
            vec!["seed-peer-2|1", "seed-peer-1|0", "seed-peer-3|2"],
        ),
        (
            3,
            SEED_PEERS,
            "task-a",
            100,
            vec![
                "seed-peer-3|2",
                "seed-peer-1|2",
                "seed-peer-3|0",
                "seed-peer-1|1",
                "seed-peer-2|2",
                "seed-peer-3|1",
                "seed-peer-2|0",
                "seed-peer-2|1",
                "seed-peer-1|0",
            ],
        ),
        (
            DEFAULT_VNODES_PER_NODE,
            SEED_CLIENTS,
            "0000000000000000000000000000000000000000000000000000000000000000",
            2,
            vec![
                "default-seed-client-1|260",
                "default-seed-client-0|383",
                "default-seed-client-0|281",
            ],
        ),
        (
            DEFAULT_VNODES_PER_NODE,
            SEED_CLIENTS,
            "0000000000000000000000000000000000000000000000000000000000001eef",
            2,
            vec![
                "default-seed-client-1|375",
                "default-seed-client-2|109",
                "default-seed-client-0|483",
            ],
        ),
        (
            DEFAULT_VNODES_PER_NODE,
            SEED_CLIENTS,
            "0000000000000000000000000000000000000000000000000000000000003dde",
            2,
            vec![
                "default-seed-client-2|366",
                "default-seed-client-0|374",
                "default-seed-client-1|30",
            ],
        ),
        (
            DEFAULT_VNODES_PER_NODE,
            SEED_CLIENTS,
            "0000000000000000000000000000000000000000000000000000000000005ccd",
            2,
            vec![
                "default-seed-client-0|69",
                "default-seed-client-0|121",
                "default-seed-client-2|243",
            ],
        ),
        (
            DEFAULT_VNODES_PER_NODE,
            SEED_CLIENTS,
            "0000000000000000000000000000000000000000000000000000000000007bbc",
            2,
            vec![
                "default-seed-client-0|239",
                "default-seed-client-1|156",
                "default-seed-client-1|157",
            ],
        ),
        (
            DEFAULT_VNODES_PER_NODE,
            SEED_CLIENTS,
            "0000000000000000000000000000000000000000000000000000000000009aab",
            2,
            vec![
                "default-seed-client-1|501",
                "default-seed-client-0|435",
                "default-seed-client-1|308",
            ],
        ),
        (
            DEFAULT_VNODES_PER_NODE,
            SEED_CLIENTS,
            "000000000000000000000000000000000000000000000000000000000000b99a",
            2,
            vec![
                "default-seed-client-0|223",
                "default-seed-client-0|9",
                "default-seed-client-2|83",
            ],
        ),
        (
            DEFAULT_VNODES_PER_NODE,
            SEED_CLIENTS,
            "000000000000000000000000000000000000000000000000000000000000d889",
            2,
            vec![
                "default-seed-client-1|429",
                "default-seed-client-2|220",
                "default-seed-client-1|220",
            ],
        ),
        (
            DEFAULT_VNODES_PER_NODE,
            SEED_CLIENTS,
            "000000000000000000000000000000000000000000000000000000000000f778",
            2,
            vec![
                "default-seed-client-2|326",
                "default-seed-client-1|155",
                "default-seed-client-0|473",
            ],
        ),
        (
            DEFAULT_VNODES_PER_NODE,
            SEED_CLIENTS,
            "0000000000000000000000000000000000000000000000000000000000011667",
            2,
            vec![
                "default-seed-client-0|293",
                "default-seed-client-1|334",
                "default-seed-client-2|199",
            ],
        ),
        (
            DEFAULT_VNODES_PER_NODE,
            SEED_CLIENTS,
            "0000000000000000000000000000000000000000000000000000000000013556",
            2,
            vec![
                "default-seed-client-2|246",
                "default-seed-client-2|65",
                "default-seed-client-0|369",
            ],
        ),
        (
            DEFAULT_VNODES_PER_NODE,
            SEED_CLIENTS,
            "0000000000000000000000000000000000000000000000000000000000015445",
            2,
            vec![
                "default-seed-client-2|480",
                "default-seed-client-1|24",
                "default-seed-client-1|147",
            ],
        ),
        (
            DEFAULT_VNODES_PER_NODE,
            SEED_CLIENTS,
            "",
            2,
            vec![
                "default-seed-client-1|163",
                "default-seed-client-2|24",
                "default-seed-client-1|252",
            ],
        ),
        (
            DEFAULT_VNODES_PER_NODE,
            SEED_CLIENTS,
            "a",
            2,
            vec![
                "default-seed-client-0|392",
                "default-seed-client-1|40",
                "default-seed-client-0|359",
            ],
        ),
        (
            DEFAULT_VNODES_PER_NODE,
            SEED_CLIENTS,
            "00000",
            2,
            vec![
                "default-seed-client-2|115",
                "default-seed-client-2|286",
                "default-seed-client-1|355",
            ],
        ),
        (
            DEFAULT_VNODES_PER_NODE,
            SEED_CLIENTS,
            "dragonfly",
            2,
            vec![
                "default-seed-client-1|397",
                "default-seed-client-0|422",
                "default-seed-client-2|153",
            ],
        ),
        (
            DEFAULT_VNODES_PER_NODE,
            SEED_CLIENTS,
            "dragonfly-seed-client-0",
            2,
            vec![
                "default-seed-client-2|477",
                "default-seed-client-0|157",
                "default-seed-client-0|199",
            ],
        ),
    ];

    for (vnodes_per_node, nodes, key, replicas, expected) in test_cases {
        let mut ring = VNodeHashRing::with_vnodes_per_node(vnodes_per_node);
        for name in nodes {
            ring.add(name.to_string());
        }

        let vnodes: Vec<String> = ring
            .get_with_replicas(key, replicas)
            .unwrap()
            .iter()
            .map(|vnode| vnode.to_string())
            .collect();
        assert_eq!(ring.get(key).unwrap().to_string(), vnodes[0], "key: {key}");
        assert_eq!(vnodes, expected, "key: {key}, replicas: {replicas}");
    }
}
