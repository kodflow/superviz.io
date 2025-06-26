## Go Design Patterns Documentation Standards

### The 5 Essential Patterns for Real Go Projects

Based on production-tested patterns, here are the design patterns that matter most in Go development, documented according to our Godoc standards.

---

### 1. Strategy Pattern - Clean Conditional Logic

Replace 50-line switch-case statements with clean interfaces.

```go
// ShippingStrategy Interface for calculating shipping costs by region
type ShippingStrategy interface {
    CalculateCost(weight float64) float64 // Calculate cost based on weight
}

// IndiaShipping Shipping cost calculation for India
type IndiaShipping struct{}

// CalculateCost Calculate shipping cost for India
// Code block:
//
//  shipping := &IndiaShipping{}
//  cost := shipping.CalculateCost(2.5)
//  fmt.Printf("Cost: %.2f\n", cost)
//
// Parameters:
//   - 1 weight: float64 - package weight in kg (must be positive)
//
// Returns:
//   - 1 cost: float64 - shipping cost in local currency
func (s IndiaShipping) CalculateCost(weight float64) float64 {
    return 50 + weight*10
}

// ShippingCalculator Calculator using appropriate strategy
type ShippingCalculator struct {
    strategy ShippingStrategy // Active calculation strategy
}

// NewShippingCalculator Create calculator with given strategy
// Code block:
//
//  calculator := NewShippingCalculator(&IndiaShipping{})
//  cost := calculator.Calculate(2.5)
//  fmt.Printf("Total: %.2f\n", cost)
//
// Parameters:
//   - 1 strategy: ShippingStrategy - calculation strategy (cannot be nil)
//
// Returns:
//   - 1 calculator: *ShippingCalculator - configured calculator
func NewShippingCalculator(strategy ShippingStrategy) *ShippingCalculator {
    return &ShippingCalculator{strategy: strategy}
}
```

---

### 2. Factory Pattern - Flexible Object Creation

Create different object types based on context.

```go
// DatabaseType Supported database type
type DatabaseType string

const (
    MySQL    DatabaseType = "mysql"    // MySQL database
    Postgres DatabaseType = "postgres" // PostgreSQL database
)

// Database Interface for database operations
type Database interface {
    Connect() error    // Establish connection
    Query(sql string)  // Execute query
    Close() error      // Close connection
}

// NewDatabase Factory to create database instance
// Code block:
//
//  db, err := NewDatabase(MySQL, "localhost:3306")
//  if err != nil {
//      log.Fatal(err)
//  }
//  defer db.Close()
//
// Parameters:
//   - 1 dbType: DatabaseType - database type to create
//   - 2 config: string - connection configuration (cannot be empty)
//
// Returns:
//   - 1 db: Database - configured database instance
//   - 2 error - nil if creation successful, error if unsupported type
func NewDatabase(dbType DatabaseType, config string) (Database, error) {
    switch dbType {
    case MySQL:
        return &MySQLDB{host: config, port: 3306}, nil
    case Postgres:
        return &PostgresDB{connectionString: config}, nil
    default:
        return nil, fmt.Errorf("unsupported database type: %s", dbType)
    }
}
```

---

### 3. Builder Pattern - Complex Object Construction

Build objects with many optional parameters.

```go
// ServerConfig Complete server configuration
type ServerConfig struct {
    host         string        // Listen address
    port         int           // Listen port
    timeout      time.Duration // Request timeout
    enableHTTPS  bool          // HTTPS activation
}

// ServerBuilder Builder for constructing server configuration
type ServerBuilder struct {
    config *ServerConfig // Configuration being built
}

// NewServerBuilder Create new builder with default values
// Code block:
//
//  builder := NewServerBuilder()
//  config := builder.Host("localhost").Port(8080).Build()
//  fmt.Printf("Server: %s:%d\n", config.host, config.port)
//
// Returns:
//   - 1 builder: *ServerBuilder - builder with default configuration
func NewServerBuilder() *ServerBuilder {
    return &ServerBuilder{
        config: &ServerConfig{
            host:    "localhost",
            port:    8080,
            timeout: 30 * time.Second,
        },
    }
}

// Host Configure listen address
// Code block:
//
//  config := NewServerBuilder().Host("0.0.0.0").Port(9000).Build()
//
// Parameters:
//   - 1 host: string - listen address (cannot be empty)
//
// Returns:
//   - 1 builder: *ServerBuilder - builder for chaining
func (b *ServerBuilder) Host(host string) *ServerBuilder {
    b.config.host = host
    return b
}

// Build Construct final configuration
// Code block:
//
//  config := NewServerBuilder().
//      Host("api.example.com").
//      Port(443).
//      Build()
//
// Returns:
//   - 1 config: *ServerConfig - complete server configuration
func (b *ServerBuilder) Build() *ServerConfig {
    return b.config
}
```

