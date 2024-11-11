# Student Code Viewer

Scripts to easily view student code in one location and track recent code submissions.

## Getting Started - Codespace Setup

1. **Create a GitHub Codespace**:

   - Sign in to GitHub.
   - Click the green "<> Code" button at the top of this repository.
   - Switch to the "Codespaces" tab.
   - Click "Create codespace on main".

2. **Install Live Server Extension**:

   - Download the "Live Server" extension.
   - Refer to [this slide](https://docs.google.com/presentation/d/1USzVPXUQK6IWOHWi8r8_Yj0rJ8gxzvjT2mo2X27KKaU/edit#slide=id.g2a825dd5b6a_0_558) for installation help.

3. **Make Scripts Executable**:
   - Run the following command in the terminal to avoid permissions errors:
     ```
     chmod +x *.sh
     ```

### Adding Student GitHub Usernames

- Add each student's GitHub username to `list.txt`, one username per line.
- You can find a copyable list in the main grading or project links spreadsheet for your class.

## Cloning Student Repositories

- Run this command to clone all student repositories:
  ```
  ./clone-all.sh
  ```

_**Note:** Re-run this command if new students are added to update the repositories._

## Pulling Student Code

- To update student code, run:
  ```
  ./pull-all.sh
  ```

### ❗❗❗ Viewing Recent Commits ❗❗❗

- Running `./pull-all.sh` shows which students have pushed code within the last hour. The terminal output will indicate recent commits.

  Example:

  ```
  commits within the last hour are below:

  dkogler
  ✅ ✅ ✅ dkogler has committed today! committed at Tue Feb 20 23:19:32 2024 +0000 ✅ ✅ ✅

  CooolHandLuke
  🚧 🚧 🚧 CooolHandLuke has not committed within the last hour! 🚧 🚧 🚧

  HolaAmigo
  🚩 🚩 🚩 HolaAmigo does not have a matching github repo. Check with them to ensure they have named their repository correctly.
  Visit https://github.com/HolaAmigo to see their existing repositories 🚩 🚩 🚩
  ```

  🔥🔥🔥 _**Tip:** Run this command at the end of a project workday to check which students haven't pushed their code._ 🔥🔥🔥

## Checking Student Work and Debugging

- After pulling student code, you can:
  - Navigate to any student's folder.
  - Use Live Server to load their project.
  - Use the browser console to view error messages.
  - Edit student code to help them debug issues.

_**Note:** Your changes won't affect the student's repository unless you have "Write" access._

## Reverting Edits

- After making changes, revert student code to its original state by running:
  ```
  ./clean-all.sh
  ```

## Handling Errors

### Permissions Errors

- If you see a permissions error like:
  ```
  bash: permission denied: ./pull-all.sh
  ```
  Ensure you have run the `chmod` command from the setup section.

### Missing Student Repositories

- If a repository is missing, you might see:

  ```
  Cloning into 'HolaAmigo'...
  remote: Repository not found.
  fatal: repository 'https://github.com/HolaAmigo/HolaAmigo.github.io/' not found
  ```

  The terminal will provide feedback with a link to the student's GitHub account.

  _**Tip:** Ensure students name their repository correctly. Re-run `./clone-all.sh` after corrections._

## Working with Multiple Classes

- Create multiple `.txt` files for different classes, named as `list____.txt` (e.g., `list2.txt`, `list4.txt`).
- Use the appropriate commands with the class number:
  ```
  ./clone-all.sh 2
  ./pull-all.sh 2
  ./clean-all.sh 2
  ```
  Adjust for any class by creating a unique `list___.txt` file.

_**Note:** All student repositories will be at the root level of the codespace._

## Final Notes

- Avoid running `git` commands directly in the codespace.
- The provided scripts handle all necessary `git` commands to track and update student work.
