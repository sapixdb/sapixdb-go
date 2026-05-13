# SapixDB Go SDK

Official Go client for [SapixDB](https://sapixdb.com) — the agent-native living database.

Zero dependencies. Uses only the Go standard library. Go 1.21+.

## Installation

```bash
go get github.com/sapixdb/sapixdb-go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"

    sapixdb "github.com/sapixdb/sapixdb-go"
)

func main() {
    ctx := context.Background()

    db := sapixdb.New(sapixdb.Config{
        URL:   "http://localhost:7475",
        Agent: "my-app",
    })

    // Check connection
    fmt.Println(db.Ping(ctx)) // true

    // Write a record
    record, err := db.Collection("products").Write(ctx, map[string]any{
        "name":  "Classic T-Shirt",
        "price": 29.99,
        "stock": 100,
    })
    if err != nil {
        panic(err)
    }
    fmt.Println(record.ID)   // "nuc_abc123"
    fmt.Println(record.Hash) // "sha3:e7f2a1..."

    // Read latest records
    products, err := db.Collection("products").Latest(ctx, nil)

    // Filter
    shirts, err := db.Collection("products").Find(ctx, map[string]any{
        "category": "apparel",
    }, 0)

    // Time travel
    snapshot, err := db.Collection("orders").
        AsOf("2024-01-01T00:00:00Z").
        Latest(ctx, nil)

    _ = products
    _ = shirts
    _ = snapshot
}
```

## API Reference

### Creating a Client

```go
db := sapixdb.New(sapixdb.Config{
    URL:     "http://localhost:7475", // SapixDB agent URL
    Agent:   "my-app",               // agent ID
    Headers: map[string]string{},    // optional extra headers
    Timeout: 10 * time.Second,       // optional, default 10s
})
```

---

### Collection API — `db.Collection(name)`

#### `.Write(ctx, data)` → `*WriteResult, error`
Append a new record. Nothing is ever overwritten — every write is permanent.

#### `.WriteBatch(ctx, records)` → `[]*WriteResult, error`
Write multiple records sequentially.

#### `.Get(ctx, recordID)` → `*NucleotideRecord, error`
Fetch by ID. Returns `*SapixNotFoundError` if missing.

#### `.Latest(ctx, filter)` → `[]NucleotideRecord, error`
Current (most recent) version of every record. Pass `nil` for no filter.

#### `.History(ctx, filter)` → `[]NucleotideRecord, error`
Full append-only history — every version ever written.

#### `.Find(ctx, filter, limit)` → `[]NucleotideRecord, error`
Filter records (latest version only). Pass `0` for no limit.

#### `.FindOne(ctx, filter)` → `*NucleotideRecord, error`
First match or `nil`. Never errors on empty result.

#### `.AsOf(timestamp)` → `*CollectionQuery`
Scope reads to a point in time. Returns a query object with `.Latest()`, `.All()`, `.Find()`, `.FindOne()`.

---

### Time Travel

```go
ts := time.Now().Add(-30 * time.Minute).UTC().Format(time.RFC3339)

snapshot, err := db.Collection("orders").
    AsOf(ts).
    Find(ctx, map[string]any{"customer_id": "cust_001"}, 0)
// Returns orders exactly as they existed 30 minutes ago
```

---

### Graph API — `db.Graph`

#### `.Relate(ctx, src, dst, edgeType, weight)` → `error`
Create a typed directed edge. The cleanest way to express relationships.

#### `.AddEdge(ctx, src, dst, edgeType, weight)` → `error`
Full edge creation (same as Relate).

#### `.RemoveEdge(ctx, src, dst, edgeType)` → `error`
Delete an edge.

#### `.Traverse(ctx, fromID, TraverseOptions)` → `*TraverseResult, error`
Walk the graph. Returns `.Nodes` and `.Edges`.
Direction: `"outbound"` | `"inbound"` | `"both"`.

#### `.Neighbors(ctx, nodeID, direction)` → `[]NucleotideRecord, error`
Direct neighbours — depth=1 shortcut.

#### `.Edges(ctx, nodeID)` → `[]GraphEdge, error`
All outbound edges from a node.

```go
// Link order → customer
_ = db.Graph.Relate(ctx, order.ID, customer.ID, "placed_by", 1.0)
_ = db.Graph.Relate(ctx, order.ID, shirt.ID, "contains", 1.0)

// Find everything connected to a customer (depth 2)
result, err := db.Graph.Traverse(ctx, customer.ID, sapixdb.TraverseOptions{
    Depth:     2,
    Direction: "inbound",
})
fmt.Println(result.Nodes) // []NucleotideRecord
fmt.Println(result.Edges) // []GraphEdge
```

---

### Ingest — `db.Ingest(ctx, collection, data)`

```go
_, err := db.Ingest(ctx, "ai_decisions", map[string]any{
    "model":      "gpt-4o",
    "action":     "approve_loan",
    "confidence": 0.94,
    "reasoning":  "Credit score 780, DTI 28%",
})
// Every decision is cryptographically signed — immutable audit trail.
```

---

### Error Handling

```go
import "errors"

record, err := db.Collection("orders").Get(ctx, "nuc_missing")
if err != nil {
    var notFound *sapixdb.SapixNotFoundError
    var netErr   *sapixdb.SapixNetworkError
    var sapixErr *sapixdb.SapixError

    switch {
    case errors.As(err, &notFound):
        fmt.Println("not found:", notFound.RecordID)
    case errors.As(err, &netErr):
        fmt.Println("network error:", netErr.Cause)
    case errors.As(err, &sapixErr):
        fmt.Printf("error %d: %s\n", sapixErr.Status, sapixErr.Message)
    }
}
```

---

### Full Example: Online Store

```go
package main

import (
    "context"
    "fmt"

    sapixdb "github.com/sapixdb/sapixdb-go"
)

func main() {
    ctx := context.Background()
    db := sapixdb.New(sapixdb.Config{
        URL:   "http://localhost:7475",
        Agent: "store",
    })

    // 1. Add product
    shirt, _ := db.Collection("products").Write(ctx, map[string]any{
        "sku": "SHIRT-001", "name": "Classic T-Shirt",
        "price": 29.99, "stock": 200, "category": "apparel",
    })

    // 2. Register customer
    customer, _ := db.Collection("customers").Write(ctx, map[string]any{
        "name": "Alice Johnson", "email": "alice@example.com",
    })

    // 3. Place order
    order, _ := db.Collection("orders").Write(ctx, map[string]any{
        "customer_id": customer.ID,
        "items": []any{map[string]any{
            "product_id": shirt.ID, "qty": 2, "unit_price": 29.99,
        }},
        "total":  59.98,
        "status": "placed",
    })

    // 4. Link in graph
    _ = db.Graph.Relate(ctx, order.ID, customer.ID, "placed_by", 1.0)
    _ = db.Graph.Relate(ctx, order.ID, shirt.ID, "contains", 1.0)

    // 5. Ship (append — "placed" version is preserved forever)
    _, _ = db.Collection("orders").Write(ctx, map[string]any{
        "customer_id": customer.ID,
        "status":      "shipped",
        "tracking":    "UPS-1Z999AA10123456784",
    })

    // 6. Audit: what was the status when it was placed?
    original, _ := db.Collection("orders").
        AsOf(order.Timestamp).
        FindOne(ctx, map[string]any{"customer_id": customer.ID})
    fmt.Println(original.Data["status"]) // "placed" — not "shipped"
}
```
