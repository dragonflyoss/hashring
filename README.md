# Dragonfly Hashring

[![GitHub release](https://img.shields.io/github/release/dragonflyoss/hashring.svg)](https://github.com/dragonflyoss/hashring/releases)
[![Crates.io](https://img.shields.io/crates/v/dragonfly-hashring.svg)](https://crates.io/crates/dragonfly-hashring)
[![Go Reference](https://pkg.go.dev/badge/d7y.io/hashring.svg)](https://pkg.go.dev/d7y.io/hashring)
[![LICENSE](https://img.shields.io/github/license/dragonflyoss/dragonfly.svg?style=flat-square)](https://github.com/dragonflyoss/dragonfly/blob/main/LICENSE)

Consistent hash ring with virtual nodes shared by the Dragonfly scheduler, the Dragonfly client and the
Dragonfly SDKs. The Go module and the Rust crate place keys bit-for-bit identically, so every component
selects the same seed peer or scheduler for the same task.

## Packages

| Package                                                           | Description                     |
| :---------------------------------------------------------------- | :------------------------------ |
| [d7y.io/hashring](https://pkg.go.dev/d7y.io/hashring)             | Go module for scheduler and SDK |
| [dragonfly-hashring](https://crates.io/crates/dragonfly-hashring) | Rust crate for client and SDK   |

## Usage

Go:

```go
import "d7y.io/hashring"

ring := hashring.New()
ring.Add("default-seed-client-0")
ring.Add("default-seed-client-1")

vnode, _ := ring.Get(taskID)
name := vnode.Name()
```

Rust:

```rust
use dragonfly_hashring::VNodeHashRing;

let mut ring = VNodeHashRing::new();
ring.add("default-seed-client-0".to_string());
ring.add("default-seed-client-1".to_string());

let name = ring.get(&task_id).unwrap().name();
```

Both constructors use `DefaultVNodesPerNode` virtual nodes per node, pass `hashring.WithVNodesPerNode(n)` or call
`VNodeHashRing::with_vnodes_per_node(n)` to override it.

## Documentation

You can find the full documentation on the [d7y.io](https://d7y.io).

## Community

Join the conversation and help the community grow. Here are the ways to get involved:

- **Slack Channel**: [#dragonfly](https://cloud-native.slack.com/messages/dragonfly/) on [CNCF Slack](https://slack.cncf.io/)
- **Github Discussions**: [Dragonfly Discussion Forum](https://github.com/dragonflyoss/dragonfly/discussions)
- **Developer Group**: <dragonfly-developers@googlegroups.com>
- **Mailing Lists**:
  - **Developers**: <dragonfly-developers@googlegroups.com>
  - **Maintainers**: <dragonfly-maintainers@googlegroups.com>
- **Twitter**: [@dragonfly_oss](https://twitter.com/dragonfly_oss)

## Contributing

You should check out our [CONTRIBUTING](./CONTRIBUTING.md) and develop the project together.
