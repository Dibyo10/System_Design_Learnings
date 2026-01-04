# CONCEPT: CACHING

## Why This Matters

Most systems fail not because they can't handle load, but because they keep doing expensive work over and over. Caching exists to solve one fundamental problem: **doing the same expensive operation repeatedly is wasteful**.

This guide explains caching through real-world analogies and practical examples. By the end, you'll understand not just *how* caching works, but *when* and *why* different strategies matter.

---

## Understanding the Fundamentals

### The Restaurant Analogy

Think of a restaurant with:
- A **chef** (your database)
- A **menu board** near the entrance (your cache)
- **Customers** continuously walking in (incoming requests)

If every customer had to walk into the kitchen, wait for the chef to prepare their meal from scratch, and then leave, the restaurant would collapse under pressure.

Instead, the restaurant does something smart:
- Frequently ordered items are prepared in advance
- They're kept closer to customers
- When ingredients change, the menu board gets updated

That menu board is your cache.

### What Makes a Cache a Cache?

A cache has three defining characteristics:

1. **Smaller than the source** - It can't store everything
2. **Faster than the source** - That's the whole point
3. **Closer to the requester** - Reduced network/access latency

But here's the critical part: **the cache is not the source of truth**. It's a temporary memory that trades correctness guarantees for speed. This trade-off is the heart of every caching decision you'll make.

---

## Cache Reading Strategies

When data is requested, how does your system decide whether to use the cache? This decision defines your read strategy.

### Cache-Aside (Lazy Loading)

**The Pattern:**
```
1. Application checks cache
2. If found (cache hit) → return immediately
3. If not found (cache miss) → fetch from database
4. Store result in cache
5. Return to user
```

**Restaurant Analogy:**
You ask the waiter: "Do you already have this dish ready?"
- If yes → you get it immediately
- If no → waiter goes to the kitchen, gets it, and remembers it for the next customer

**Why It's Popular:**
- Simple to implement
- Application stays in full control
- If cache fails, application still works (degraded but functional)

**The Hidden Danger:**
If 100 users simultaneously request the same uncached data, all 100 requests hit the database. This is called a **cache stampede** or **thundering herd problem**.

