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

Student Code Viewer (scv) provides an interactive Terminal User Interface (TUI) to:

- 📚 **Manage Multiple Classes**: Easily create, select, and organize different cohorts or sections.
- 🧑‍🎓 **Track Students per Class**: Add and manage lists of students for each class.
- 📊 **Monitor Student Activity**:
    - View the time since each student's last GitHub push.
    - See a "Week View" summarizing daily commit activity (Mon-Fri) for all students in a class.
- 🔄 **Batch Repository Operations**:
    - Clone all student repositories for a selected class with a single action.
    - Pull the latest updates for all student repositories in a class.
- ⚙️ **Simple Configuration**: Automatic saving of class data and GitHub token.
- ✨ **User-Friendly Interface**: Navigate all features through intuitive menus.

## Prerequisites

- Git installed on your system
- GitHub Personal Access Token (for activity tracking)

## Usage

### Launching scv

To start Student Code Viewer, simply run the `scv` command in your terminal:

```bash
scv
```
If you're running from the source code, you can use:
```bash
go run .
```
This will launch the interactive Terminal User Interface (TUI).

### First-Time Setup (GitHub Token)

On your first run, `scv` will likely guide you to set up your GitHub Personal Access Token. This token is necessary for fetching student activity from GitHub.
- When prompted by the TUI, enter your GitHub token.
- The token will be securely saved for future sessions.
For details on generating a token, see the "GitHub Token Setup" section further down.

### Navigating the TUI

Use your keyboard to navigate the TUI:
- **Arrow Keys (Up/Down/Left/Right)**: Move selection.
- **Enter**: Confirm selection or action.
- **Esc**: Go back or exit a view/menu.
- Follow on-screen prompts and menu options.

### Core Functionality via TUI

All primary operations are performed within the TUI:

1.  **Managing Classes**:
    *   From the main menu, select an option like "Manage Classes" or "Select Class."
    *   You can typically:
        *   **Add a New Class**: Enter a name for your new class.
        *   **Select an Existing Class**: Choose from a list of your classes to make it active.
        *   (Other class management options like renaming or deleting might be available).

2.  **Managing Students**:
    *   Once a class is active, navigate to an option like "Manage Students" or "View Students."
    *   You can typically:
        *   **Add Students**: Enter GitHub usernames for students in the current class.
        *   (Other student management options like removing students might be available).

3.  **Viewing Student Activity**:
    *   With a class selected, choose an option like "View Student Activity."
    *   **Last Push Details**: This view will show a list of students and how long ago they made their last commit.
    *   **Week View**: From the student activity view (or a dedicated menu option), select "Show Week View" (or similar). This displays a table with daily (Mon-Fri) commit indicators (✓/✗) for each student.

4.  **Cloning & Pulling Repositories**:
    *   Select the desired class.
    *   Choose options like:
        *   "Clone All Repositories": Clones all repositories for students in the selected class.
        *   "Pull All Repositories": Fetches the latest updates for all repositories in the selected class.
    *   Success messages will be displayed directly in the TUI.

5.  **Configuration**:
    *   The application typically stores its configuration (including the GitHub token and class/student lists) in a file like `~/.scv.json`. You might find options within the TUI to view or manage settings.

## GitHub Token Setup

To allow `scv` to fetch data from GitHub (like student activity), it needs a Personal Access Token.

1.  Go to [GitHub Settings](https://github.com/settings/tokens) to generate a new token.
2.  Click "Generate new token (classic)".
3.  Give your token a descriptive name (e.g., "scv-tool").
4.  Select the following scopes:
    *   `public_repo` (if all student repositories are public)
    *   `repo` (if you need to access private student repositories)
5.  Copy the generated token.
6.  **Launch `scv`**. If it's your first time or the token is missing, the TUI will prompt you to enter it. Paste your token there. It will be saved for future use.

As an alternative, or for specific environments, you can set the `GITHUB_TOKEN` environment variable. However, the TUI prompt is the recommended method for most users.

## Configuration

The tool stores configuration in `~/.scv.json`. You can view current settings with:

```bash
scv config show
```

## Error Handling

### Common Issues

1.  **Permission Denied**
    ```bash
    # Make sure scv is executable
    chmod +x /usr/local/bin/scv
    ```

2.  **GitHub Token Not Set / Invalid Token**
    - If `scv` cannot access GitHub data, ensure your token is correctly set.
    - **Relaunch `scv`**: The TUI should prompt you for the token if it's missing or invalid.
    - **Check Configuration**: The token is stored in `~/.scv.json`. You can inspect this file (with caution) or look for TUI options to re-enter/update the token.
    - **Environment Variable**: As a fallback, you can try setting the `GITHUB_TOKEN` environment variable:
      ```bash
      export GITHUB_TOKEN=your_token_here
      ```
    - Verify the token has the correct scopes on GitHub.

3.  **Repository Not Found**
    - Check that student usernames are correct
    - Verify repository naming convention

## Contributing

Contributions are welcome! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.