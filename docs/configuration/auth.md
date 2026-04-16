# Auth

Authentication can be set per step in the scenario file.

## Basic auth

```yaml
steps:
  - method: GET
    url: /secure
    auth:
      basic:
        username: admin
        password: secret
```

## Bearer token

```yaml
steps:
  - method: GET
    url: /secure
    auth:
      bearer: your-token-here
```

## Custom header

```yaml
steps:
  - method: GET
    url: /secure
    auth:
      header:
        key: X-API-Key
        value: your-api-key
```

Or use it for any arbitrary header-based auth scheme:

```yaml
auth:
  header:
    key: Authorization
    value: "Token abc123"
```
