"""Test Configuration Macros for Bazel

This module provides optimized test configurations and macros for Go tests
with performance optimizations, caching, and parallel execution.
"""

load("@io_bazel_rules_go//go:def.bzl", "go_test")

def optimized_go_test(name, **kwargs):
    """Optimized go_test macro with common settings.

    Args:
        name: Name of the test target
        **kwargs: Additional arguments passed to go_test
    """
    size = kwargs.get("size", "small")
    timeout = kwargs.get("timeout", "short")

    # Default tags for better test organization
    tags = kwargs.get("tags", [])
    if "unit" not in tags:
        tags.append("unit")

    # Performance optimizations
    if size == "small":
        tags.append("fast")

    # Test environment optimizations
    env = kwargs.get("env", {})
    env.update({
        "GO111MODULE": "on",
        "GOCACHE": "/tmp/gocache",
        "GOMODCACHE": "/tmp/gomodcache",
    })

    # Memory optimization for small tests
    if size == "small":
        env["GOMAXPROCS"] = "2"

    go_test(
        name = name,
        size = size,
        timeout = timeout,
        tags = tags,
        env = env,
        **{k: v for k, v in kwargs.items() if k not in ["size", "timeout", "tags", "env"]}
    )

def parallel_go_test(name, shard_count = 4, **kwargs):
    """Parallel go_test with sharding for large test suites.

    Args:
        name: Name of the test target
        shard_count: Number of shards for parallel execution
        **kwargs: Additional arguments passed to optimized_go_test
    """
    optimized_go_test(
        name = name,
        shard_count = shard_count,
        size = kwargs.get("size", "medium"),
        tags = kwargs.get("tags", []) + ["parallel"],
        **{k: v for k, v in kwargs.items() if k not in ["shard_count"]}
    )

def coverage_go_test(name, **kwargs):
    """go_test optimized for coverage collection.

    Args:
        name: Name of the test target
        **kwargs: Additional arguments passed to optimized_go_test
    """
    tags = kwargs.get("tags", [])
    tags.append("coverage")

    optimized_go_test(
        name = name,
        tags = tags,
        **kwargs
    )
