# Prime Directive - Core Principles

## Fundamental Optimization Order

**ALWAYS optimize in this order: Memory → Disk → CPU**

## Core Behaviors

### Analysis Before Action

- **Understand context** before suggesting changes
- **Respect scope** - only modify requested files
- **One task at a time** - focus on specific request
- **Confirm major changes** - provide detailed plan before significant modifications

### Language Requirements

- **ALL code and documentation in ENGLISH only**
- **No exceptions** - French or other languages strictly forbidden
- **Comments, variables, functions** - all must use English

### Error Handling Pattern

```go
// ALWAYS wrap errors with context
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// NEVER ignore errors
// Bad: _ = someFunction()
// Good: if err := someFunction(); err != nil { return err }
```

### Resource Management

```go
// ALWAYS use defer for cleanup
file, err := os.Open(filename)
if err != nil {
    return err
}
defer file.Close()

// ALWAYS use context for cancellation
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

## Decision Framework

### When to Optimize

1. **Hot paths** - code executed frequently (>1000/sec)
2. **Memory pressure** - high allocation rates visible
3. **User complaints** - explicit performance issues mentioned
4. **Production scale** - mentions of "millions of users"

### When NOT to Optimize

1. **Configuration code** - executed once at startup
2. **Test code** - performance not critical
3. **Prototypes/POC** - clarity over performance
4. **Migration scripts** - one-time execution

## Response Template

1. **Acknowledge**: "I see you want to [summary]"
2. **Analyze**: "Here's what I found..."
3. **Propose**: "I suggest these changes..."
4. **Confirm**: "Shall I proceed?"
