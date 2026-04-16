# Load Profiles

A load profile controls how worker concurrency changes over time.

## constant

Fixed number of workers for the full duration. Use this for steady-state benchmarks.

```yaml
load:
  profile: constant
  workers: 20
  duration: 60s
```

---

## ramp

Linearly increases workers from `start_workers` to `end_workers` over the ramp duration. Use this to find the degradation point without shocking the system.

```yaml
load:
  profile: ramp
  duration: 60s
  ramp_up:
    start_workers: 1
    end_workers: 50
    duration: 30s
```

---

## step

Adds workers in fixed increments at regular intervals. Use this to observe how latency changes as load increases in discrete steps.

```yaml
load:
  profile: step
  duration: 60s
  step:
    start_workers: 5
    step_size: 5
    step_duration: 10s
    max_workers: 50
```

---

## spike

Runs a base level of workers with periodic traffic bursts. Use this to test how your system recovers after a sudden surge.

```yaml
load:
  profile: spike
  duration: 60s
  spike:
    base_workers: 10
    spike_workers: 50
    spike_duration: 5s
    spike_every: 15s
```

---

## Stopping conditions

Use `duration` to run for a fixed time, or `max_requests` to stop after a total request count:

```yaml
load:
  profile: constant
  workers: 10
  max_requests: 1000
```
