# Modular Monolith Architecture

## Why This Matters

Most beginners think architecture choices are about **scale**.

In reality, **architecture is about controlling complexity**.

A system usually fails long before it runs out of CPU or memory. It fails when:

- Changes become risky
- Ownership becomes unclear
- Everything depends on everything

**A Modular Monolith exists to prevent that.**

---

## What a Modular Monolith Is (In One Line)

> A Modular Monolith is a **single deployable application** where internal modules have **strict boundaries that are enforced in code**.

If boundaries are only conventions, it's not modular.

---

## High-Level Architecture (Mental Model)

This is what a modular monolith looks like conceptually:

```
+--------------------------------------------------+
|                  Application                     |
|                                                  |
|  +-----------+   +-----------+   +-----------+  |
|  |  Users    |   |  Orders   |   | Payments  |  |
|  |  Module   |   |  Module   |   |  Module   |  |
|  +-----------+   +-----------+   +-----------+  |
|        |                |                |      |
|        |     Explicit Interfaces / Events        |
|        +-----------------------------------------+
|                                                  |
+--------------------------------------------------+
```

**Key idea:**
- One process
- One deployment
- But modules do not freely access each other

---

## Why This Is Not "Just a Monolith"

Compare this to a traditional monolith:

```
Controllers
    |
Services
    |
Repositories
    |
Database
```

Everything can call everything. Coupling spreads silently.

**In a Modular Monolith, dependencies are directional and limited.**

---

## Domain-Based Module Split (Important)

Modules should represent **business capabilities**, not technical layers.

### ❌ Bad split:
```
controllers/
services/
repositories/
```

### ✅ Good split:
```
users/
orders/
payments/
```

Each module is a **closed world**.

---

## File Structure Example 

Here is a minimal but correct modular monolith layout in Go:

```
myapp/
├── cmd/
│   └── server/
│       └── main.go
│
├── internal/
│   ├── users/
│   │   ├── api.go          // public interface of users module
│   │   └── internal/
│   │       ├── service.go
│   │       ├── repository.go
│   │       └── model.go
│   │
│   ├── orders/
│   │   ├── api.go
│   │   └── internal/
│   │       ├── service.go
│   │       ├── repository.go
│   │       └── model.go
│   │
│   └── payments/
│       ├── api.go
│       └── internal/
│           ├── service.go
│           ├── repository.go
│           └── model.go
│
└── go.mod
```

**Critical rules enforced by Go:**
- Code inside `internal/` cannot be imported outside the module
- Other modules can only access what is exposed in `api.go`

This is how **boundaries become real**.

---

## Module Communication 

Modules talk via **explicit interfaces**, not direct imports.

**Example:** Orders module needs user data.

### `users/api.go`

```go
package users

import "context"

type User struct {
    ID   string
    Name string
}

type Reader interface {
    GetUser(ctx context.Context, id string) (User, error)
}
```

### `orders/internal/service.go`

```go
package internal

import "context"

type UserReader interface {
    GetUser(ctx context.Context, id string) (User, error)
}

type OrderService struct {
    users UserReader
}

func (s *OrderService) PlaceOrder(ctx context.Context, userID string) error {
    user, err := s.users.GetUser(ctx, userID)
    if err != nil {
        return err
    }
    // business logic here
    return nil
}
```

**Key idea:**
- Orders knows **what it needs**
- It does **not** know how users module works internally

---

## Module Communication Diagram

```
+-----------+        interface        +-----------+
|  Orders   | ---------------------> |  Users    |
|  Module   |    UserReader           |  Module   |
+-----------+                         +-----------+
       ^                                    |
       |           internal details hidden |
       +------------------------------------+
```

This mirrors microservices contracts, **without HTTP**.

---

## Data Ownership (Even With One Database)

Even if all modules share one database:

```
Database
├── users_table     (owned by users module)
├── orders_table    (owned by orders module)
└── payments_table  (owned by payments module)
```

**Rules:**
- Orders does **NOT** query `users_table`
- Users module decides what user data is exposed
- Ownership is **logical**, not physical

---

## Domain Events (Inside the Monolith)

Instead of deep call chains, use **events**.

```
[ Users Module ]
      |
      | UserCreated
      v
[ Orders Module ] -----> reacts
      |
      v
[ Payments Module ]
```

This keeps modules decoupled and prepares future extraction.

---

## Transactions (Common Failure Point)

### ❌ Bad approach:
```
One giant transaction touching users + orders + payments
```

### ✅ Better approach:
```
Users updates state
→ emits event
→ other modules react independently
```

This avoids tight coupling and long-lived locks.

---

## Modular Monolith vs Microservices (Summary)

| **Modular Monolith** | **Microservices** |
|----------------------|-------------------|
| One deploy | Many deploys |
| In-process calls | Network calls |
| Simple debugging | Complex failure modes |
| Low ops overhead | High ops overhead |

**Modular Monolith is often the right default.**

---

## When a Modular Monolith Fails

It fails when:

- Boundaries are ignored
- Modules reach into each other
- Shared helpers become shared logic
- Discipline is manual, not enforced

At that point, it becomes a **big ball of mud** again.

---

## Final Mental Model

> A Modular Monolith is not anti-microservices.
> 
> It is **microservices thinking without the network**.
> 
> You earn distribution by first mastering boundaries.

---

## Closing

If caching taught you where inconsistency is acceptable, and rate limiting taught you where load must stop...

**Modular Monoliths teach you where coupling must stop.**

That skill transfers everywhere.

---

## Learning Resources

### 📺 Videos

- **Intro to Modular Monolith Architecture** — Simon Brown (C4 architecture expert)  
  [Watch on YouTube](https://www.youtube.com/watch?v=kbKxmEeuvc4)

- **Prime Video architectural shift discussion** (microservices → monolith)  
  [Watch on YouTube](https://www.youtube.com/watch?v=dV3wAe8HV7Q)

- **Another take on monolith reversion and costs**  
  [Watch on YouTube](https://www.youtube.com/watch?v=9JPYCOpeDnY)

### 📄 Articles

- **What is a Modular Monolith?** (GeeksforGeeks explanation)  
  [Read Article](https://www.geeksforgeeks.org/system-design/what-is-a-modular-monolith/)

- **Case study of Amazon Prime Video's monolith move** and cost/performance outcomes  
  [Read on Medium](https://medium.com/@hellomeenu1/why-amazon-prime-video-reverted-to-a-monolith-a-case-study-on-cloud-architecture-evolution-bd2582b438a5)

- **General architectural perspective** on modular monolith patterns  
  [Read on Dev.to](https://dev.to/naveens16/behold-the-modular-monolith-the-architecture-balancing-simplicity-and-scalability-2d4)

- **Broader reflections on why distributed microservices sometimes fail**  
  [Read on Nordic APIs](https://nordicapis.com/back-to-the-monolith-why-did-amazon-dump-microservices/)