# Rate Limiting – From First Principles

## Overview

When I first heard the term "rate limiting," it sounded simple: "Limit how many requests are allowed." But once I tried building real systems, I realized that rate limiting is not one problem—it's a family of problems, each with different trade-offs.

This guide covers:
- Understanding rate limiting from scratch
- Implementing the most common algorithms end-to-end
- Testing them under concurrency
- Observing behavior over time
- Building intuition, not just code

---

## The Core Problem

Any real system eventually faces this question: **"What should I do if too many requests arrive at once?"**

### Common Scenarios

- Too many API requests from a single user
- Sudden traffic spikes
- Protecting downstream services (databases, third-party APIs, LLMs)
- Preventing abuse (login attempts, password resets)

### The Challenge

- Allow legitimate usage
- Prevent overload
- Do this fairly, predictably, and efficiently

Different rate limiting algorithms answer different versions of this problem.

---

## Mental Model (Read This First)

All rate limiters answer one specific question:

| Algorithm | Core Question |
|-----------|--------------|
| **Token Bucket** | How many requests can I allow right now? |
| **Leaky Bucket** | How fast should requests leave my system? |
| **Sliding Window Log** | How many requests happened recently? |
| **Sliding Window Counter** | Approximately how many requests happened recently? |

*If you mix these questions up, the algorithm will feel confusing.*

---

## Algorithm Implementations

All implementations in this repository:
- Use monotonic time
- Are concurrency-safe
- Use lazy updates (no background goroutines)
- Include stress tests to observe real behavior

---

### 1. Token Bucket

#### Problem It Solves
"Allow bursts, but control the average rate."

#### Ideal For
- User-facing APIs
- Interactive systems
- Anything where low latency matters

#### How It Works

Imagine a bucket that holds tokens:
1. Each request needs 1 token
2. Tokens refill at a fixed rate
3. Bucket has a maximum capacity
4. If tokens are available → **allow**
5. If tokens are empty → **reject**

Unused tokens accumulate, which allows bursts.

#### Key Properties
- ✅ Burst-friendly
- ✅ Low latency
- ✅ Average rate is enforced, not instant rate
- ✅ Very common in API gateways

#### When It Breaks
- Downstream systems cannot handle bursts
- Strict smoothing is required

---

### 2. Leaky Bucket

#### Problem It Solves
"Protect downstream systems by enforcing a smooth output rate."

#### Ideal For
- External APIs (payments, third-party services)
- Databases with strict QPS limits
- Infrastructure protection

#### How It Works

Imagine a queue with a hole at the bottom:
1. Requests enter the queue
2. Queue has a fixed capacity
3. Requests "leak" out at a constant rate
4. If the queue is full → **reject immediately**
5. If not → **enqueue**

> **Important Clarification:** Leaky bucket does NOT smooth admission—it smooths output. Requests can arrive in a burst, but they are released downstream slowly.

#### Key Properties
- ✅ No bursts reach downstream
- ⚠️ Adds latency under load
- ✅ Predictable output rate
- ⚠️ System-friendly, not user-friendly

---

### 3. Sliding Window Log

#### Problem It Solves
"Enforce a strict limit over a moving time window."

#### Ideal For
- Login attempts
- Security-sensitive endpoints
- Abuse prevention

#### How It Works

Keep a list of timestamps for each request:
1. For a new request, remove timestamps older than the window
2. Count remaining timestamps
3. If count < limit → **allow**
4. Else → **reject**

#### Key Properties
- ✅ Perfectly accurate
- ✅ Strict enforcement
- ⚠️ Memory grows with traffic
- ⚠️ Requires cleanup

#### When It Breaks
- High traffic systems
- Long windows
- Memory-constrained environments

---

### 4. Sliding Window Counter

#### Problem It Solves
"Approximate sliding window enforcement with constant memory."

#### Ideal For
- High throughput systems
- When exact precision is not required

#### How It Works

Instead of storing timestamps:
1. Track count for current window
2. Track count for previous window
3. Compute a weighted average based on overlap

This approximates a sliding window without storing logs.

