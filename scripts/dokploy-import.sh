#!/bin/bash
#
# dokploy-import.sh - Generate Terraform import blocks for all Dokploy resources
#
# Usage:
#   ./scripts/dokploy-import.sh > imports.tf
#   terraform plan -generate-config-out=generated.tf
#
# Environment Variables (required):
#   DOKPLOY_HOST    - The Dokploy server URL (e.g., https://dokploy.example.com)
#   DOKPLOY_API_KEY - The Dokploy API key
#
# Options:
#   --projects-only    Only generate import blocks for projects
#   --infrastructure   Only generate import blocks for infrastructure (servers, ssh keys, etc.)
#   --help             Show this help message
#

set -e

# Colors for output (if terminal supports it)
if [ -t 1 ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    NC='\033[0m' # No Color
else
    RED=''
    GREEN=''
    YELLOW=''
    NC=''
fi

# Help function
show_help() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Generate Terraform import blocks for all Dokploy resources."
    echo ""
    echo "Required environment variables:"
    echo "  DOKPLOY_HOST      The Dokploy server URL"
    echo "  DOKPLOY_API_KEY   The Dokploy API key"
    echo ""
    echo "Options:"
    echo "  --projects-only     Only generate import blocks for projects and their contents"
    echo "  --infrastructure    Only generate import blocks for infrastructure resources"
    echo "  --dry-run           Show what would be fetched without generating output"
    echo "  --help              Show this help message"
    echo ""
    echo "Example:"
    echo "  export DOKPLOY_HOST='https://dokploy.example.com'"
    echo "  export DOKPLOY_API_KEY='your-api-key'"
    echo "  $0 > imports.tf"
    echo "  terraform plan -generate-config-out=generated.tf"
}

# Check for required environment variables
check_env() {
    if [ -z "$DOKPLOY_HOST" ]; then
        echo -e "${RED}Error: DOKPLOY_HOST environment variable is not set${NC}" >&2
        exit 1
    fi

    if [ -z "$DOKPLOY_API_KEY" ]; then
        echo -e "${RED}Error: DOKPLOY_API_KEY environment variable is not set${NC}" >&2
        exit 1
    fi
}

# Make API request
api_request() {
    local endpoint="$1"
    curl -s -f \
        -H "x-api-key: $DOKPLOY_API_KEY" \
        -H "Accept: application/json" \
        "${DOKPLOY_HOST}/api${endpoint}" 2>/dev/null
}

# Sanitize name for Terraform resource names
sanitize_name() {
    echo "$1" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9]/_/g' | sed 's/__*/_/g' | sed 's/^_//' | sed 's/_$//'
}

# Generate import block
import_block() {
    local resource_type="$1"
    local name="$2"
    local id="$3"

    echo "import {"
    echo "  to = ${resource_type}.${name}"
    echo "  id = \"${id}\""
    echo "}"
    echo ""
}

# Parse options
PROJECTS_ONLY=false
INFRASTRUCTURE_ONLY=false
DRY_RUN=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --projects-only)
            PROJECTS_ONLY=true
            shift
            ;;
        --infrastructure)
            INFRASTRUCTURE_ONLY=true
            shift
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --help|-h)
            show_help
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}" >&2
            show_help
            exit 1
            ;;
    esac
done

# Check environment
check_env

# Header
if [ "$DRY_RUN" = false ]; then
    echo "# Dokploy Import Blocks"
    echo "# Generated on $(date)"
    echo "# Host: $DOKPLOY_HOST"
    echo "#"
    echo "# Usage:"
    echo "#   terraform plan -generate-config-out=generated.tf"
    echo "#   terraform apply"
    echo ""
fi

# Fetch and process infrastructure resources
if [ "$PROJECTS_ONLY" = false ]; then
    # Servers
    echo -e "${GREEN}Fetching servers...${NC}" >&2
    servers=$(api_request "/server.all" || echo "[]")
    if [ "$DRY_RUN" = true ]; then
        echo "Found $(echo "$servers" | jq -r 'length') servers" >&2
    else
        echo "# ============================================================================="
        echo "# Servers"
        echo "# ============================================================================="
        echo ""
        echo "$servers" | jq -r '.[] | "\(.serverId) \(.name)"' | while read -r id name; do
            sname=$(sanitize_name "$name")
            import_block "dokploy_server" "$sname" "$id"
        done
    fi

    # SSH Keys
    echo -e "${GREEN}Fetching SSH keys...${NC}" >&2
    sshkeys=$(api_request "/sshKey.all" || echo "[]")
    if [ "$DRY_RUN" = true ]; then
        echo "Found $(echo "$sshkeys" | jq -r 'length') SSH keys" >&2
    else
        echo "# ============================================================================="
        echo "# SSH Keys"
        echo "# ============================================================================="
        echo ""
        echo "$sshkeys" | jq -r '.[] | "\(.sshKeyId) \(.name)"' | while read -r id name; do
            sname=$(sanitize_name "$name")
            import_block "dokploy_ssh_key" "$sname" "$id"
        done
    fi

    # Registries
    echo -e "${GREEN}Fetching registries...${NC}" >&2
    registries=$(api_request "/registry.all" || echo "[]")
    if [ "$DRY_RUN" = true ]; then
        echo "Found $(echo "$registries" | jq -r 'length') registries" >&2
    else
        echo "# ============================================================================="
        echo "# Container Registries"
        echo "# ============================================================================="
        echo ""
        echo "$registries" | jq -r '.[] | "\(.registryId) \(.registryName)"' | while read -r id name; do
            sname=$(sanitize_name "$name")
            import_block "dokploy_registry" "$sname" "$id"
        done
    fi

    # Certificates
    echo -e "${GREEN}Fetching certificates...${NC}" >&2
    certificates=$(api_request "/certificates.all" || echo "[]")
    if [ "$DRY_RUN" = true ]; then
        echo "Found $(echo "$certificates" | jq -r 'length') certificates" >&2
    else
        echo "# ============================================================================="
        echo "# SSL Certificates"
        echo "# ============================================================================="
        echo ""
        echo "$certificates" | jq -r '.[] | "\(.certificateId) \(.name)"' | while read -r id name; do
            sname=$(sanitize_name "$name")
            import_block "dokploy_certificate" "$sname" "$id"
        done
    fi

    # Destinations
    echo -e "${GREEN}Fetching backup destinations...${NC}" >&2
    destinations=$(api_request "/destination.all" || echo "[]")
    if [ "$DRY_RUN" = true ]; then
        echo "Found $(echo "$destinations" | jq -r 'length') destinations" >&2
    else
        echo "# ============================================================================="
        echo "# Backup Destinations"
        echo "# ============================================================================="
        echo ""
        echo "$destinations" | jq -r '.[] | "\(.destinationId) \(.name)"' | while read -r id name; do
            sname=$(sanitize_name "$name")
            import_block "dokploy_destination" "$sname" "$id"
        done
    fi
