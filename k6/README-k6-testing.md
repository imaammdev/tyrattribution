# K6 Load Testing for TyrAttribution

This directory contains K6 load testing scripts for stress testing the click event and conversion event APIs.

## Test Scripts

### 1. `k6-stress-test.js` - Main Stress Test
Comprehensive stress test that simulates realistic traffic patterns.

**Features:**
- Ramps up from 0 to 200 concurrent users
- 70% click events, 30% conversion events
- Realistic data generation with limited campaign/source sets
- Attribution scenarios testing
- Performance thresholds (95% requests < 500ms, error rate < 10%)

**Usage:**
```bash
k6 run k6-stress-test.js
```

**Environment Variables:**
```bash
# Custom base URL (default: http://localhost:8080)
k6 run -e BASE_URL=http://your-api-url k6-stress-test.js
```

### 2. `k6-attribution-test.js` - Attribution Testing
Focused test for click-to-conversion attribution scenarios.

**Features:**
- Creates click event followed by conversion event
- Tests attribution logic within time windows
- Tracks successful attribution rate
- Simulates realistic user behavior timing

**Usage:**
```bash
k6 run k6-attribution-test.js
```

### 3. `k6-load-profiles.js` - Multiple Load Profiles
Contains different load testing patterns for various scenarios.

**Profiles Available:**
- **Spike Test:** Sudden traffic spikes
- **Soak Test:** Sustained load over long periods
- **Breakpoint Test:** Find maximum system capacity
- **Volume Test:** High data throughput testing

**Usage:**
```bash
# Spike test
k6 run --config '{"stages":[{"duration":"1m","target":10},{"duration":"10s","target":500},{"duration":"1m","target":10}]}' k6-load-profiles.js

# Soak test (30 minutes)
k6 run --config '{"stages":[{"duration":"5m","target":50},{"duration":"30m","target":50},{"duration":"5m","target":0}]}' k6-load-profiles.js

# Breakpoint test
k6 run --config '{"executor":"ramping-arrival-rate","startRate":10,"timeUnit":"1s","preAllocatedVUs":50,"maxVUs":1000,"stages":[{"duration":"2m","target":10},{"duration":"5m","target":50},{"duration":"5m","target":100},{"duration":"5m","target":200},{"duration":"5m","target":300},{"duration":"5m","target":500}]}' k6-load-profiles.js
```

## API Endpoints Tested

### Click Event API
- **Endpoint:** `POST /api/clicks`
- **Payload:**
```json
{
  "campaign_id": "uuid",
  "user_id": "uuid",
  "click_date": "2025-09-21T10:00:00Z",
  "source": "google"
}
```

### Conversion Event API
- **Endpoint:** `POST /api/conversions`
- **Payload:**
```json
{
  "user_id": "uuid",
  "campaign_id": "uuid",
  "conversion_date": "2025-09-21T10:05:00Z",
  "value": 99.99,
  "type": "purchase",
  "source": "google"
}
```

## Performance Thresholds

### Standard Thresholds
- **Response Time:** 95% of requests < 500ms
- **Error Rate:** < 10% failed requests
- **Availability:** > 99% success rate

### Spike Test Thresholds (More Lenient)
- **Response Time:** 95% of requests < 2000ms
- **Error Rate:** < 20% failed requests

### Soak Test Thresholds (Stricter)
- **Response Time:** 95% of requests < 500ms
- **Error Rate:** < 5% failed requests

## Test Data Generation

### Campaign IDs
Uses a limited set of predefined campaign IDs to simulate realistic scenarios:
- `a1b2c3d4-e5f6-7890-abcd-ef1234567890`
- `b2c3d4e5-f6g7-8901-bcde-f23456789012`
- `c3d4e5f6-g7h8-9012-cdef-345678901234`
- And more...

### Traffic Sources
- google, facebook, twitter, instagram, youtube, tiktok

### Conversion Types
- purchase, signup, download, subscription, contact

### Realistic Patterns
- Random timestamps within last 24 hours
- Value ranges: $10-$510 for conversions
- Attribution windows: 1-24 hours between click and conversion

## Monitoring and Metrics

### Built-in K6 Metrics
- `http_req_duration`: Request response time
- `http_req_failed`: Failed request rate
- `http_reqs`: Total number of HTTP requests
- `vus`: Number of active virtual users

### Custom Metrics
- `errors`: Custom error rate tracking
- `successful_attributions`: Attribution success rate (attribution test)

## Running Tests

### Prerequisites
1. Install K6: https://k6.io/docs/getting-started/installation/
2. Ensure your API server is running
3. Database and Redis should be available

### Basic Commands
```bash
# Run main stress test
k6 run k6-stress-test.js

# Run with custom URL
k6 run -e BASE_URL=http://localhost:8080 k6-stress-test.js

# Run attribution test
k6 run k6-attribution-test.js

# Run with different VU count
k6 run --vus 50 --duration 5m k6-stress-test.js

# Generate HTML report
k6 run --out json=results.json k6-stress-test.js
```

### Interpreting Results

**Good Performance Indicators:**
- Response times consistently under 500ms
- Error rate below 5%
- Successful attribution rate above 95%
- No memory leaks during soak tests

**Warning Signs:**
- Response times trending upward
- Error rates above 10%
- Failed attribution scenarios
- High database connection errors

## Troubleshooting

### Common Issues

1. **High Error Rates**
   - Check database connection limits
   - Verify Redis memory capacity
   - Review application logs

2. **Slow Response Times**
   - Check database query performance
   - Monitor Redis performance
   - Review application resource usage

3. **Attribution Failures**
   - Verify time window configuration
   - Check Redis key expiration settings
   - Review attribution logic

### Database Considerations
- Ensure adequate connection pool size
- Monitor query performance
- Consider indexing on frequently queried columns

### Redis Considerations
- Monitor memory usage
- Check key expiration policies
- Verify connection limits

## Example Output

```
running (16m0.0s), 000/200 VUs, 50000 complete and 0 interrupted iterations
default ✓ [======================================] 200 VUs  16m0s

✓ click event status is 201
✓ conversion event status is 201
✓ click event response time < 500ms
✓ conversion event response time < 500ms

http_req_duration..............: avg=245ms    min=12ms    med=198ms   max=2.1s    p(95)=456ms  p(99)=1.2s
http_req_failed................: 2.34%        ✓ 1170     ✗ 48830
http_reqs......................: 50000        52.08/s
errors.........................: 2.34%        ✓ 1170     ✗ 48830
```