#### Key Properties
- ✅ O(1) memory
- ✅ Very fast
- ⚠️ Slightly inaccurate near window boundaries
- ✅ Widely used in production

---

## Algorithm Comparison Table

| Scenario / Requirement | Token Bucket | Leaky Bucket | Sliding Window Log | Sliding Window Counter |
|------------------------|--------------|--------------|-------------------|----------------------|
| **User-facing APIs (REST, GraphQL)** | ✅ Yes | ❌ No | ⚠️ Sometimes | ⚠️ Sometimes |
| **Allow short bursts** | ✅ Yes | ❌ No | ❌ No | ❌ No |
| **Smooth, constant output rate** | ❌ No | ✅ Yes | ❌ No | ❌ No |
| **Protect fragile downstream systems** | ⚠️ Partial | ✅ Best choice | ⚠️ Partial | ⚠️ Partial |
| **Low latency required** | ✅ Yes | ❌ No | ❌ No | ❌ No |
| **Strict "N requests in last T seconds"** | ❌ No | ❌ No | ✅ Yes | ⚠️ Approx |
| **Abuse / brute-force prevention** | ❌ No | ❌ No | ✅ Yes | ⚠️ Approx |
| **Memory efficiency** | ✅ High | ✅ High | ❌ Low | ✅ Very high |
| **Precision** | ⚠️ Approx | ⚠️ Approx | ✅ Exact | ⚠️ Approx |
| **Easy to reason about** | ✅ Yes | ⚠️ Medium | ⚠️ Medium | ❌ Harder |
| **Scales to very high QPS** | ✅ Yes | ✅ Yes | ❌ No | ✅ Yes |
| **Common in API gateways** | ✅ Yes | ⚠️ Sometimes | ⚠️ Sometimes | ✅ Yes |
| **Adds latency under load** | ❌ No | ✅ Yes | ❌ No | ❌ No |

---

## Design Considerations

### Lazy Time Computation

All implementations use lazy updates instead of:
- Background goroutines
- Timers
- Periodic ticks

Instead, we compute:
- "How much time has passed?"
- "What should the state be now?"

**Benefits:**
- Consumes zero resources when idle
- Easier to reason about
- Avoids race conditions

*This is how real systems do it.*

### Concurrency Model

All rate limiters:
- Use mutexes to protect shared state
- Are safe under concurrent access
- Are tested using goroutines and WaitGroups
- Pass the race detector (`go run -race`)

---

## Running the Tests

Each algorithm has a test that:
- Fires concurrent requests
- Measures allowed vs rejected
- Observes behavior over time
- Logs timing gaps to visualize bursts vs smoothing

**Run the entire repository:**
```bash
go run .
```

Select the algorithm from the menu and observe:
- Burst behavior
- Refill behavior
- Rejection patterns

---

## What This Repository Is NOT

- ❌ Not a production-ready library
- ❌ Not optimized for every edge case
- ❌ Not an opinionated framework

**This repository is for learning and understanding.**

---

## Key Takeaways

1. **Rate limiting is about trade-offs, not rules**
2. **Bursts are sometimes good, sometimes dangerous**
3. **Output rate matters more than input rate in many systems**
4. **Time-based systems must be reasoned about carefully**
5. **Correctness under concurrency is non-negotiable**

---

## Learning Resources

I used these resources to learn! ->

- **Video Tutorial:** [System Design - Rate Limiting](https://youtu.be/CVItTb_jdkE?si=wBX1fT5omGPgSQgP)
- **System Design Interview Guide:** [ByteByteGo - Design a Rate Limiter](https://bytebytego.com/courses/system-design-interview/design-a-rate-limiter)
- **Technical Deep Dive:** [Arpit Bhayani - Sliding Window Rate Limiter](https://arpitbhayani.me/blogs/sliding-window-ratelimiter/)

---

## How to Learn From This

1. **Take it slow** — Don't copy-paste
2. **Predict behavior** before running tests
3. **Modify parameters** and observe changes
4. **Ask "why"** for every line

*That's how these systems actually stick.*

---

**This is a learn-in-public repository. If you are a beginner trying to understand rate limiting deeply, this repo is for you.**
