AI GENERATED DOC FOR BETTER UNDERSTANDING


# Transaction Monitoring System - Technical Documentation

> **For Java Developers**: This document explains the Go-based transaction monitoring system using Java terminology and concepts you're already familiar with.

## Table of Contents
1. [System Overview](#system-overview)
2. [Architecture](#architecture)
3. [Backend Deep Dive](#backend-deep-dive)
4. [Frontend Deep Dive](#frontend-deep-dive)
5. [Data Flow](#data-flow)
6. [Fraud Detection Algorithms](#fraud-detection-algorithms)
7. [Deployment](#deployment)

---

## System Overview

This is a **real-time fraud detection system** for financial transactions. Think of it as a Spring Boot application, but written in Go for better concurrency and performance.

### Tech Stack Translation (Java → This Project)
| Java Ecosystem | This Project |
|----------------|--------------|
| Spring Boot | Go + Gorilla Mux |
| Spring WebSocket | Gorilla WebSocket |
| JPA/Hibernate | database/sql (raw SQL) |
| Jedis/Lettuce | go-redis |
| React (same) | React + Tailwind CSS |
| Maven/Gradle | Go modules |
| Tomcat | Go's built-in HTTP server |

---

## Architecture

### High-Level Architecture
```
┌─────────────┐     WebSocket      ┌──────────────┐
│   React     │ ←─────────────────→ │   Go Backend │
│  Frontend   │     REST API        │   (Port 8080)│
└─────────────┘                     └──────┬───────┘
                                           │
                    ┌──────────────────────┼──────────────────┐
                    │                      │                  │
              ┌─────▼─────┐         ┌─────▼─────┐     ┌─────▼─────┐
              │ PostgreSQL│         │   Redis   │     │ WebSocket │
              │ (Persist) │         │  (Cache)  │     │    Hub    │
              └───────────┘         └───────────┘     └───────────┘
```

### Component Breakdown

**Backend (Go)**: Like a Spring Boot microservice
- **HTTP Server**: Similar to `@RestController` in Spring
- **WebSocket Hub**: Like Spring's `@MessageMapping` for real-time updates
- **Database Layer**: Similar to JPA repositories (but manual SQL)
- **Redis Client**: Like Spring Data Redis for caching

**Frontend (React)**: Standard React SPA
- **WebSocket Client**: Connects to backend for real-time updates
- **REST Client**: Fetches initial data via HTTP

---

## Backend Deep Dive

### Project Structure
```
backend/
├── main.go              # Entry point (like Application.java with @SpringBootApplication)
├── models.go            # DTOs/Entities (like @Entity classes)
├── database.go          # DAO layer (like @Repository)
├── redis_client.go      # Redis operations (like RedisTemplate)
├── handlers.go          # Controllers (like @RestController)
├── websocket.go         # WebSocket server (like @MessageMapping)
├── fraud_detector.go    # Business logic (like @Service)
└── go.mod               # Dependencies (like pom.xml or build.gradle)
```

### 1. main.go - Application Entry Point

**Java Equivalent**: `Application.java` with `@SpringBootApplication`

```go
func main() {
    // Initialize Redis (like @Autowired RedisTemplate)
    InitRedis()
    
    // Initialize Database (like @Autowired DataSource)
    InitDB("postgres://...")
    
    // Create router (like Spring's DispatcherServlet)
    r := mux.NewRouter()
    
    // Start WebSocket Hub (background thread)
    go hub.run()  // Like @Async or ExecutorService
    
    // Register endpoints (like @GetMapping, @PostMapping)
    r.HandleFunc("/ws", serveWs)
    r.HandleFunc("/api/transactions", GetTransactions).Methods("GET")
    r.HandleFunc("/api/simulate", SimulateTransaction).Methods("POST")
    
    // Start server (like embedded Tomcat)
    http.ListenAndServe(":8080", handler)
}
```

**Key Differences from Java**:
- `go hub.run()` starts a goroutine (lightweight thread) - like `CompletableFuture.runAsync()` but more efficient
- No annotations - routing is explicit
- Single binary deployment (no WAR/JAR with embedded server)

### 2. models.go - Data Models

**Java Equivalent**: `@Entity` classes

```go
type Transaction struct {
    ID              int       `json:"id"`              // Like @Id @GeneratedValue
    UserID          string    `json:"user_id"`         // Like @Column
    Amount          float64   `json:"amount"`
    CardNumber      string    `json:"card_number"`
    MerchantDetails string    `json:"merchant_details"`
    Location        string    `json:"location"`
    IsFraud         bool      `json:"is_fraud"`
    CreatedAt       time.Time `json:"created_at"`      // Like @CreatedDate
}
```

**Key Differences**:
- Struct tags (`` `json:"id"` ``) are like Jackson's `@JsonProperty`
- No ORM - you write SQL manually
- Exported fields (capitalized) are like `public` in Java

### 3. database.go - Data Access Layer

**Java Equivalent**: `@Repository` interface with JPA

```go
var db *sql.DB  // Like @Autowired EntityManager

func CreateTransaction(t *Transaction) (*Transaction, error) {
    query := `INSERT INTO transactions (...) VALUES ($1, $2, ...) RETURNING id`
    err := db.QueryRow(query, t.UserID, t.Amount, ...).Scan(&t.ID)
    return t, err
}
```

**Java Comparison**:
```java
// Java (Spring Data JPA)
@Repository
public interface TransactionRepository extends JpaRepository<Transaction, Integer> {
    Transaction save(Transaction t);
}

// Go equivalent is manual SQL
```

**Key Functions**:
- `CreateTransaction()`: Like `repository.save()`
- `GetRecentTransactions()`: Like `repository.findTop50ByOrderByCreatedAtDesc()`
- `MarkTransactionAsFraud()`: Like `repository.updateIsFraudById()`

### 4. redis_client.go - Caching Layer

**Java Equivalent**: `RedisTemplate` in Spring Data Redis

```go
var redisClient *redis.Client  // Like @Autowired RedisTemplate

func IncrementVelocity(userID string) (int64, error) {
    key := fmt.Sprintf("velocity:%s", userID)
    count, err := redisClient.Incr(ctx, key).Result()
    
    if count == 1 {
        redisClient.Expire(ctx, key, time.Minute)  // TTL
    }
    return count, nil
}
```

**Java Comparison**:
```java
// Java (Spring Data Redis)
@Autowired
private RedisTemplate<String, String> redisTemplate;

public Long incrementVelocity(String userId) {
    String key = "velocity:" + userId;
    Long count = redisTemplate.opsForValue().increment(key);
    if (count == 1) {
        redisTemplate.expire(key, 1, TimeUnit.MINUTES);
    }
    return count;
}
```

**Key Operations**:
- `Incr()`: Atomic increment (for velocity tracking)
- `LPush()` + `LTrim()`: Maintain sliding window of recent amounts
- `LRange()`: Retrieve recent amounts for Z-Score calculation

### 5. handlers.go - REST Controllers

**Java Equivalent**: `@RestController` with `@PostMapping`

```go
func SimulateTransaction(w http.ResponseWriter, r *http.Request) {
    var t Transaction
    json.NewDecoder(r.Body).Decode(&t)  // Like @RequestBody
    
    t.CreatedAt = time.Now()
    t.IsFraud = false
    
    CreateTransaction(&t)  // Save to DB
    
    go RunFraudChecks(t)  // Async processing (like @Async)
    
    // Broadcast to WebSocket clients
    hub.Broadcast <- WebSocketMessage{
        Type:    "new_transaction",
        Payload: t,
    }
    
    json.NewEncoder(w).Encode(...)  // Like @ResponseBody
}
```

**Java Comparison**:
```java
@RestController
@RequestMapping("/api")
public class TransactionController {
    
    @PostMapping("/simulate")
    public ResponseEntity<?> simulateTransaction(@RequestBody Transaction t) {
        t.setCreatedAt(LocalDateTime.now());
        t.setFraud(false);
        
        transactionRepository.save(t);
        
        // Async fraud check
        CompletableFuture.runAsync(() -> fraudService.runChecks(t));
        
        // Broadcast via WebSocket
        messagingTemplate.convertAndSend("/topic/transactions", t);
        
        return ResponseEntity.ok(Map.of("status", "success"));
    }
}
```

### 6. websocket.go - Real-Time Communication

**Java Equivalent**: Spring WebSocket with STOMP

**Hub Pattern** (like a message broker):
```go
type Hub struct {
    clients    map[*Client]bool           // Connected clients
    Broadcast  chan WebSocketMessage      // Message queue
    register   chan *Client               // New connections
    unregister chan *Client               // Disconnections
}

func (h *Hub) run() {
    for {
        select {
        case client := <-h.register:
            h.clients[client] = true
        case message := <-h.Broadcast:
            // Send to all connected clients
            for client := range h.clients {
                client.send <- message
            }
        }
    }
}
```

**Java Comparison**:
```java
// Java (Spring WebSocket)
@Configuration
@EnableWebSocketMessageBroker
public class WebSocketConfig implements WebSocketMessageBrokerConfigurer {
    
    @Override
    public void configureMessageBroker(MessageBrokerRegistry config) {
        config.enableSimpleBroker("/topic");  // Like Hub.Broadcast
    }
}

// Broadcasting
@Autowired
private SimpMessagingTemplate messagingTemplate;

messagingTemplate.convertAndSend("/topic/transactions", message);
```

**Client Management**:
- `readPump()`: Listens for client messages (keeps connection alive)
- `writePump()`: Sends messages to client + periodic pings
- Goroutines handle each client concurrently (like virtual threads in Java 21)

### 7. fraud_detector.go - Business Logic

**Java Equivalent**: `@Service` class with business logic

```go
type FraudRule func(t Transaction) (bool, string)  // Like Function<Transaction, Boolean>

func RunFraudChecks(t Transaction) {
    rules := []FraudRule{
        CheckVelocity,
        CheckHighValue,
        CheckGeographicInconsistency,
        CheckZScore,
    }
    
    // Add to history asynchronously
    go AddTransactionAmount(t.UserID, t.Amount)
    
    // Run each rule
    for _, rule := range rules {
        isFraud, ruleName := rule(t)
        if isFraud {
            MarkTransactionAsFraud(t.ID)
            
            // Broadcast fraud alert
            hub.Broadcast <- WebSocketMessage{
                Type: "fraud_alert",
                Payload: map[string]interface{}{
                    "rule_violated": ruleName,
                    "transaction": t,
                },
            }
            return  // Stop on first fraud detection
        }
    }
}
```

**Java Comparison**:
```java
@Service
public class FraudDetectorService {
    
    private List<FraudRule> rules = List.of(
        this::checkVelocity,
        this::checkHighValue,
        this::checkGeographic,
        this::checkZScore
    );
    
    @Async
    public void runFraudChecks(Transaction t) {
        CompletableFuture.runAsync(() -> addToHistory(t));
        
        for (FraudRule rule : rules) {
            if (rule.check(t)) {
                markAsFraud(t.getId());
                broadcastAlert(t);
                return;
            }
        }
    }
}
```

---

## Fraud Detection Algorithms

### 1. Velocity Check (Redis-based)
**Purpose**: Detect rapid-fire transactions (e.g., stolen card)

```go
func CheckVelocity(t Transaction) (bool, string) {
    count, _ := IncrementVelocity(t.UserID)  // Atomic increment in Redis
    return count > 5, "Velocity Check (Redis)"
}
```

**How it works**:
- Redis key: `velocity:user_123`
- Increments on each transaction
- TTL of 1 minute (auto-expires)
- Flags if > 5 transactions/minute

**Java Equivalent**: Using Redis with Spring
```java
Long count = redisTemplate.opsForValue().increment("velocity:" + userId);
if (count == 1) {
    redisTemplate.expire("velocity:" + userId, 1, TimeUnit.MINUTES);
}
return count > 5;
```

### 2. High Value Check
**Purpose**: Flag unusually large transactions

```go
func CheckHighValue(t Transaction) (bool, string) {
    const highValueThreshold = 1500.00
    return t.Amount > highValueThreshold, "High Value Transaction"
}
```

Simple threshold-based rule.

### 3. Geographic Inconsistency
**Purpose**: Detect impossible travel (e.g., NYC → London in 5 minutes)

```go
func CheckGeographicInconsistency(t Transaction) (bool, string) {
    sixtySecondsAgo := time.Now().Add(-60 * time.Second)
    recentLocations, _ := GetRecentTransactionLocations(t.UserID, sixtySecondsAgo)
    
    for _, loc := range recentLocations {
        if loc != "" && t.Location != loc {
            return true, "Geographic Inconsistency"
        }
    }
    return false, ""
}
```

**How it works**:
- Queries PostgreSQL for transactions in last 60 seconds
- If location changed → fraud
- Simple but effective for demo purposes

### 4. Z-Score Anomaly Detection (Statistical)
**Purpose**: Detect transactions that deviate significantly from user's normal behavior

```go
func CheckZScore(t Transaction) (bool, string) {
    amounts, _ := GetRecentAmounts(t.UserID)  // From Redis
    if len(amounts) < 5 {
        return false, ""  // Need baseline
    }
    
    // Calculate mean
    mean := sum(amounts) / len(amounts)
    
    // Calculate standard deviation
    variance := sum((amount - mean)² for amount in amounts) / len(amounts)
    stdDev := sqrt(variance)
    
    // Calculate Z-Score
    zScore := (t.Amount - mean) / stdDev
    
    // Flag if > 3 standard deviations (99.7th percentile)
    return abs(zScore) > 3, fmt.Sprintf("Statistical Anomaly (Z-Score: %.2f)", zScore)
}
```

**Example**:
- User normally spends $50-$100
- Suddenly spends $500
- Z-Score = (500 - 75) / 20 = 21.25 → **FRAUD**

**Java Equivalent**:
```java
public boolean checkZScore(Transaction t) {
    List<Double> amounts = getRecentAmounts(t.getUserId());
    if (amounts.size() < 5) return false;
    
    double mean = amounts.stream().mapToDouble(a -> a).average().orElse(0);
    double variance = amounts.stream()
        .mapToDouble(a -> Math.pow(a - mean, 2))
        .average().orElse(0);
    double stdDev = Math.sqrt(variance);
    
    double zScore = (t.getAmount() - mean) / stdDev;
    return Math.abs(zScore) > 3;
}
```

---

## Frontend Deep Dive

### React Component Structure
```
frontend/src/
├── App.jsx          # Main component (like App.tsx in Angular)
├── index.js         # Entry point (like main.tsx)
└── index.css        # Tailwind styles
```

### Key Frontend Concepts

#### 1. WebSocket Connection
```javascript
useEffect(() => {
    ws.current = new WebSocket('ws://localhost:8080/ws');
    
    ws.current.onmessage = (event) => {
        const message = JSON.parse(event.data);
        
        if (message.type === 'new_transaction') {
            setTransactions(prev => [message.payload, ...prev]);
        } else if (message.type === 'fraud_alert') {
            setTransactions(prev => [
                {...message.payload.transaction, fraud_reason: message.payload.rule_violated},
                ...prev
            ]);
        }
    };
}, []);
```

**Java Equivalent** (Spring WebSocket Client):
```java
@Component
public class WebSocketClient {
    
    @EventListener
    public void handleWebSocketMessage(WebSocketMessage message) {
        if ("new_transaction".equals(message.getType())) {
            transactions.add(message.getPayload());
        } else if ("fraud_alert".equals(message.getType())) {
            // Handle fraud alert
        }
    }
}
```

#### 2. Sticky Location Logic (Simulation)
```javascript
const userLocations = useRef({});  // Like a HashMap<String, Location>

const simulateTransaction = () => {
    const userId = `user_${Math.floor(Math.random() * 10)}`;
    
    // Sticky location (user stays in same city)
    let location;
    if (!userLocations.current[userId] || Math.random() < 0.05) {
        location = randomLocation();
        userLocations.current[userId] = location;  // Update cache
    } else {
        location = userLocations.current[userId];  // Reuse
    }
    
    // Send to backend
    fetch('http://localhost:8080/api/simulate', {
        method: 'POST',
        body: JSON.stringify({user_id: userId, location: location.name, ...})
    });
};
```

**Why Sticky Locations?**
- Without: Every transaction is random location → 70%+ fraud rate
- With: Users stay in one city → realistic 5-10% fraud rate

---

## Data Flow

### Transaction Flow (End-to-End)
```
1. User clicks "Simulate" button
   ↓
2. Frontend sends POST /api/simulate
   ↓
3. Backend (handlers.go):
   - Saves to PostgreSQL
   - Broadcasts "new_transaction" via WebSocket
   - Starts fraud check (async goroutine)
   ↓
4. Fraud Detector (fraud_detector.go):
   - Runs 4 rules sequentially
   - If fraud detected:
     * Updates PostgreSQL (is_fraud = true)
     * Broadcasts "fraud_alert" via WebSocket
   ↓
5. Frontend receives WebSocket message:
   - "new_transaction" → Add to feed
   - "fraud_alert" → Add to feed with red border + reason
   ↓
6. User sees real-time update on dashboard
```

### WebSocket Message Format
```json
{
  "type": "new_transaction",
  "payload": {
    "id": 123,
    "user_id": "user_5",
    "amount": 450.50,
    "location": "New York",
    "is_fraud": false,
    "created_at": "2025-11-19T18:30:00Z"
  }
}

{
  "type": "fraud_alert",
  "payload": {
    "rule_violated": "Geographic Inconsistency",
    "transaction": { /* same as above */ }
  }
}
```

---

## Deployment

### Docker Compose Architecture
```yaml
services:
  backend:
    build: ./backend
    ports: ["8080:8080"]
    depends_on: [postgres, redis]
    
  frontend:
    build: ./frontend
    ports: ["3000:3000"]
    
  postgres:
    image: postgres:15-alpine
    ports: ["5432:5432"]
    
  redis:
    image: redis:7-alpine
    ports: ["6379:6379"]
```

**Java Equivalent**: Like a `docker-compose.yml` with Spring Boot + PostgreSQL + Redis

### Running the System
```bash
# Build and start all services
docker-compose up --build

# Access frontend
http://localhost:3000

# Access backend API
http://localhost:8080/api/transactions
```

---

## Key Takeaways for Java Developers

### Go vs Java Paradigms

| Concept | Java | Go |
|---------|------|-----|
| Concurrency | Threads, ExecutorService | Goroutines (lightweight) |
| Dependency Injection | @Autowired, Spring | Manual (pass dependencies) |
| ORM | Hibernate, JPA | Manual SQL |
| Error Handling | Exceptions | Return `(value, error)` |
| Generics | Full support | Limited (improving) |
| Package Management | Maven/Gradle | Go modules |
| Compilation | JVM bytecode | Native binary |

### Performance Benefits
- **Goroutines**: 1000x lighter than threads (2KB vs 2MB stack)
- **Channels**: Built-in message passing (safer than shared memory)
- **Native Compilation**: No JVM startup time
- **Garbage Collection**: Optimized for low latency

### When to Use Go vs Java
**Use Go**:
- High-concurrency systems (WebSocket servers, proxies)
- Microservices with simple logic
- CLI tools, DevOps tooling

**Use Java**:
- Complex business logic (Spring ecosystem)
- Enterprise applications (mature libraries)
- Android development

---

## Further Reading

- [Go by Example](https://gobyexample.com/) - Quick Go syntax reference
- [Effective Go](https://go.dev/doc/effective_go) - Best practices
- [Gorilla WebSocket](https://github.com/gorilla/websocket) - WebSocket library docs
- [Go Redis](https://redis.uptrace.dev/) - Redis client docs

---