---

### 4. Observer Pattern - Event Communication

Notify multiple components on state changes.

```go
// Event Represents system event
type Event struct {
    Type      string      // Event type
    Data      interface{} // Associated data
    Timestamp time.Time   // Event moment
}

// Observer Interface for event observers
type Observer interface {
    HandleEvent(event Event) // Process received event
    GetID() string           // Return unique identifier
}

// EventManager Central event manager
type EventManager struct {
    observers map[string]Observer // Map of registered observers
    mu        sync.RWMutex        // Mutex for concurrent access
}

// NewEventManager Create new event manager
// Code block:
//
//  manager := NewEventManager()
//  observer := NewEmailNotifier("test@example.com")
//  manager.Subscribe(observer)
//
// Returns:
//   - 1 manager: *EventManager - initialized manager
func NewEventManager() *EventManager {
    return &EventManager{
        observers: make(map[string]Observer),
    }
}

// Subscribe Register observer
// Code block:
//
//  manager := NewEventManager()
//  notifier := NewEmailNotifier("admin@example.com")
//  manager.Subscribe(notifier)
//
// Parameters:
//   - 1 observer: Observer - observer to register (cannot be nil)
func (em *EventManager) Subscribe(observer Observer) {
    em.mu.Lock()
    defer em.mu.Unlock()
    em.observers[observer.GetID()] = observer
}

// Publish Publish event to all observers
// Code block:
//
//  event := Event{Type: "order.completed", Data: "Order #123"}
//  manager.Publish(event)
//
// Parameters:
//   - 1 event: Event - event to publish
func (em *EventManager) Publish(event Event) {
    em.mu.RLock()
    defer em.mu.RUnlock()
    
    for _, observer := range em.observers {
        go observer.HandleEvent(event) // Asynchronous notification
    }
}
```

---

### 5. Dependency Injection - Decoupling Components

Decouple components and facilitate testing.

```go
// Logger Interface for logging
type Logger interface {
    Info(message string)  // Log info level
    Error(message string) // Log error level
}

// Repository Interface for persistence
type Repository interface {
    Save(data interface{}) error // Save data
    Find(id string) interface{}  // Find by ID
}

// UserService Business service with injected dependencies
type UserService struct {
    logger Logger     // Injected logger
    repo   Repository // Injected repository
}

// NewUserService Create service with dependency injection
// Code block:
//
//  logger := &ConsoleLogger{}
//  repo := &DatabaseRepository{}
//  service := NewUserService(logger, repo)
//  
//  err := service.CreateUser("John Doe")
//  if err != nil {
//      log.Fatal(err)
//  }
//
// Parameters:
//   - 1 logger: Logger - logger for traces (cannot be nil)
//   - 2 repo: Repository - repository for persistence (cannot be nil)
//
// Returns:
//   - 1 service: *UserService - service configured with dependencies
func NewUserService(logger Logger, repo Repository) *UserService {
    return &UserService{
        logger: logger,
        repo:   repo,
    }
}

// CreateUser Create new user
// Code block:
//
//  err := service.CreateUser("Jane Smith")
//  if err != nil {
//      log.Printf("Creation error: %v", err)
//  }
//
// Parameters:
//   - 1 name: string - user name (cannot be empty)
//
// Returns:
//   - 1 error - nil if creation successful, error otherwise
func (s *UserService) CreateUser(name string) error {
    s.logger.Info(fmt.Sprintf("Creating user: %s", name))
    
    user := map[string]string{"name": name}
    err := s.repo.Save(user)
    if err != nil {
        s.logger.Error(fmt.Sprintf("Save failed: %v", err))
        return err
    }
    
    s.logger.Info("User created successfully")
    return nil
}
```

---

### Pattern Usage in superviz.io

These patterns are particularly useful in superviz.io context:

- **Strategy**: Managing different package managers (apt, yum, apk, etc.)
- **Factory**: Creating SSH clients based on configuration
- **Builder**: Building complex installation configurations
- **Observer**: Installation progress notifications
- **DI**: Injecting infrastructure services into CLI commands

All examples follow our Godoc standards with code blocks, parameters, and returns documentation.
