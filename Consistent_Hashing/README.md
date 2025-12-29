# Consistent Hashing

This repository documents my hands-on implementation and evaluation of consistent hashing, a fundamental system design technique used in distributed systems to achieve stable data partitioning under node churn.

Rather than relying on theory alone, I implemented the algorithm from scratch and measured its behavior empirically, including a direct comparison against naïve modulo hashing.

---

## What is Consistent Hashing?

Consistent hashing is a data partitioning strategy that minimizes key reassignment when the number of nodes in a system changes.

In distributed systems, keys (e.g., cache entries, database rows, requests) must be mapped to nodes. A naïve approach is:
```python
node = hash(key) % N
```

This works only when `N` (the number of nodes) never changes. As soon as a node is added or removed, most keys are reassigned, causing large-scale cache invalidation, data reshuffling, and performance degradation.

Consistent hashing solves this problem by ensuring that only a small, bounded fraction of keys move when the cluster topology changes.

---

## Core Idea

- Both keys and nodes are hashed into the same hash space.
- The hash space is treated as a ring.
- A key is assigned to the next node clockwise on the ring.
- When nodes are added or removed, only neighboring ranges are affected.

This preserves locality and prevents global reshuffling.

---

## Virtual Nodes

To improve load distribution, each physical node is represented by multiple virtual nodes on the ring:
```
A#0, A#1, A#2, ...
```

Virtual nodes:
- Reduce variance in key distribution
- Prevent load skew
- Make rebalancing smoother and more predictable

This implementation uses virtual nodes explicitly.

---

## What I Implemented

From scratch, without external libraries:

- Consistent hash ring with virtual nodes
- Deterministic node addition and removal
- O(log N) key lookup using binary search
- Proper wrap-around handling on the ring
- Experimental framework to measure key movement
- Side-by-side comparison with modulo hashing

---

## Experiment: Key Movement Measurement

### Setup

- **100,000 keys**
- **Initial nodes:** A, B, C, D
- **Virtual nodes per physical node:** 100
- **Operation:** add one node (E)

### Metric
```
Key Movement % = (keys whose assigned node changed) / (total keys)
```

### Results

#### Consistent Hashing
```
~22%
```

This matches the theoretical expectation:
```
Expected movement ≈ 1 / (N + 1)
For N = 4 → ~20%
```

#### Modulo Hashing
```
~80–90%
```

Almost all keys are reassigned when a node is added.

### Interpretation

- Consistent hashing bounds disruption and preserves stability.
- Modulo hashing causes catastrophic reshuffling under topology changes.

This empirical difference is the primary reason consistent hashing is used in:

- Distributed caches
- Sharded databases
- Load balancers
- Object stores
- Service discovery systems

---

## What Consistent Hashing Solves (and What It Doesn't)

### Solves

- Stable partitioning under node addition/removal
- Predictable, bounded key movement
- Scalable data distribution

### Does NOT solve

- Hot keys
- Replication
- Fault tolerance by itself
- Strong consistency
- Skewed access patterns

These require additional system-level mechanisms layered on top.

---

## Why This Exists

This project was built to move beyond theoretical understanding and validate system design behavior through implementation and measurement.

Instead of treating consistent hashing as a buzzword, this repository demonstrates:

- How it works
- Why it exists
- What breaks without it
- What guarantees it actually provides

---

## Possible Extensions

- Measure load skew per node
- Simulate hot keys
- Add replication (multiple successors)
- Compare behavior under node removal
- Make the ring thread-safe
- Build a sharded key-value store on top

---

## Takeaway

Consistent hashing is not an optimization, it is a **stability mechanism**.

Without it, elastic distributed systems do not scale safely.

This repository demonstrates that fact with real code and real numbers.