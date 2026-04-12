# Platform API Load Baseline Template

## Context

- Environment:
- Image digest:
- Database version / migration status:
- Date:
- Operator:

## Test Command

```powershell
$env:BASE_URL='http://127.0.0.1:18080'
k6 run --summary-export .\perf\platform-api\platform-api-k6-summary.json .\perf\platform-api\k6-smoke.js
```

## Result Summary

- Total requests:
- Failed request rate:
- Average duration:
- p95 duration:
- Max VUs:

## Decision

- Pass / Fail:
- Notes:

