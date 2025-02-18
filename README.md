# Student Code Viewer (scv)

A CLI tool to easily manage and track student code submissions on GitHub.

## Quick Install

```bash
# For Mac/Linux:
curl -sSL https://raw.githubusercontent.com/asp2131/student-code-viewer/main/install.sh | bash

# OR install manually with Go:
go install github.com/asp2131/student-code-viewer@latest
```

## Features

- 📚 Manage multiple classes
- 👥 Track student repositories
- ⏰ Monitor commit activity
- 🔄 Bulk clone and update repositories
- 🧹 Clean local changes

## Prerequisites

- Git installed on your system
- GitHub Personal Access Token (for activity tracking)

## Usage

### First Time Setup

```bash
# Set your GitHub token (only needed once)
export GITHUB_TOKEN=your_github_token_here
scv check-activity section1

# Token will be saved automatically for future use
```

### Basic Commands

```bash
# Create a new class
scv add-class section1

# Add students to a class
scv add-student section1 student1 student2 student3

# List students in a class
scv list-students section1

# Check recent activity
scv check-activity section1

# Clone all repositories
scv clone section1

# Pull latest changes
scv pull section1

# Clean local changes
scv clean section1
```

### Activity Monitoring

The `check-activity` command shows when students last pushed code:

```bash
scv check-activity section1

Activity Report for section1:
----------------------------------------
✅ student1: Last push 2h 15m ago
🟡 student2: Last push 2d 5h ago
⚠️ student3: Last push 5d 12h ago

Legend:
✅ - Pushed within last 24 hours
🟡 - Pushed within last 72 hours
⚠️ - No push in over 72 hours
❌ - Error checking activity
```

## GitHub Token Setup

1. Go to [GitHub Settings](https://github.com/settings/tokens)
2. Click "Generate new token (classic)"
3. Select the following scopes:
   - `public_repo` (for public repositories)
   - `repo` (if using private repositories)
4. Copy the generated token
5. Run any scv command and enter the token when prompted

## Configuration

The tool stores configuration in `~/.scv.json`. You can view current settings with:

```bash
scv config show
```

## Error Handling

### Common Issues

1. **Permission Denied**
   ```bash
   # Make sure scv is executable
   chmod +x /usr/local/bin/scv
   ```

2. **GitHub Token Not Set**
   ```bash
   # Set token manually
   export GITHUB_TOKEN=your_token_here
   ```

3. **Repository Not Found**
   - Check that student usernames are correct
   - Verify repository naming convention

## Contributing

Contributions are welcome! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.