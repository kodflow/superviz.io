"""Starlark rules for dynamic build metadata injection."""

load("@bazel_skylib//lib:shell.bzl", "shell")

def _get_build_vars_impl(ctx):
    """Implementation for getting build variables."""
    
    # Get the workspace status files that contain stamping info
    stamp_inputs = [ctx.info_file, ctx.version_file]
    
    # Create output file
    out = ctx.actions.declare_file("build_vars.bzl")
    
    # Create script to extract variables
    script_template = """#!/bin/bash
set -e

# Default values
VERSION="dev"
COMMIT="unknown"
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILT_BY="bazel"

# Try to get values from git if available
if command -v git >/dev/null 2>&1 && [ -d .git ]; then
    COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    DATE=$(git log -1 --format="%%aI" 2>/dev/null || date -u +"%Y-%m-%dT%H:%M:%SZ")
fi

# Try to get user info
if [ -n "$USER" ]; then
    BUILT_BY="$USER"
elif [ -n "$USERNAME" ]; then
    BUILT_BY="$USERNAME"
fi

# Read from stamp files if they exist
if [ -f "{version_file}" ]; then
    while IFS=' ' read -r key value; do
        case "$key" in
            "STABLE_BUILD_SCM_REVISION")
                if [ "$value" != "" ] && [ "$value" != "unknown" ]; then
                    COMMIT="$value"
                fi
                ;;
            "BUILD_TIMESTAMP")
                if [ "$value" != "" ] && [ "$value" != "0" ]; then
                    DATE=$(date -u -d "@$value" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "$DATE")
                fi
                ;;
            "BUILD_USER")
                if [ "$value" != "" ]; then
                    BUILT_BY="$value"
                fi
                ;;
        esac
    done < "{version_file}"
fi

if [ -f "{info_file}" ]; then
    while IFS=' ' read -r key value; do
        case "$key" in
            "BUILD_SCM_REVISION")
                if [ "$value" != "" ] && [ "$value" != "unknown" ] && [ "$COMMIT" = "unknown" ]; then
                    COMMIT="$value"
                fi
                ;;
        esac
    done < "{info_file}"
fi

# Generate the Starlark file
cat > "{output}" << 'EOF'
"""Auto-generated build variables."""

BUILD_VARS = {{
    "version": "{version}",
    "commit": "$COMMIT",
    "date": "$DATE",
    "builtBy": "$BUILT_BY",
}}
EOF
"""
    
    script = script_template.format(
        version_file = ctx.version_file.path,
        info_file = ctx.info_file.path,
        output = out.path,
        version = ctx.attr.version,
    )
    
    ctx.actions.run_shell(
        inputs = stamp_inputs,
        outputs = [out],
        command = script,
        mnemonic = "GenerateBuildVars",
        use_default_shell_env = True,
    )
    
    return [DefaultInfo(files = depset([out]))]

get_build_vars = rule(
    implementation = _get_build_vars_impl,
    attrs = {
        "version": attr.string(default = "dev"),
    },
    doc = "Generates build variables from git and stamping info",
)

def _generate_x_defs_impl(ctx):
    """Generate x_defs for go_binary."""
    
    build_vars_file = ctx.file.build_vars
    out = ctx.actions.declare_file("x_defs.bzl")
    
    script_template = """
# Source the build vars
source {build_vars}

# Generate x_defs
cat > {output} << 'EOF'
"""Auto-generated x_defs for go_binary."""

def get_x_defs(os_arch):
    return {{
        "github.com/kodflow/superviz.io/internal/providers.version": BUILD_VARS["version"],
        "github.com/kodflow/superviz.io/internal/providers.commit": BUILD_VARS["commit"],
        "github.com/kodflow/superviz.io/internal/providers.date": BUILD_VARS["date"],
        "github.com/kodflow/superviz.io/internal/providers.builtBy": BUILD_VARS["builtBy"],
        "github.com/kodflow/superviz.io/internal/providers.osArch": os_arch,
    }}
EOF
"""
    
    script = script_template.format(
        build_vars = build_vars_file.path,
        output = out.path,
    )
    
    ctx.actions.run_shell(
        inputs = [build_vars_file],
        outputs = [out],
        command = script,
        mnemonic = "GenerateXDefs",
    )
    
    return [DefaultInfo(files = depset([out]))]

generate_x_defs = rule(
    implementation = _generate_x_defs_impl,
    attrs = {
        "build_vars": attr.label(allow_single_file = True),
    },
)
