# Command Menu Structure

## Current Structure
Currently the commands that can be run are all on a carousel-style menu, with only two options displayed at once. Here are the list of commands:

- Add class
- Remove class
- List classes
- Add Students
- Remove Students
- List Students
- Clone Repos
- Pull Changes
- Clean Changes
- Check Activity
- Week History
- Quiz

## Potential Update to Use Submenus

### Top level menus
- Manage Classes
  - *Prompted to select a class*
- Create Class
- Quit

### After selecting a class from "Manage Classes" prompt
- Manage Students
  - Add students
  - Delete students
  - Back
- Manage Repos
  - Clone Repos
  - Pull Repos
  - Clean Repos
  - Back
- View GH Activity
  - Week View
  - Check Activity (last commit date, color-coded)
  - Back
- Delete Class
  - Press enter to confirm you'd like to delete <XYZ> class → back to main menu if confirming
- Back