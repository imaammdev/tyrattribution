
## Architecture Overview

### Core Architecture Features
1. **Event-Driven Architecture**: Uses Kafka for reliable message streaming
2. **Dual Storage Strategy**: Redis for real-time data, PostgreSQL for persistent storage
3. **Microservice-Oriented**: Clean separation of concerns with repository pattern
4. **High Throughput**: Asynchronous processing for optimal performance

## Why Kafka Direct Publishing?

The system publishes events directly to Kafka instead of processing them synchronously for several critical reasons:

### 1. **Decoupling & Resilience**
- **Fault Tolerance**: If the database is down, events are safely queued in Kafka
- **Service Independence**: HTTP endpoints remain responsive regardless of downstream processing load
- **Graceful Degradation**: System can continue accepting events even during heavy processing loads

### 2. **Performance Benefits**
- **Low Latency Response**: HTTP endpoints return immediately after publishing to Kafka (~1-2ms vs 50-100ms for database writes)
- **High Throughput**: Can handle burst traffic without overwhelming the database
- **Horizontal Scaling**: Multiple consumers can process events in parallel

### 3. **Reliability & Durability**
- **Message Persistence**: Kafka provides durable storage with configurable retention
- **At-Least-Once Delivery**: Ensures no events are lost during processing
- **Replay Capability**: Can reprocess events if needed for data recovery or analytics

### 4. **Event Sourcing Benefits**
- **Audit Trail**: Complete history of all events in chronological order
- **Data Recovery**: Can rebuild state from event stream
- **Analytics**: Raw events available for complex analytical queries

### Configuration
```go
// Kafka Producer Configuration
config.Producer.Return.Successes = true
config.Producer.RequiredAcks = sarama.WaitForAll  // Wait for all replicas to acknowledge
config.Producer.Retry.Max = 3                     // Retry failed messages
```

## Statistical API Performance Optimizations

The system implements a sophisticated dual-layer caching strategy for optimal performance:

### 1. **Real-time Data Layer (Redis)**
- **Current Day Counters**: Live click and conversion counts stored as Redis counters
- **TTL Management**: Automatic expiration prevents memory bloat
- **Sub-second Access**: Redis provides microsecond-level response times

```go
// Real-time counter examples
clickKey := fmt.Sprintf("click_count:%s:%s", campaignID, today)
conversionKey := fmt.Sprintf("conversion_count:%s:%s", campaignID, today)
```

### 2. **Historical Data Layer (PostgreSQL)**
- **Pre-aggregated Tables**: Daily campaign journals for historical periods
- **Batch Processing**: Historical data updated via background jobs
- **Minimal Query Scope**: Excludes current day to avoid cache invalidation

### 3. **Hybrid Data Fusion**
The statistics API combines both layers intelligently:

```go
func (s *CampaignStatisticsServiceImpl) GetCampaignStatistics() {
    // 1. Get historical data from PostgreSQL (excludes today)
    historicalData := s.campaignStatsRepo.GetHistoricalData(ctx, campaignID, groupBy)

    // 2. Get real-time today data from Redis
    todayData := s.getTodayData(ctx, campaignID)

    // 3. Merge datasets for complete view
    return s.combineData(historicalData, todayData, groupBy)
}
```

### 4. **Performance Benefits**

| Metric | Traditional Approach | Optimized Approach |
|--------|---------------------|-------------------|
| Response Time | 200-500ms | 10-50ms |
| Database Load | High (real-time queries) | Low (pre-aggregated) |
| Scalability | Limited by DB | Redis throughput |
| Consistency | Eventual | Real-time current + consistent historical |


## Getting Started

### Prerequisites

- Go 1.25+
- Docker & Docker Compose
- PostgreSQL 17
- Apache Kafka 7.4.0
- Redis 7

### Installation & Setup

1. **Clone the repository**
```bash
git clone <repository-url>
cd tyrattribution
```

2. **Environment Configuration**
Create a `.env` file with the following variables:
```env
# Database Configuration
POSTGRES_USER=tyrattribution
POSTGRES_PASSWORD=yourpassword
POSTGRES_DB=tyrattribution
DB_PORT=5432
DB_HOST=localhost
DB_SSLMODE=disable

# Kafka Configuration
KAFKA_URL=localhost:9092
KAFKA_CLICK_TOPIC=click-events
KAFKA_CONVERSION_TOPIC=conversion-events

# Redis Configuration
REDIS_URL=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
```

3. **Start Infrastructure Services**
```bash
docker-compose up -d
```

4. **Build and Run the Application**
```bash
go mod download
go build -o tyrattribution
./tyrattribution
```

### API Endpoints

#### Event Tracking
- `POST /api/events/click` - Track click events
- `POST /api/events/conversion` - Track conversion events

#### Campaign Management
- `POST /api/campaigns/journal` - Update campaign journal
- `GET /api/campaigns/statistics?campaign_id=UUID&group_by=daily` - Get campaign statistics

#### Example Usage

**Track a Click Event:**
```bash
curl -X POST http://localhost:8080/api/events/click \
  -H "Content-Type: application/json" \
  -d '{
    "campaign_id": "123e4567-e89b-12d3-a456-426614174000",
    "user_id": "987fcdeb-51a2-43d1-9f12-345678901234",
    "source": "google_ads",
    "click_date": "2024-01-15T10:30:00Z"
  }'
```

**Get Campaign Statistics:**
```bash
curl "http://localhost:8080/api/campaigns/statistics?campaign_id=123e4567-e89b-12d3-a456-426614174000&group_by=daily"
```

### Testing

The project includes K6 load testing scripts for performance validation:

```bash
# Install K6
# Run performance tests
k6 run k6/click-event-test.js
k6 run k6/conversion-event-test.js
```

### Database Schema

The system uses the following key entities:

- **campaigns**: Campaign definitions and metadata
- **click_events**: Individual click tracking records
- **conversion_events**: Conversion tracking with attribution
- **campaign_journals**: Daily aggregated campaign metrics
- **campaign_statistics**: Pre-computed statistical summaries

### Scaling Considerations

1. **Horizontal Scaling**: Multiple application instances can run behind a load balancer
2. **Consumer Scaling**: Kafka consumers can be scaled independently
3. **Database Partitioning**: Consider time-based partitioning for large datasets
4. **Redis Clustering**: Use Redis cluster for high-availability scenarios
5. **Database Option**: Use leaderless database like casandra or other, better for heavy write
6. **CDC Tools**: Use Debezium for realtime data capture into data warehouse for better statistical

### Performance Tuning

1. **Kafka Partitioning**: Increase topic partitions for higher throughput
2. **Connection Pooling**: Configure appropriate database connection pools
3. **Redis Memory**: Monitor Redis memory usage and configure appropriate limits
4. **Batch Processing**: Optimize batch sizes for consumer processing

## Development

### Project Structure
```
tyrattribution/
├── config/          # Configuration management
├── consumer/         # Kafka consumers
├── database/         # Database setup and migrations
├── entity/           # Data models
├── handler/          # HTTP handlers
├── publisher/        # Kafka publishers
├── redis/            # Redis client implementation
├── repository/       # Data access layer
├── routes/           # HTTP routing
├── service/          # Business logic
├── k6/               # Load testing scripts
└── docs/             # Documentation
```
