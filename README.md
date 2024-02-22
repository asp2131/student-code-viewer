# student-code-viewer

Scripts to allow easy viewing of student code in one location, and info about students that have pushed code recently

## Getting Started - Codespace Setup

To get started with this tool, create a GitHub Codespace from this repository. Click the green "<> Code" button at the top of the repo, switch to the "Codespaces" tab, and click the "Create codespace on main" button.

Once in the codespace, download the "Live Server" extension. See [this slide from the setup instructions](https://docs.google.com/presentation/d/1USzVPXUQK6IWOHWi8r8_Yj0rJ8gxzvjT2mo2X27KKaU/edit#slide=id.g2a825dd5b6a_0_558) for help downloading if needed.

Next, run the following command to make all scripts in the repository executable. If this command is not run, running any of the commands in the below sections will result in a permissions error. It only needs to be run the first time you are working in your Codespace.

- `chmod +x *.sh`

### Adding Student GitHub Usernames

In the `list.txt` file, add each student's github username in a separate line. A copy-able list can be found in the main grading spreadsheet or the project links spreadsheet for your class.

## Cloning Student Repos

To clone student repositories, run the following command in the terminal:

- `./clone-all.sh`

These commands will clone down all students' repositories and give you access to student work.

_**Note:** if a new student is added to a class, this command can be re-run to install the repository of any new username that has been added to the `list.txt` file._

## Pulling Student Code

To get updated student code, run the following command in the terminal:

- `./pull-all.sh`

These commands will run `git pull` and pull students' code down from GitHub so that the code you're looking at in the workspace is up to date with what they have pushed to GitHub.

### ❗❗❗ Viewing Recent Commits ❗❗❗

Running the `./pull-all.sh` command from the above step will allow you to see which students have pushed up code within the last hour. At the bottom of the output in the terminal, you'll see text output that looks similar to the following:

```
------------------

❗❗❗❗❗❗❗❗❗❗❗❗❗❗❗

commits within the last hour are below:

dkogler
✅ ✅ ✅ dkogler has committed today! committed at Tue Feb 20 23:19:32 2024 +0000 ✅ ✅ ✅

gregthompson27
✅ ✅ ✅ gregthompson27 has committed today! committed at Tue Feb 20 23:24:09 2024 +0000 ✅ ✅ ✅

CooolHandLuke
🚧 🚧 🚧 CooolHandLuke has not committed within the last hour! 🚧 🚧 🚧

susieklu
✅ ✅ ✅ susieklu has committed today! committed at Tue Feb 20 23:23:21 2024 +0000 ✅ ✅ ✅

asp2131
🚧 🚧 🚧 asp2131 has not committed within the last hour! 🚧 🚧 🚧

mayukhraychauduri
✅ ✅ ✅ mayukhraychauduri has committed today! committed at Tue Feb 20 23:27:35 2024 +0000 ✅ ✅ ✅
```

In this example, you can see that two students haven't pushed code 🚧 🚧 🚧 - those with GitHub usernames `CooolHandLuke` and `asp2131` 🚧 🚧 🚧

🔥🔥🔥 This command is extremely helpful to run at the end of a project work day to determine which students haven't pushed up code. The `./pull-all.sh` script can be run multiple times, and each time it is run the text output in the terminal will show an updated list of "commits within the last hour" if any student has pushed code since the last time you ran the `pull-all` script. 🔥🔥🔥

## Checking Student Work + Bugs

This tool can be useful for helping debug student code if a student has run into an issue and pushed their code up. After running the [pull-all commands](#pulling-student-code) from the above section, you can go into any student's folder and go into the specific files of projects they are working on. You can use live server to load the project, use the console in your browser to see any error messages, and edit student code to help determine what students need to do to get past roadblocks.

_**Note:** doing this will not change student's code in their repositories in any way. You are just viewing and editing their code remotely in a separate codespace from theirs. Also, you won't be able to push code up to their repository without "Write" access, which you won't have._

## Clear Any Edits to Pull Code Again

After making any changes to student code, you'll want to revert their code to its original state or the `pull-all` script will not work for that student. To revert **all** students' code in your Codespace, run the following command in your terminal:

- `./clean-all.sh`

## Errors When Using these Scripts

### Permissions Errors When Running Scripts

If you run on of the commands (`./clone-all.sh`, `./pull-all.sh`, or `./clean-all.sh`) and get the following error in the terminal:

```
bash: permission denied: ./pull-all.sh
```

this is likely a result of not running the `chmod` command from the [Getting Started Section](#getting-started---codespace-setup). Run the provided command, and then re-run the command that caused the initial permissions error.

### Errors Resulting From Missing Student Repositories

If a student has not properly named their repository to match the `<username>.github.io` format that creates a GitHub Pages site as described [here on GitHub](https://pages.github.com/), there may be some errors that occur when attempting to run the `clone-all` and `pull-all` scripts. If the `clone-all` script does not find a correctly named repository for a student, you'll get an error in the terminal output that looks similar to the following, and there will be no folder created for that student in your Codespace.

```
Cloning into 'HolaAmigo'...
remote: Repository not found.
fatal: repository 'https://github.com/HolaAmigo/HolaAmigo.github.io/' not found
```

If left unfixed, any usernames with missing repos will be skipped over when the `pull-all` script is run, but there will be text feedback in the terminal with a link to the student's GitHub account to help determine what has gone wrong for this student, that looks like the below message:

```
🚩 🚩 🚩 HolaAmigo does not have a matching github repo. Check with them to ensure they have named their repository correctly.
Visit https://github.com/HolaAmigo to see their existing repositories 🚩 🚩 🚩
```

🩹🩹🩹 - Once a student has correctly named their repository (or created one if they didn't have one previously), you can re-run the `clone-all` [command](#cloning-student-repos) to clone down that student's repo into your codespace.

## Final Notes

Don't try to run any `git` commands in your codespace. The three different scripts are designed to run all required git commands to track student work and determine students that have pushed code up at the end of class and those that haven't. Running any `git` commands may result in unwanted consequences 💀 💀 💀.
