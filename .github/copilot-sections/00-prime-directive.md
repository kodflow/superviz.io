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

- **ALL automation scripts created by AI** go in `.tmp/` folder at project root
- **These are helper scripts** for automation, not part of final deliverable
- **Use existing Makefile commands** - never recreate build logic
- **Clean temporary files** after creation - only keep requested files
- **Never create build artifacts** in `.tmp/` - those belong in standard locations

#### .tmp Usage Rules

```
project-root/
├── .tmp/           # AI-created automation scripts only
│   ├── test-runner.sh    # Any test automation script
│   ├── deploy-helper.py  # Any deployment script
│   ├── setup-env.sh      # Any environment setup
│   └── batch-process.js  # Any processing automation
├── Makefile        # Use existing targets
└── dist/           # Build artifacts (standard location)
```

#### What Goes in .tmp/

- **Scripts created to automate tasks** (testing, deployment, processing)
- **Helper utilities** that aren't part of the project deliverable
- **Temporary automation tools** that could be deleted without affecting the project

#### What NEVER Goes in .tmp/

- **Project source code**
- **Build artifacts** (binaries, dist files, compiled assets)
- **Configuration files** needed by the application
- **Documentation** or files requested as deliverables

#### Automation Pattern

```bash
# Create automation script in .tmp
cat > .tmp/run-tests.sh << 'EOF'
#!/bin/bash
make test  # Use existing Makefile targets
EOF
chmod +x .tmp/run-tests.sh

# Execute and clean up if temporary
./.tmp/run-tests.sh
# rm .tmp/run-tests.sh  # Only if not requested to keep
```

#### Integration with Existing Build System

- **ALWAYS check for Makefile** before creating custom logic
- **Use `make test`, `make build`, `make deploy`** instead of custom commands
- **Extend Makefile** if new targets needed, don't bypass it

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
