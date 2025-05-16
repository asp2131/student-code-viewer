# 📝 Task Board – Student Code Viewer UI Refactoring

### Legend
- 🔵 **To‑Do**
- 🟡 **In‑Progress**
- 🟢 **Done**

## Phase 1 – Refactor Menu Structure
| ID | Task | Status |
|----|------|-------|
| T-1 | Create top-level menu structure with 3 main options (Manage Classes, Create Class, Quit) | 🟢 |
| T-2 | Implement class selection prompt after selecting "Manage Classes" | 🟢 |
| T-3 | Create model structure for handling nested menus and state transitions | 🟢 |
| T-4 | Update UI styling for improved readability and modern appearance | 🟢 |

## Phase 2 – Implement Class Management Submenu
| ID | Task | Status |
|----|------|-------|
| T-5 | Break main.go into multiple files | 🟢 |
| T-6 | Create "Manage Students" submenu with Add/Delete/Back options | 🟢 |
| T-7 | Create "Manage Repos" submenu with Clone/Pull/Clean/Back options | 🟢 |
| T-8 | Create "View GH Activity" submenu with Week View/Check Activity/Back options | 🟢 |
| T-9 | Implement Github Activity Check Activity functionality | 🟢 |

## Phase 3 – Enhance UI/UX
| ID | Task | Status |
|----|------|-------|
| T-10 | Add visual indicators for current menu location (breadcrumb navigation) | 🔵 |
| T-12 | Implement keyboard shortcuts for common actions | 🔵 |
| T-13 | Add confirmation dialogs for destructive actions | 🔵 |
| T-14 | Improve error handling and user feedback messages | 🔵 |
| T-15 | Add help text/tooltips for menu options | 🔵 |
| T-16 | Implement progressive disclosure (only show relevant options based on context) | 🔵 |
| T-17 | Add visual feedback for long operations (loading indicators/progress bars) | 🔵 |

## Phase 4 – Testing & Refinement
| ID | Task | Status |
|----|------|-------|
| T-18 | Test navigation flows between all menus and submenus | 🔵 |
| T-19 | Verify all functionality works with new menu structure | 🔵 |
| T-20 | Gather teacher feedback on new UI organization | 🔵 |
| T-21 | Make refinements based on feedback | 🔵 |

## Technical Implementation Details

### Menu State Management
- Extend the current state model to handle nested menu states
- Create a menu history stack for back navigation
- Implement state transitions between menus and submenus

### UI Components to Modify
- Update the `initialModel()` function to create the new menu structure
- Modify the `Update()` method to handle submenu navigation
- Enhance the `View()` method to display current menu context
- Create helper functions for generating submenu displays

## Backlog
- Consider adding user preferences/settings menu
- Explore adding visual graphs for student activity statistics
- Investigate adding search/filter capabilities for large classes

> Update statuses daily at stand‑up. Sync with GitHub Projects kanban.