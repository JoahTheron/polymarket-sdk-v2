# shared Package (types)

Shared JSON scalar helpers used across all Polymarket client packages.

## Custom Types

Polymarket's API returns numeric values as both raw JSON numbers **and** decimal strings (e.g. `"0.50"` vs `0.5`). These types handle the ambiguity:

### Float64

Always unmarshals as `float64`, regardless of input format:

```go
// Accepts: "0.50" → 0.5, 0.5 → 0.5, null → 0
var f shared.Float64
json.Unmarshal([]byte(`"0.50"`), &f)
fmt.Println(f.Value) // 0.5
```

### String

Converts strings, numbers, and booleans into a stable string form:

```go
var s shared.String
json.Unmarshal([]byte(`123`), &s)
fmt.Println(s.Value) // "123"
```

### Int / Int64 / Uint64

Integer types that accept both JSON numbers and quoted strings:

```go
var i shared.Int
json.Unmarshal([]byte(`"42"`), &i)
fmt.Println(i.Value) // 42
```

### Time / Date

Accept common Polymarket timestamp and date encodings:

```go
var t shared.Time
json.Unmarshal([]byte(`"2024-01-15T10:30:00Z"`), &t)
```

### StringSlice / Float64Slice

Accept arrays and string-encoded arrays:

```go
// Accepts: ["0.1", "0.2"] or "0.1,0.2"
var fs shared.Float64Slice
json.Unmarshal([]byte(`"0.1,0.2"`), &fs)
```

## Page[T]

Generic pagination wrapper:

```go
type Page[T any] struct {
    Results    []T
    NextCursor string
    Limit      *int64
}
```
