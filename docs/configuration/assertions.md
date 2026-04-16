# Assertions

Assertions define SLA thresholds that automatically mark a test as passed or failed.

## Configuration

```yaml
assertions:
  - metric: p95_latency
    operator: less_than
    value: 500
    enabled: true
```

## Available metrics

| Metric | Description |
|--------|-------------|
| `p50_latency` | 50th percentile latency (ms) |
| `p90_latency` | 90th percentile latency (ms) |
| `p95_latency` | 95th percentile latency (ms) |
| `p99_latency` | 99th percentile latency (ms) |
| `avg_latency` | Average latency (ms) |
| `max_latency` | Maximum latency (ms) |
| `rps` | Requests per second |
| `error_rate` | Percentage of failed requests |
| `success_rate` | Percentage of successful requests |
| `total_requests` | Total requests made |
| `total_errors` | Total failed requests |

## Available operators

| Operator | Meaning |
|----------|---------|
| `less_than` | metric < value |
| `less_than_or_equal` | metric <= value |
| `greater_than` | metric > value |
| `greater_than_or_equal` | metric >= value |
| `equal` | metric == value |

## Example

```yaml
assertions:
  - metric: p95_latency
    operator: less_than
    value: 500
    enabled: true

  - metric: p99_latency
    operator: less_than
    value: 1000
    enabled: true

  - metric: error_rate
    operator: less_than
    value: 1
    enabled: true

  - metric: rps
    operator: greater_than
    value: 100
    enabled: true
```

Use `enabled: false` to define an assertion without enforcing it — useful for tracking a metric without failing the test.