**When to Use:**
- Most general-purpose applications
- When you need fault tolerance (cache failure shouldn't break the app)
- When access patterns are unpredictable

```
┌──────────┐
│  Client  │
└────┬─────┘
     │ 1. Check cache
     ▼
┌─────────────┐    Hit?    ┌──────────┐
│    Cache    │◄───────────►│   App    │
└─────────────┘             └────┬─────┘
                                 │ 2. If miss
                                 ▼
                            ┌──────────┐
                            │ Database │
                            └──────────┘
                                 │
                                 │ 3. Store in cache
                                 ▼
                            ┌─────────────┐
                            │    Cache    │
                            └─────────────┘
```

---

### Read-Through Cache

**The Pattern:**
```
1. Application asks cache
2. Cache automatically fetches from database on miss
3. Cache returns data to application
```

**Restaurant Analogy:**
You don't ask the waiter. You order from a vending machine. If the vending machine doesn't have your item, *it* automatically restocks itself.

**Why Teams Like It:**
- Cleaner application code (no cache logic scattered everywhere)
- Centralized caching behavior
- Consistent cache population strategy

**Why Senior Engineers Are Cautious:**
- Cache becomes a critical dependency (if it fails, reads fail)
- Debugging becomes harder (logic hidden in cache layer)
- Cache failures now block all reads

**When to Use:**
- When you want simplified application logic
- When cache infrastructure is highly reliable
- When you have dedicated cache administrators

```
┌──────────┐
│  Client  │
└────┬─────┘
     │
     ▼
┌─────────────┐
│    Cache    │ (handles fetch internally)
│             │
│   ┌─────┐   │
│   │ DB  │   │
│   └─────┘   │
└─────────────┘
```

---

### Refresh-Ahead Cache

**The Pattern:**
```
- Cache monitors TTL (time-to-live)
- Before expiration, cache proactively refreshes data
- Users never experience cache misses for hot data
```

**Restaurant Analogy:**
The restaurant notices biryani sells out every 15 minutes. Before it runs out, the chef starts cooking the next batch preemptively.

**Why It Exists:**
- Eliminates cache misses for frequently accessed data
- Reduces tail latency (no user waits for a refresh)
- Keeps hot data always warm

**Why It's Dangerous:**
- If predictions are wrong, you hammer the database unnecessarily
- Rapidly changing data means you refresh based on stale assumptions
- Complex to tune correctly

**When to Use:**
- Predictable access patterns (e.g., homepage data, trending items)
- When you need consistent low latency
- When database can handle predictable refresh load

```
┌─────────────┐
│    Cache    │
│             │
│  TTL: 50%   │──────► Proactive refresh triggered
│             │
│  ┌──────┐   │
│  │Timer │   │
│  └──────┘   │
└──────┬──────┘
       │
       ▼
  ┌──────────┐
  │ Database │
  └──────────┘
```

---

## Cache Writing Strategies

Reads are straightforward. Writes are where systems break. Every write strategy answers one question: **When does the database know about changes?**

### Write-Through Cache

**The Pattern:**
```
1. Write goes to cache
2. Write goes to database synchronously (before returning success)
3. Both must succeed for operation to succeed
```

**Restaurant Analogy:**
You update the menu board **and** tell the chef immediately. Both happen before you confirm to the customer.

**What You Get:**
- Strong consistency (reads always see latest writes)
- No data loss (database is always up-to-date)
- Simple reasoning about data correctness

**What You Pay:**
- Slower writes (must wait for database write)
- Cache fills with write-only data that might never be read

**When to Use:**
- Financial transactions
- User authentication data
- Any data where correctness > performance

```
┌─────────┐
│  Write  │
└────┬────┘
     │
     ▼
┌─────────────┐
│    Cache    │ (write)
└──────┬──────┘
       │
       │ (synchronous)
       ▼
  ┌──────────┐
  │ Database │ (write)
  └──────────┘
       │
       └──────► Success returned only after both complete
```

---

### Write-Around Cache

**The Pattern:**
```
1. Write goes directly to database
2. Cache is NOT updated immediately
3. Cache gets updated later on next read (or invalidated)
```

**Restaurant Analogy:**
You tell the chef directly. The menu board updates only when someone asks for that item.

**Why It Exists:**
- Prevents cache pollution (data written but never read)
- Better for write-heavy workloads
- Reduces cache storage pressure

**The Hidden Danger:**
If you forget to invalidate the existing cache entry, users read stale data. Write-around systems often fail silently.

**When to Use:**
- Write-heavy applications with infrequent reads
- Bulk import operations
- Logging or analytics data

```
┌─────────┐
│  Write  │
└────┬────┘
     │
     │ (bypass cache)
     ▼
┌──────────┐
│ Database │
└──────────┘
     │
     │ Cache unaware until:
     ├──► Invalidation event
     └──► Next read (cache miss)
```

---

### Write-Back (Write-Behind) Cache

**The Pattern:**
```
1. Write goes to cache (returns immediately)
2. Cache asynchronously writes to database later
3. Application doesn't wait for database write
```

**Restaurant Analogy:**
You update the menu board. Someone will tell the chef... eventually.

**Why Teams Love It:**
- Extremely fast writes
- Absorbs traffic spikes
- Can batch multiple writes efficiently

**Why Engineers Fear It:**
- Cache crash = data loss (unwritten changes lost)
- Requires sophisticated replication and durability
- Complex failure recovery scenarios

**When to Use:**
- Non-critical data (view counts, like counts)
- Metrics and analytics
- High-throughput logging
- **Never for financial or user-critical data**

```
┌─────────┐
│  Write  │
└────┬────┘
     │
     ▼
┌─────────────┐
│    Cache    │ (write + return immediately)
└──────┬──────┘
       │
       │ (async, batched)
       ▼
  ┌──────────┐
  │ Database │ (eventual write)
  └──────────┘
```

⚠️ **Critical Warning:** Write-back caching trades durability for performance. Only use when you can tolerate data loss.

---

## Cache Eviction Policies

Caches are finite. When memory fills up, something must be removed. Your eviction policy answers: **"What do we sacrifice?"**

### LRU (Least Recently Used)

**Logic:** Remove items that haven't been accessed recently.

**Restaurant Analogy:** "If people ordered this recently, they'll probably order it again soon."

**Strengths:**
- Works well for typical user traffic patterns
- Simple to implement
- Good default choice

**Weaknesses:**
- Large scans can evict important data
- Batch jobs pollute the cache
- Sequential access patterns perform poorly

**When to Use:** Most general-purpose applications with typical access patterns.

---

### LFU (Least Frequently Used)

**Logic:** Remove items accessed least often over time.

**Restaurant Analogy:** "This dish is ordered all the time. Keep it no matter what."

**Strengths:**
- Great for stable, predictable workloads
- Protects genuinely popular data

**Weaknesses:**
- Old popular items never die (even when no longer relevant)
- Slow to adapt to changing trends
- Requires frequency tracking (more memory overhead)

**When to Use:** Stable access patterns where popularity is consistent (e.g., reference data, configuration).

---

### FIFO / LIFO / MRU / MFU

**Why They Exist:** Primarily for teaching and specific edge cases.

These policies ignore access behavior and work purely on insertion order or inverse logic. They're rarely used in production except for very specific workload patterns.

**When to Use:** Don't, unless you have a very specific reason.

---

## Eviction vs Invalidation (Critical Distinction)

Many engineers confuse these. They're completely different:

| Aspect | Eviction | Invalidation |
|--------|----------|--------------|
| **Trigger** | Memory pressure | Data changed |
| **Based on** | Usage patterns | Correctness |
| **Purpose** | Free space | Maintain consistency |
| **Timing** | When cache is full | When source data updates |

**Key Insight:** A system with perfect eviction but broken invalidation is **wrong**, not just slow. Correctness always trumps performance.

---

## Cache Invalidation Strategies

Phil Karlton famously said: *"There are only two hard things in Computer Science: cache invalidation and naming things."*

The problem: Your database changes. Your cache doesn't magically know that.

### TTL-Based Invalidation

**How It Works:**
Every cache entry has an expiration time. After that time, it's considered invalid (regardless of whether data actually changed).

**Restaurant Analogy:** The menu board updates every 10 minutes, whether ingredients changed or not.

**Strengths:**
- Simple to implement
- No coordination needed
- Guaranteed maximum staleness

**Weaknesses:**
- Always serves stale data until expiry
- Hard to pick correct TTL (too short = database load, too long = stale data)
- Wastes resources refreshing unchanged data

**When to Use:**
- When you can tolerate bounded staleness
- When changes are unpredictable
- As a safety net for other strategies

```
Cache Entry:
┌──────────────────────┐
│ Key: user:123        │
│ Value: {...}         │
│ TTL: 300 seconds     │  ──► Expires at: 14:35:00
│ Created: 14:30:00    │
└──────────────────────┘

Timeline:
14:30:00 ──► Cache populated
14:32:00 ──► DB changes (cache still serves old data)
14:35:00 ──► TTL expires, next read fetches fresh data
```

**TTL is not about performance. TTL is about bounding incorrectness.**

---

### Event-Based Invalidation

**How It Works:**
When data changes in the database, an event is triggered that invalidates or updates the corresponding cache entry.

**Restaurant Analogy:** Every time the chef changes a dish, someone immediately updates the menu board.

**Strengths:**
- Minimal staleness
- Cache always reflects recent changes
- No unnecessary refreshes

**Weaknesses:**
- Events can be lost in distributed systems
- Partial failures cause inconsistency
- Complex to implement correctly
- Tight coupling between database and cache

**When to Use:**
- When strong consistency is required
- When you have reliable event infrastructure
- When change frequency is moderate

```
┌──────────┐
│ Database │
└────┬─────┘
     │ 1. Write occurs
     │
     ▼
┌──────────────┐
│ Event Queue  │
└──────┬───────┘
       │ 2. Invalidation event
       │
       ▼
┌─────────────┐
│    Cache    │ 3. Remove/update entry
└─────────────┘
```

---

### Hybrid Invalidation (The Production Reality)

**How It Works:**
Event-based invalidation + TTL as safety net.

**Why This Exists:**
Real systems assume events will fail. TTL ensures staleness is bounded even when invalidation events are lost.

**The Pattern:**
```
- Primary: Event-based invalidation (best effort)
- Backup: TTL expiration (guaranteed maximum staleness)
```

**When to Use:** Almost always in production systems.

This is how robust systems survive network partitions, event delivery failures, and other real-world problems.

```
┌──────────┐
│ Database │ ──► Event ──► Cache invalidation
└──────────┘                (primary mechanism)
                                  │
                                  │ If event fails
                                  ▼
                            TTL expires ──► Cache refresh
                            (safety net)
```

---

## Real-World Example: User Profile Service

Let's apply everything we've learned to a concrete example.

**Requirements:**
- User profiles are read frequently (10:1 read-to-write ratio)
- Profile updates must be immediately visible
- Profile data is ~5KB per user
- 1M active users, peak 100K requests/second

**Design Decisions:**

| Aspect | Choice | Reasoning |
|--------|--------|-----------|
| **Read Strategy** | Cache-Aside | Fault tolerance important; if cache fails, app still works |
| **Write Strategy** | Write-Through | Consistency critical; users must see their updates immediately |
| **Eviction Policy** | LRU | Typical access patterns; recently accessed profiles likely accessed again |
| **Invalidation** | Hybrid (Event + TTL) | Events for normal case, TTL=300s for safety |
| **Cache Size** | 100K entries | Fits most active users (500MB total), evicts inactive ones |

**What Happens:**

1. **Read Path:**
   - Check cache → 90% hit rate
   - On miss: fetch from DB, populate cache, return
   - Latency: 5ms (cache hit) vs 50ms (cache miss)

2. **Write Path:**
   - Update cache
   - Update database synchronously
   - Emit invalidation event (other cache instances)
   - Return success

3. **Failure Scenarios:**
   - Cache dies: App degrades to 50ms reads, but stays up
   - Event lost: TTL ensures stale data expires within 5 minutes
   - Database slow: Write-through means writes are slower, but consistency maintained

**Every choice answers a failure question.**

---

## When NOT to Use Caching

Caching isn't always the answer. Avoid caching when:

### 1. Data Changes Frequently and Consistency is Critical

**Example:** Stock prices, bank account balances, inventory counts in high-frequency trading.

**Why:** The invalidation overhead exceeds caching benefits. You'll spend more time invalidating than serving cached data.

**Alternative:** Optimize database queries, use read replicas, or accept reading from source.

---

### 2. Data is Accessed Only Once (No Reuse)

**Example:** One-time report generation, unique user-specific computations.

**Why:** Caches benefit from repeated access. If every request is unique, you're just adding latency and complexity.

**Alternative:** Scale your compute/database layer directly.

---

### 3. Data Size Exceeds Cache Capacity by Orders of Magnitude

**Example:** Multi-terabyte datasets where only megabytes can be cached.

**Why:** Low hit rates mean you're mostly experiencing cache misses anyway. The overhead of cache checking isn't worth it.

**Alternative:** Use database query optimization, partitioning, or specialized data stores.

---

### 4. Correctness Requirements Prohibit Staleness

**Example:** Medical records, legal documents, financial transactions during processing.

**Why:** Even brief staleness is unacceptable. The risk of serving incorrect data outweighs performance gains.

**Alternative:** Use strongly consistent databases, synchronous replication, or distributed transactions.

---

### 5. Your System is Too Simple to Justify Complexity

**Example:** Internal tool with 10 users, CRUD app with minimal traffic.

**Why:** Caching adds operational complexity, monitoring requirements, and failure modes. If your database handles load fine, don't cache.

**Alternative:** Keep it simple. Optimize when you have actual performance problems.

---

### 6. Debugging and Observability Become Impossible

**Example:** When cache misses cascade into complex failure modes you can't trace.

**Why:** If you can't debug your system, you can't fix it. Complexity has a cost.

**Alternative:** Start without cache. Add it when benefits clearly outweigh complexity cost.

---

## The Mental Model

Caching is not fundamentally about speed. **Caching is about deciding where you tolerate inconsistency.**

Every caching decision answers three questions:

1. **What can go stale?**
2. **For how long?**
3. **What happens when it fails?**

If you can answer these three questions confidently, you understand caching.

---

## Key Takeaways

✅ **Caches trade correctness for performance** - Know what you're trading  
✅ **Choose strategies based on your failure scenarios** - Not what sounds clever  
✅ **Eviction ≠ Invalidation** - One is about space, the other about correctness  
✅ **TTL is about bounding incorrectness** - Not about performance optimization  
✅ **Event-based invalidation + TTL safety net** - How production systems survive  
✅ **Write-back caching can lose data** - Only use for non-critical data  
✅ **Cache-aside is fault-tolerant** - Cache failure doesn't break your app  
✅ **Not every problem needs caching** - Sometimes your database is fine  

---

## Further Learning

- **Deep dive:** Distributed caching, cache coherence protocols
- **Practice:** Design a caching strategy for an e-commerce product catalog
- **Read:** "Caching at Scale" talks from major tech companies
- **Explore:** Redis, Memcached, CDN caching strategies

---

*This is part of my system design learning journey. For questions or discussions, feel free to open an issue.*