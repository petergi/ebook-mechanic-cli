#!/bin/bash
# Wiki Sync Script for ebook-mechanic-cli
# Syncs documentation from docs/ to GitHub Wiki

set -e

# ==================== Configuration ====================
WIKI_DIR="wiki"
DOCS_DIR="docs"
WIKI_REMOTE="https://github.com/petergi/ebook-mechanic-cli.wiki.git"

# ==================== Helper Functions ====================
log_info() { echo "[INFO] $1"; }
log_success() { echo "[OK] $1"; }
log_warning() { echo "[WARN] $1"; }
log_error() { echo "[ERROR] $1"; }

clone_wiki() {
    log_info "Cloning wiki repository..."
    if [ -d "$WIKI_DIR" ]; then
        if [ ! -f "$WIKI_DIR/.git/config" ]; then
            log_warning "Wiki directory exists without .git; removing and recloning"
            rm -rf "$WIKI_DIR"
        else
            log_warning "Wiki directory already exists, skipping clone"
            return 0
        fi
    fi

    git clone "$WIKI_REMOTE" "$WIKI_DIR"
    log_success "Wiki repository cloned"
}

sync_docs() {
    log_info "Syncing documentation to wiki..."

    if [ ! -d "$WIKI_DIR" ] || [ ! -d "$WIKI_DIR/.git" ]; then
        log_error "Wiki directory does not exist or is not a git repo. Run './scripts/wiki-sync.sh clone' first."
        exit 1
    fi

    if [ ! -d "$DOCS_DIR" ]; then
        log_error "Docs wiki source does not exist: $DOCS_DIR"
        exit 1
    fi

    log_info "Syncing docs content..."
    rsync -av --delete --exclude ".git" "$DOCS_DIR/" "$WIKI_DIR/"

    # Ensure Home page exists
    if [ -f "$DOCS_DIR/README.md" ]; then
        log_info "Creating Home.md from docs/README.md..."
        cp "$DOCS_DIR/README.md" "$WIKI_DIR/Home.md"
    elif [ ! -f "$WIKI_DIR/Home.md" ] && [ -f "README.md" ]; then
        log_info "Creating Home.md from README.md..."
        cp "README.md" "$WIKI_DIR/Home.md"
    fi

    create_sidebar

    log_success "Documentation sync complete"
}

create_sidebar() {
    log_info "Creating wiki sidebar..."

    cat > "$WIKI_DIR/_Sidebar.md" << 'EOF'
## ebook-mechanic-cli Wiki

### Getting Started
* **[Home](Home)** - Overview
* **[User Guide](USER_GUIDE)** - CLI usage
* **[CLI Reference](CLI_REFERENCE)** - Commands and key bindings
* **[Error Codes](ERROR_CODES)** - Validation and repair codes

### Architecture
* **[Architecture](ARCHITECTURE)** - System design
* **[ADR Index](ADR_INDEX)** - Architecture Decision Records

### Validation and Repair
* **[Specs](SPECS)** - EPUB and PDF specs
* **[Testing](TESTING)** - Test suite and fixtures

### Maintenance
* **[CI and Automation](CI)** - CI/CD and docs checks

---
*Edit this wiki from docs/ and run `make wiki-update`*
EOF

    log_success "Sidebar created"
}

commit_and_push() {
    log_info "Committing and pushing changes to wiki..."

    cd "$WIKI_DIR"

    if ! git diff --quiet || ! git diff --cached --quiet || [ -n "$(git ls-files --others --exclude-standard)" ]; then
        git add .
        git commit -m "Update wiki documentation from repository

Auto-synced from ebook-mechanic-cli repository at $(date -u +"%Y-%m-%d %H:%M:%S UTC")"

        git push origin master
        log_success "Wiki changes pushed successfully"
    else
        log_info "No changes to commit"
    fi

    cd ..
}

pull_latest() {
    log_info "Pulling latest wiki changes..."

    if [ ! -d "$WIKI_DIR" ]; then
        log_error "Wiki directory does not exist. Run './scripts/wiki-sync.sh clone' first."
        exit 1
    fi

    cd "$WIKI_DIR"
    git pull origin master
    cd ..

    log_success "Wiki updated from remote"
}

show_status() {
    log_info "Wiki repository status..."

    if [ ! -d "$WIKI_DIR" ]; then
        log_warning "Wiki directory does not exist. Run './scripts/wiki-sync.sh clone' first."
        exit 0
    fi

    cd "$WIKI_DIR"
    git status
    cd ..
}

clean_wiki() {
    log_warning "Removing wiki directory..."

    if [ -d "$WIKI_DIR" ]; then
        rm -rf "$WIKI_DIR"
        log_success "Wiki directory removed"
    else
        log_info "Wiki directory does not exist"
    fi
}

show_help() {
    cat << EOF
Wiki Sync Script for ebook-mechanic-cli

Usage: $0 <command>

Commands:
  clone       Clone the wiki repository
  sync        Sync documentation to wiki (does not push)
  push        Commit and push changes to wiki
  pull        Pull latest wiki changes
  full        Full sync: clone (if needed) + sync + push
  status      Show wiki git status
  clean       Remove wiki directory
  help        Show this help message

Examples:
  $0 clone              # Clone wiki repository
  $0 sync               # Sync docs to wiki
  $0 push               # Commit and push changes
  $0 full               # Do everything: sync and push

EOF
}

# ==================== Main ====================
case "${1:-help}" in
    clone)
        clone_wiki
        ;;
    sync)
        sync_docs
        ;;
    push)
        commit_and_push
        ;;
    pull)
        pull_latest
        ;;
    full)
        clone_wiki
        sync_docs
        commit_and_push
        ;;
    status)
        show_status
        ;;
    clean)
        clean_wiki
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        log_error "Unknown command: $1"
        show_help
        exit 1
        ;;
esac