fi

# Fetch and process projects
if [ "$INFRASTRUCTURE_ONLY" = false ]; then
    echo -e "${GREEN}Fetching projects...${NC}" >&2
    projects=$(api_request "/project.all" || echo "[]")

    if [ "$DRY_RUN" = true ]; then
        echo "Found $(echo "$projects" | jq -r 'length') projects" >&2
        echo "$projects" | jq -r '.[] | "  Project: \(.name) (\(.projectId)) - \(.environments | length) environments"' >&2
    else
        echo "# ============================================================================="
        echo "# Projects"
        echo "# ============================================================================="
        echo ""

        # Process each project
        echo "$projects" | jq -c '.[]' | while read -r project; do
            proj_id=$(echo "$project" | jq -r '.projectId')
            proj_name=$(echo "$project" | jq -r '.name')
            proj_sname=$(sanitize_name "$proj_name")

            import_block "dokploy_project" "$proj_sname" "$proj_id"

            # Process environments
            echo "$project" | jq -c '.environments[]?' | while read -r env; do
                env_id=$(echo "$env" | jq -r '.environmentId')
                env_name=$(echo "$env" | jq -r '.name')
                env_sname="${proj_sname}_$(sanitize_name "$env_name")"

                import_block "dokploy_environment" "$env_sname" "$env_id"

                # Applications
                echo "$env" | jq -c '.applications[]?' 2>/dev/null | while read -r app; do
                    app_id=$(echo "$app" | jq -r '.applicationId')
                    app_name=$(echo "$app" | jq -r '.name')
                    app_sname="${env_sname}_$(sanitize_name "$app_name")"
                    import_block "dokploy_application" "$app_sname" "$app_id"
                done

                # Compose services
                echo "$env" | jq -c '.compose[]?' 2>/dev/null | while read -r comp; do
                    comp_id=$(echo "$comp" | jq -r '.composeId')
                    comp_name=$(echo "$comp" | jq -r '.name')
                    comp_sname="${env_sname}_$(sanitize_name "$comp_name")"
                    import_block "dokploy_compose" "$comp_sname" "$comp_id"
                done

                # PostgreSQL
                echo "$env" | jq -c '.postgres[]?' 2>/dev/null | while read -r pg; do
                    pg_id=$(echo "$pg" | jq -r '.postgresId')
                    pg_name=$(echo "$pg" | jq -r '.name')
                    pg_sname="${env_sname}_$(sanitize_name "$pg_name")"
                    import_block "dokploy_postgres" "$pg_sname" "$pg_id"
                done

                # MySQL
                echo "$env" | jq -c '.mysql[]?' 2>/dev/null | while read -r mysql; do
                    mysql_id=$(echo "$mysql" | jq -r '.mysqlId')
                    mysql_name=$(echo "$mysql" | jq -r '.name')
                    mysql_sname="${env_sname}_$(sanitize_name "$mysql_name")"
                    import_block "dokploy_mysql" "$mysql_sname" "$mysql_id"
                done

                # MariaDB
                echo "$env" | jq -c '.mariadb[]?' 2>/dev/null | while read -r mariadb; do
                    mariadb_id=$(echo "$mariadb" | jq -r '.mariadbId')
                    mariadb_name=$(echo "$mariadb" | jq -r '.name')
                    mariadb_sname="${env_sname}_$(sanitize_name "$mariadb_name")"
                    import_block "dokploy_mariadb" "$mariadb_sname" "$mariadb_id"
                done

                # MongoDB
                echo "$env" | jq -c '.mongo[]?' 2>/dev/null | while read -r mongo; do
                    mongo_id=$(echo "$mongo" | jq -r '.mongoId')
                    mongo_name=$(echo "$mongo" | jq -r '.name')
                    mongo_sname="${env_sname}_$(sanitize_name "$mongo_name")"
                    import_block "dokploy_mongo" "$mongo_sname" "$mongo_id"
                done

                # Redis
                echo "$env" | jq -c '.redis[]?' 2>/dev/null | while read -r redis; do
                    redis_id=$(echo "$redis" | jq -r '.redisId')
                    redis_name=$(echo "$redis" | jq -r '.name')
                    redis_sname="${env_sname}_$(sanitize_name "$redis_name")"
                    import_block "dokploy_redis" "$redis_sname" "$redis_id"
                done
            done
        done
    fi
fi

echo -e "${GREEN}Done!${NC}" >&2
