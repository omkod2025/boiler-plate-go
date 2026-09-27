# JWT Duration Configuration Examples

## Overview
ตัวอย่างการกำหนดค่า JWT_DURATION ในรูปแบบต่างๆ สำหรับ environment ที่แตกต่างกัน

## Duration Format Reference

| Format | Example | Description |
|--------|---------|-------------|
| `ns` | `1000000000ns` | nanoseconds |
| `us` | `1000000us` | microseconds |
| `ms` | `1000ms` | milliseconds |
| `s` | `60s` | seconds |
| `m` | `30m` | minutes |
| `h` | `2h` | hours |

ค่าถูกอ่านด้วย `time.ParseDuration` ของ Go ซึ่ง **ไม่รองรับหน่วย `d` (วัน)** ให้ใช้ชั่วโมงแทน
เช่น 7 วัน = `168h`, 30 วัน = `720h` ถ้ารูปแบบไม่ถูกต้องระบบจะใช้ค่าเริ่มต้น `24h` แทนโดยไม่แจ้ง error

## Environment-Specific Examples

### Development Environment
```env
# Long duration for development convenience
JWT_DURATION=24h                    # 24 hours
JWT_DURATION=168h                   # 7 days
JWT_DURATION=720h                   # 30 days
```

### Staging Environment
```env
# Medium duration for testing
JWT_DURATION=2h                     # 2 hours
JWT_DURATION=6h                     # 6 hours
JWT_DURATION=12h                    # 12 hours
```

### Production Environment
```env
# Standard production duration
JWT_DURATION=1h                     # 1 hour
JWT_DURATION=30m                    # 30 minutes

# High security production
JWT_DURATION=15m                    # 15 minutes
JWT_DURATION=5m                     # 5 minutes
```

### Microservices Architecture
```env
# Auth Service
JWT_DURATION=1h                     # Standard access token

# Payment Service (Higher security)
JWT_DURATION=30m                    # Shorter duration for sensitive operations

# Admin Service (Highest security)
JWT_DURATION=15m                    # Very short duration for admin access

# Public API Service
JWT_DURATION=2h                     # Longer duration for public APIs
```

## Complex Duration Examples

### Combined Units
```env
JWT_DURATION=1h30m                  # 1 hour 30 minutes
JWT_DURATION=2h15m30s               # 2 hours 15 minutes 30 seconds
JWT_DURATION=36h                    # 1 day 12 hours
JWT_DURATION=174h                   # 7 days 6 hours
```

### Short Durations (High Security)
```env
JWT_DURATION=5m                     # 5 minutes
JWT_DURATION=10m                    # 10 minutes
JWT_DURATION=15m                    # 15 minutes
JWT_DURATION=30m                    # 30 minutes
```

### Medium Durations (Standard)
```env
JWT_DURATION=1h                     # 1 hour
JWT_DURATION=2h                     # 2 hours
JWT_DURATION=4h                     # 4 hours
JWT_DURATION=6h                     # 6 hours
```

### Long Durations (Development/Refresh)
```env
JWT_DURATION=12h                    # 12 hours
JWT_DURATION=24h                    # 24 hours
JWT_DURATION=168h                   # 7 days
JWT_DURATION=720h                   # 30 days
JWT_DURATION=2160h                  # 90 days
```

## Best Practices by Use Case

### 1. Development
```env
# For development convenience
JWT_DURATION=24h                    # 24 hours
JWT_DURATION=168h                   # 7 days
```

### 2. Testing/Staging
```env
# For testing environments
JWT_DURATION=2h                     # 2 hours
JWT_DURATION=6h                     # 6 hours
```

### 3. Production - Standard
```env
# For regular production applications
JWT_DURATION=1h                     # 1 hour
JWT_DURATION=30m                    # 30 minutes
```

### 4. Production - High Security
```env
# For high-security applications
JWT_DURATION=15m                    # 15 minutes
JWT_DURATION=5m                     # 5 minutes
```

### 5. Microservices
```env
# Auth Service
JWT_DURATION=1h

# Payment Service
JWT_DURATION=30m

# Admin Service
JWT_DURATION=15m

# Public API
JWT_DURATION=2h
```

### 6. Refresh Token Pattern
```env
# Access Token (Short-lived)
JWT_DURATION=15m

# Refresh Token (Long-lived)
JWT_DURATION=168h
```

## Security Considerations

### 1. Token Duration vs Security
- **Shorter duration** = Higher security, more frequent re-authentication
- **Longer duration** = Lower security, less frequent re-authentication

### 2. Environment-Specific Security
```env
# Development: Convenience over security
JWT_DURATION=24h

# Staging: Balance between convenience and security
JWT_DURATION=2h

# Production: Security over convenience
JWT_DURATION=1h
```

### 3. Application-Specific Security
```env
# Banking/Financial: Very high security
JWT_DURATION=5m

# E-commerce: High security
JWT_DURATION=15m

# Social Media: Standard security
JWT_DURATION=1h

# Internal Tools: Lower security
JWT_DURATION=4h
```

## Implementation Examples

### 1. Single Environment
```env
# .env file
JWT_DURATION=1h
```

### 2. Multiple Environments
```env
# .env.development
JWT_DURATION=24h

# .env.staging
JWT_DURATION=2h

# .env.production
JWT_DURATION=1h
```

### 3. Docker/Kubernetes
```yaml
# docker-compose.yml
environment:
  - JWT_DURATION=1h

# kubernetes deployment.yaml
env:
- name: JWT_DURATION
  value: "1h"
```

### 4. CI/CD Pipeline
```bash
# Set duration based on environment
if [ "$ENVIRONMENT" = "production" ]; then
  export JWT_DURATION="1h"
elif [ "$ENVIRONMENT" = "staging" ]; then
  export JWT_DURATION="2h"
else
  export JWT_DURATION="24h"
fi
```

## Monitoring and Logging

### 1. Log Token Duration
```go
logger.Info("JWT token duration configured", map[string]interface{}{
    "duration": config.Env.JWT_DURATION,
    "environment": config.Env.APP_ENV,
})
```

### 2. Alert on Long Durations
```go
if config.Env.JWT_DURATION > 1*time.Hour && config.Env.APP_ENV == "production" {
    logger.Warn("Long JWT duration detected in production", map[string]interface{}{
        "duration": config.Env.JWT_DURATION,
    })
}
```

## Migration Guide

### From Fixed Duration to Environment-Based
```go
// Before
const tokenDuration = 24 * time.Hour

// After
tokenDuration := config.Env.JWT_DURATION
```

### From Secret Key to RSA with Duration
```go
// Before
jwtConfig := middleware.JWTConfig{
    SecretKey: "your-secret",
    TokenDuration: 24 * time.Hour,
}

// After
jwtConfig := middleware.JWTConfig{
    PrivateKey: privateKey,
    PublicKey: publicKey,
    TokenDuration: config.Env.JWT_DURATION,
}
``` 