# Primitive Rules - Universal Core Principles

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

### Build and Test File Management

- **ALL build artifacts for testing** must be placed in `.tmp/` folder at project root
- **Test builds, prototypes, experiments** go in `.tmp/`
- **Never pollute main directories** with temporary build files
- **Clean `.tmp/` regularly** - it's meant to be disposable

#### .tmp Directory Structure

```
project-root/
├── .tmp/           # All temporary builds and test files
│   ├── builds/     # Test builds
│   ├── tests/      # Test artifacts
│   └── experiments/ # Prototype files
├── src/            # Source code (keep clean)
└── dist/           # Production builds only
```

#### Implementation Pattern for .tmp Usage

```bash
# Always create .tmp if it doesn't exist
mkdir -p .tmp/builds
```

```python
# Python - use .tmp for temporary outputs
import os
os.makedirs('.tmp/tests', exist_ok=True)
```

### Universal Error Handling Pattern

```bash
# Bash/Shell - ALWAYS check return codes
command || { echo "Command failed"; exit 1; }
```

```python
# Python - ALWAYS capture exceptions
try:
    operation()
except Exception as e:
    logger.error(f"Operation failed: {e}")
    raise
```

```javascript
// JavaScript - ALWAYS handle errors
try {
  await operation();
} catch (error) {
  console.error("Operation failed:", error);
  throw error;
}
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
