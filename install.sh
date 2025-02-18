#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Installing Student Code Viewer (scv)...${NC}"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    echo "Please install Go first:"
    echo "  Mac: brew install go"
    echo "  Linux: sudo apt install golang"
    echo "  Windows: https://golang.org/dl/"
    exit 1
fi

# Create temporary directory
TMP_DIR=$(mktemp -d)
cd $TMP_DIR

# Clone the repository
echo "Downloading source code..."
git clone https://github.com/asp2131/student-code-viewer.git
cd student-code-viewer

# Initialize Go module
echo "Initializing Go module..."
go mod init github.com/asp2131/student-code-viewer
go mod tidy

# Build the binary
echo "Building scv..."
go build -o scv

# Install the binary
echo "Installing to /usr/local/bin..."
sudo mv scv /usr/local/bin/
sudo chmod +x /usr/local/bin/scv

# Clean up
cd ..
rm -rf $TMP_DIR

echo -e "${GREEN}Installation complete!${NC}"
echo -e "Run ${BLUE}scv --help${NC} to get started"
echo
echo -e "${RED}Important:${NC} You'll need to set up a GitHub token for activity tracking:"
echo "1. Visit: https://github.com/settings/tokens"
echo "2. Generate a new token with 'repo' scope"
echo "3. Run: export GITHUB_TOKEN=your_token_here"
echo