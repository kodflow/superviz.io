#!/bin/bash

# Script to generate copilot-instructions.md from sectioned files
# Usage: ./scripts/generate-copilot.sh

set -euo pipefail

# Configuration
SECTIONS_DIR=".github/copilot-sections"
OUTPUT_FILE=".github/copilot-instructions.md"
TEMP_FILE=".github/copilot-instructions.tmp"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to validate prerequisites
validate_prerequisites() {
    print_status "Validating prerequisites..."
    
    if [ ! -d "$SECTIONS_DIR" ]; then
        print_error "Sections directory not found: $SECTIONS_DIR"
        exit 1
    fi
    
    local section_files=(
        "01-prime-directive.md"
        "02-go-1-24-performance-standards.md"
        "03-documentation-format.md"
        "04-optimization-workflow.md"
        "05-cpu-optimization.md"
        "06-disk-optimization.md"
        "07-memory-optimization.md"
        "08-code-quality.md"
        "09-test-coverage.md"
        "10-review-checklist.md"
        "11-summary.md"
        "12-go-design-patterns.md"
        "13-production-scale-patterns.md"
    )
    
    for file in "${section_files[@]}"; do
        if [ ! -f "$SECTIONS_DIR/$file" ]; then
            print_error "Section file not found: $SECTIONS_DIR/$file"
            exit 1
        fi
    done
    
    print_success "All section files found"
}

# Function to process a section file
process_section() {
    local file="$1"
    local is_first_section="$2"
    
    print_status "Processing section: $file"
    
    # Read the file content
    local content
    content=$(cat "$SECTIONS_DIR/$file")
    
    # For the first section, keep the title as H1, otherwise convert ## to ##
    if [ "$is_first_section" = "true" ]; then
        # Convert the first ## to #
        content=$(echo "$content" | sed '1s/^## /# /')
    fi
    
    echo "$content"
}

# Function to generate the final file
generate_copilot_instructions() {
    print_status "Generating copilot instructions file..."
    
    # Remove temp file if it exists
    rm -f "$TEMP_FILE"
    
    # Add the instructions wrapper
    echo '````instructions' > "$TEMP_FILE"
    
    # Process each section
    local section_files=(
        "01-prime-directive.md"
        "02-go-1-24-performance-standards.md"
        "03-documentation-format.md"
        "04-optimization-workflow.md"
        "05-cpu-optimization.md"
        "06-disk-optimization.md"
        "07-memory-optimization.md"
        "08-code-quality.md"
        "09-test-coverage.md"
        "10-review-checklist.md"
        "11-summary.md"
        "12-go-design-patterns.md"
        "13-production-scale-patterns.md"
    )
    
    local is_first=true
    for file in "${section_files[@]}"; do
        print_status "Adding section: $file"
        
        # Add separator between sections (except before first)
        if [ "$is_first" != "true" ]; then
            echo "" >> "$TEMP_FILE"
            echo "---" >> "$TEMP_FILE"
            echo "" >> "$TEMP_FILE"
        fi
        
        # Process and add the section
        process_section "$file" "$is_first" >> "$TEMP_FILE"
        
        is_first=false
    done
    
    # Close the instructions wrapper
    echo '````' >> "$TEMP_FILE"
    
    # Move temp file to final location
    mv "$TEMP_FILE" "$OUTPUT_FILE"
    
    print_success "Generated $OUTPUT_FILE"
}

# Function to validate the generated file
validate_generated_file() {
    print_status "Validating generated file..."
    
    if [ ! -f "$OUTPUT_FILE" ]; then
        print_error "Generated file not found: $OUTPUT_FILE"
        exit 1
    fi
    
    # Check file size
    local file_size
    file_size=$(wc -l < "$OUTPUT_FILE")
    if [ "$file_size" -lt 100 ]; then
        print_error "Generated file seems too small: $file_size lines"
        exit 1
    fi
    
    # Check for proper instruction wrapper
    if ! head -1 "$OUTPUT_FILE" | grep -q '````instructions'; then
        print_error "File does not start with proper instruction wrapper"
        exit 1
    fi
    
    if ! tail -1 "$OUTPUT_FILE" | grep -q '````'; then
        print_error "File does not end with proper instruction wrapper"
        exit 1
    fi
    
    print_success "Generated file validation passed ($file_size lines)"
}

# Function to show summary
show_summary() {
    print_success "Copilot instructions generation completed!"
    echo ""
    echo "📁 Sections directory: $SECTIONS_DIR"
    echo "📄 Generated file: $OUTPUT_FILE"
    echo "📊 File size: $(wc -l < "$OUTPUT_FILE") lines"
    echo ""
    echo "To regenerate: make generate-copilot"
}

# Main execution
main() {
    print_status "Starting copilot instructions generation..."
    
    validate_prerequisites
    generate_copilot_instructions
    validate_generated_file
    show_summary
}

# Execute main function
main "$@"
