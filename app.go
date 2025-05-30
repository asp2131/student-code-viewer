package main

import (
	"fmt"
	"strings"
	"time"
	"math" // Added for Modulo operation

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput" // Re-added this import
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func initialModel() Model {
	// Initialize list component
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 40, 12)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	// Initialize text inputs
	classInput := textinput.New()
	classInput.Placeholder = "Enter class name"
	classInput.Focus()
	classInput.CharLimit = 50
	classInput.Width = 40

	studentInput := textinput.New()
	studentInput.Placeholder = "Enter student GitHub username(s), comma-separated"
	studentInput.Focus()
	studentInput.CharLimit = 200
	studentInput.Width = 40

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	// Create main menu items
	menuItems := []Item{
		{title: "Manage Classes", description: "Select and manage an existing class"},
		{title: "Create Class", description: "Create a new class"},
		{title: "Quit", description: "Exit the application"},
	}

	return Model{
		list:          l,
		state:         stateMainMenu,
		classInput:    classInput,
		studentInput:  studentInput,
		output:        "",
		menuHistory:   make([]int, 0),
		classList:     []string{},
		selectedClass: "",
		currentMenu:   "Main Menu",
		breadcrumbPath: []string{"Home"},
		menuItems:     menuItems,
		studentList:   []string{}, // Initialize studentList
		selectedItem:  0,
		spinner:       s, // Initialize spinner
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle errors first
	if m.err != nil {
		// If there's an error, just wait for any key press to continue
		if _, ok := msg.(tea.KeyMsg); ok {
			m.err = nil
			return m, nil
		}
		return m, nil
	}

	// We'll handle list updates at the end of the function to ensure key presses are processed
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "esc":
			// Allow escaping from class input (creating class) back to main menu
			if m.state == stateClassInput && m.currentMenu == "Create Class" {
				// Return to main menu
				m.state = stateMainMenu
				m.currentMenu = "Main Menu"
				m.selectedItem = 0
				return m, nil
			} else if m.state == stateStudentInput { // Added for escaping student input
				// Return to Manage Students menu
				if len(m.menuHistory) > 0 {
					lastIndex := len(m.menuHistory) - 1
					previousState := m.menuHistory[lastIndex]
					m.menuHistory = m.menuHistory[:lastIndex]

					if previousState == stateManageStudents { // Ensure it's the correct previous state
						m.state = stateManageStudents
						m.currentMenu = "Manage Students: " + m.selectedClass
						m.selectedItem = 0 // Reset selection for the manage students menu
						return m, nil
					}
				}
				// Fallback if history is empty or not what we expect (go to main menu)
				m.state = stateMainMenu
				m.currentMenu = "Main Menu"
				m.selectedItem = 0
				return m, nil
			}

		// Handle menu navigation
		case "up", "k":
			if m.state == stateMainMenu {
				// Move selection up in main menu
				if m.selectedItem > 0 {
					m.selectedItem--
				}
				return m, nil
			} else if m.state == stateClassSelection {
				// Move selection up in class selection
				if m.selectedItem > 0 {
					m.selectedItem--
				}
				return m, nil
			} else if m.state == stateClassManagement {
				// Move selection up in class management
				if m.selectedItem > 0 {
					m.selectedItem--
				}
				return m, nil
			} else if m.state == stateManageStudents { // Added for Manage Students menu
				// Move selection up in manage students menu
				if m.selectedItem > 0 {
					m.selectedItem--
				}
				return m, nil
			} else if m.state == stateStudentSelectionForDelete { // Added for student deletion selection menu
				// Move selection up
				if m.selectedItem > 0 {
					m.selectedItem--
				}
				return m, nil
			} else if m.state == stateManageRepos { // Added for Manage Repos menu
				// Move selection up
				if m.selectedItem > 0 {
					m.selectedItem--
				}
				return m, nil
			} else if m.state == stateViewGHActivity { // Added for GitHub Activity menu
				// Move selection up
				if m.selectedItem > 0 {
					m.selectedItem--
				}
				return m, nil
			}

		case "down", "j":
			if m.state == stateMainMenu {
				// Move selection down in main menu
				if m.selectedItem < len(m.menuItems)-1 {
					m.selectedItem++
				}
				return m, nil
			} else if m.state == stateClassSelection {
				// Move selection down in class selection
				if m.selectedItem < len(m.classList) {
					m.selectedItem++
				}
				return m, nil
			} else if m.state == stateClassManagement {
				// Move selection down in class management
				managementOptions := []string{
					"Manage Students",
					"Manage Repos",
					"View GH Activity",
					"Delete Class",
					"Back",
				}
				if m.selectedItem < len(managementOptions)-1 {
					m.selectedItem++
				}
				return m, nil
			} else if m.state == stateManageStudents { // Added for Manage Students menu
				// Move selection down in manage students menu (3 items total)
				// Items: "Add Student(s)", "Delete Student", "Back"
				if m.selectedItem < 2 { // 0, 1, 2 are valid indices
					m.selectedItem++
				}
				return m, nil
			} else if m.state == stateStudentSelectionForDelete { // Added for student deletion selection menu
				// Move selection down (studentList + Back option)
				if m.selectedItem < len(m.studentList) { // studentList is 0-indexed, Back is at len(m.studentList)
					m.selectedItem++
				}
				return m, nil
			} else if m.state == stateManageRepos { // Added for Manage Repos menu
				// Move selection down (4 items total)
				// Items: "Clone All Repos", "Pull All Repos", "Clean All Repos", "Back"
				if m.selectedItem < 3 { // 0, 1, 2, 3 are valid indices
					m.selectedItem++
				}
				return m, nil
			} else if m.state == stateViewGHActivity { // Added for GitHub Activity menu
				// Move selection down
				if m.selectedItem < 2 { // 0, 1, 2 are valid indices
					m.selectedItem++
				}
				return m, nil
			}

		case "enter":
			// Handle state transitions based on current state
			if m.state == stateMainMenu {
				// Handle main menu selection
				if m.selectedItem < len(m.menuItems) {
					selectedOption := m.menuItems[m.selectedItem].title
					switch selectedOption {
					case "Quit":
						return m, tea.Quit
					case "Manage Classes":
						// Load available classes
						classes, err := getClasses()
						if err != nil {
							m.err = err
							return m, nil
						}

						if len(classes) == 0 {
							m.output = "No classes found. Please create a class first."
							m.menuHistory = append(m.menuHistory, m.state)
							m.state = stateOutput
							return m, nil
						}

						// Save the current state to history for back navigation
						m.menuHistory = append(m.menuHistory, m.state)

						// Create and set the class selection menu
						m.list = createClassSelectionMenu(classes)
						m.classList = classes
						m.currentMenu = "Class Selection"
						m.selectedItem = 0
						m.state = stateClassSelection
						return m, nil

					case "Create Class":
						// Save current state to history
						m.menuHistory = append(m.menuHistory, m.state)

						// Reset class input and switch to class input state
						m.classInput.SetValue("")
						m.classInput.Focus()
						m.state = stateClassInput
						m.currentMenu = "Create Class"
						return m, nil
					}
				}
			} else if m.state == stateClassSelection {
				// Handle class selection
				if m.selectedItem == len(m.classList) {
					// Back option selected
					if len(m.menuHistory) > 0 {
						// Pop the last state from history
						lastIndex := len(m.menuHistory) - 1
						m.state = m.menuHistory[lastIndex]
						m.menuHistory = m.menuHistory[:lastIndex]

						// Reset to main menu
						m.selectedItem = 0
						m.currentMenu = "Main Menu"
					}
					return m, nil
				} else if m.selectedItem < len(m.classList) {
					// Save the selected class name
					m.selectedClass = m.classList[m.selectedItem]

					// Save current state to history
					m.menuHistory = append(m.menuHistory, m.state)

					// Create and set the class management menu
					m.list = createClassManagementMenu(m.selectedClass)
					m.currentMenu = "Class Management: " + m.selectedClass
					m.selectedItem = 0
					m.state = stateClassManagement
					return m, nil
				}
			} else if m.state == stateClassManagement {
				// Determine selected option based on index
				// 0: Manage Students, 1: Manage Repos, 2: View GH Activity, 3: Delete Class, 4: Back
				selectedOptionIndex := m.selectedItem

				switch selectedOptionIndex {
				case 0: // Manage Students
					m.menuHistory = append(m.menuHistory, m.state)
					m.state = stateManageStudents
					m.currentMenu = "Manage Students: " + m.selectedClass
					m.selectedItem = 0 // Reset for the new menu
					return m, nil
				case 1: // Manage Repos
					m.menuHistory = append(m.menuHistory, m.state)
					m.state = stateManageRepos
					m.currentMenu = "Manage Repositories: " + m.selectedClass
					m.selectedItem = 0 // Reset for the new menu
					return m, nil
				case 2: // View GH Activity
					m.currentMenu = fmt.Sprintf("GitHub Activity for Class: %s", m.selectedClass)
					m.menuHistory = append(m.menuHistory, m.state)
					m.state = stateViewGHActivity
					m.selectedItem = 0 // Initialize to first menu item
					return m, nil
				case 3: // Delete Class
					// Transition to delete confirmation state
					m.menuHistory = append(m.menuHistory, m.state)
					m.state = stateDeleteConfirmation
					m.currentMenu = "Confirm Class Deletion: " + m.selectedClass
					
					// Initialize the menu items for delete confirmation
					confirmItems := []list.Item{
						Item{title: "Yes, delete class", description: "Permanently delete this class and all student data"},
						Item{title: "No, cancel", description: "Return to class management menu"},
					}
					
					// Configure the list for the confirmation menu
					m.list.Title = "Confirm Deletion"
					m.list.SetShowStatusBar(false)       // Hide the status bar
					m.list.SetFilteringEnabled(false)    // No filtering needed
					m.list.Styles.Title = titleStyle     // Use consistent styling
					m.list.SetShowHelp(true)             // Show keyboard navigation help
					m.list.SetShowPagination(false)      // No pagination for 2 items
					m.list.SetItems(confirmItems)        // Set the menu items
					m.list.Select(0)                     // Select the first item by default
					return m, nil
				case 4: // Back
					if len(m.menuHistory) > 0 {
						lastIndex := len(m.menuHistory) - 1
						m.state = m.menuHistory[lastIndex]
						m.menuHistory = m.menuHistory[:lastIndex]
						m.currentMenu = "Select a Class" // Or derive from previous state
						m.selectedItem = 0               // Reset selection
						return m, nil
					}
					// If no history, go to main menu
					m.state = stateMainMenu
					m.currentMenu = "Main Menu"
					m.selectedItem = 0
					return m, nil
				}
			} else if m.state == stateManageStudents { // Added for Manage Students menu
				// Determine selected option based on index
				// 0: Add Student(s), 1: Delete Student, 2: Back
				selectedOptionIndex := m.selectedItem

				switch selectedOptionIndex {
				case 0: // Add Student(s)
					m.menuHistory = append(m.menuHistory, m.state)
					m.state = stateStudentInput
					m.currentMenu = "Add Students to: " + m.selectedClass
					m.studentInput.SetValue("") // Clear previous input
					m.studentInput.Focus()
					m.selectedItem = 0 // Reset for student input (though not a menu)
					return m, textinput.Blink
				case 1: // Delete Student
					students, err := getStudents(m.selectedClass)
					if err != nil {
						m.err = fmt.Errorf("failed to get students for class '%s': %w", m.selectedClass, err)
						return m, nil
					}
					if len(students) == 0 {
						m.output = fmt.Sprintf("No students found in class '%s' to delete.", m.selectedClass)
						m.menuHistory = append(m.menuHistory, m.state)
						m.state = stateOutput
						return m, nil
					}
					m.studentList = students
					m.menuHistory = append(m.menuHistory, m.state)
					m.state = stateStudentSelectionForDelete
					m.currentMenu = "Delete Student from: " + m.selectedClass
					m.selectedItem = 0 // Reset for the new student selection menu
					return m, nil
				case 2: // Back
					if len(m.menuHistory) > 0 {
						lastIndex := len(m.menuHistory) - 1
						previousState := m.menuHistory[lastIndex]
						m.menuHistory = m.menuHistory[:lastIndex]

						// Ensure we return to the correct previous state and menu
						if previousState == stateClassManagement {
							m.state = stateClassManagement
							m.currentMenu = "Managing Class: " + m.selectedClass
							m.selectedItem = 0 // Sensible default, or restore previous m.selectedItem for this menu
						} else {
							// Fallback if history is unexpected (should ideally not happen)
							m.state = stateMainMenu
							m.currentMenu = "Main Menu"
							m.selectedItem = 0
						}
						return m, nil
					}
					// If no history, go to main menu
					m.state = stateMainMenu
					m.currentMenu = "Main Menu"
					m.selectedItem = 0
					return m, nil
				}
			} else if m.state == stateStudentSelectionForDelete { // Added for student deletion menu
				selectedStudentIndex := m.selectedItem

				if selectedStudentIndex == len(m.studentList) { // "Back" option selected
					if len(m.menuHistory) > 0 {
						lastIndex := len(m.menuHistory) - 1
						previousState := m.menuHistory[lastIndex]
						m.menuHistory = m.menuHistory[:lastIndex]

						if previousState == stateManageStudents {
							m.state = stateManageStudents
							m.currentMenu = "Manage Students: " + m.selectedClass
							m.selectedItem = 0
						} else {
							// Fallback
							m.state = stateMainMenu
							m.currentMenu = "Main Menu"
							m.selectedItem = 0
						}
						return m, nil
					}
					// Fallback if no history
					m.state = stateMainMenu
					m.currentMenu = "Main Menu"
					m.selectedItem = 0
					return m, nil
				} else if selectedStudentIndex < len(m.studentList) { // A student is selected
					studentToDelete := m.studentList[selectedStudentIndex]
					err := deleteStudent(m.selectedClass, studentToDelete)
					if err != nil {
						m.output = fmt.Sprintf("Error deleting student '%s': %v", studentToDelete, err)
					} else {
						m.output = fmt.Sprintf("Successfully deleted student '%s' from class '%s'.", studentToDelete, m.selectedClass)
					}
					// After deletion (or error), go to output state, then allow user to return to Manage Students
					// The history should still have ManageStudents from when we entered StudentSelectionForDelete
					m.state = stateOutput
					// m.currentMenu will be set by stateOutput's logic or can be set here if needed
					return m, nil
				}
			} else if m.state == stateManageRepos { // Added for Manage Repos menu
				selectedOptionIndex := m.selectedItem
				// 0: Clone, 1: Pull, 2: Clean, 3: Back

				switch selectedOptionIndex {
				case 0: // Clone All Repos
					m.menuHistory = append(m.menuHistory, m.state)
					m.state = stateLoading
					m.loadingMessage = fmt.Sprintf("Cloning all repositories for class '%s'...", m.selectedClass)
					return m, tea.Batch(
						m.spinner.Tick,
						func() tea.Msg {
							studentUsernames, err := getStudents(m.selectedClass)
							if err != nil {
								return operationResultMsg{err: fmt.Errorf("Error fetching students for class '%s': %w", m.selectedClass, err)}
							}
							logs := cloneAllRepos(m.selectedClass, studentUsernames)
							return operationResultMsg{logs: logs}
						},
					)
				case 1: // Pull All Repos
					m.menuHistory = append(m.menuHistory, m.state)
					m.state = stateLoading
					m.loadingMessage = fmt.Sprintf("Pulling all repositories for class '%s'...", m.selectedClass)
					return m, tea.Batch(
						m.spinner.Tick,
						func() tea.Msg {
							studentUsernames, err := getStudents(m.selectedClass)
							if err != nil {
								return operationResultMsg{err: fmt.Errorf("Error fetching students for class '%s': %w", m.selectedClass, err)}
							}
							logs := pullAllRepos(m.selectedClass, studentUsernames)
							return operationResultMsg{logs: logs}
						},
					)
				case 2: // Clean All Repos
					m.menuHistory = append(m.menuHistory, m.state)
					m.state = stateLoading
					m.loadingMessage = fmt.Sprintf("Cleaning all repositories for class '%s'...", m.selectedClass)
					return m, tea.Batch(
						m.spinner.Tick,
						func() tea.Msg {
							// No need to fetch students for clean, it removes the whole class directory
							logs := cleanAllRepos(m.selectedClass)
							return operationResultMsg{logs: logs}
						},
					)
				case 3: // Back
					if len(m.menuHistory) > 0 {
						lastIndex := len(m.menuHistory) - 1
						previousState := m.menuHistory[lastIndex]
						m.menuHistory = m.menuHistory[:lastIndex]

						if previousState == stateClassManagement {
							m.state = stateClassManagement
							m.currentMenu = "Managing Class: " + m.selectedClass
							m.selectedItem = 0
						} else {
							// Fallback
							m.state = stateMainMenu
							m.currentMenu = "Main Menu"
							m.selectedItem = 0
						}
						return m, nil
					}
					// Fallback if no history
					m.state = stateMainMenu
					m.currentMenu = "Main Menu"
					m.selectedItem = 0
					return m, nil
				}
			} else if m.state == stateViewGHActivity {
				if msg.String() == "enter" {
					selectedOptionIndex := m.selectedItem

					// Options: 0: Week View, 1: Check Specific Activity, 2: Back
					switch selectedOptionIndex {
					case 0: // Week View
						m.state = stateLoading
						m.currentMenu = "Week View"
						return m, tea.Batch(
							m.spinner.Tick,
							func() tea.Msg {
								studentUsernames, err := getStudents(m.selectedClass)
								if err != nil {
									return operationResultMsg{err: fmt.Errorf("Error fetching students for class '%s': %w", m.selectedClass, err)}
								}
								styledOutput, err := getStudentsCommitWeekViewActivity(m.selectedClass, studentUsernames)
								if err != nil {
									return operationResultMsg{err: fmt.Errorf("Error fetching GitHub week activity: %w", err)}
								}
								return operationResultMsg{logs: []string{styledOutput}}
							},
						)
					case 1: // Check Latest Activity (previously index 1)
						m.state = stateLoading
						m.currentMenu = "Check Latest Activity"
						return m, tea.Batch(
							m.spinner.Tick,
							func() tea.Msg {
								studentUsernames, err := getStudents(m.selectedClass)
								if err != nil {
									return operationResultMsg{err: fmt.Errorf("Error fetching students for class '%s': %w", m.selectedClass, err)}
								}
								styledOutput, err := getStudentsLatestCommitActivity(m.selectedClass, studentUsernames)
								if err != nil {
									return operationResultMsg{err: fmt.Errorf("Error fetching GitHub activity: %w", err)}
								}
								return operationResultMsg{logs: []string{styledOutput}} // Send as a single log entry
							},
						)
					case 2: // Back
						if len(m.menuHistory) > 0 {
							lastIndex := len(m.menuHistory) - 1
							m.state = m.menuHistory[lastIndex]
							m.menuHistory = m.menuHistory[:lastIndex]
							// Restore the correct list and title for the previous state
							switch m.state {
							case stateClassManagement:
								m.list = createClassManagementMenu(m.selectedClass)
								m.currentMenu = fmt.Sprintf("Managing Class: %s", m.selectedClass)
							}
							m.selectedItem = 0 // Reset selection for the previous menu
						} else {
							// Should not happen if navigating from a menu, but as a fallback:
							m.state = stateMainMenu
							mainMenuItems := []list.Item{
								Item{title: "Select Class", description: "Select an existing class to manage"},
								Item{title: "Create Class", description: "Create a new class"},
								Item{title: "Quit", description: "Exit the application"},
							}
							m.list.SetItems(mainMenuItems)
							m.list.Title = "Student Code Viewer" // Reset title
							m.currentMenu = "Main Menu"
						}
						return m, nil
					}
				}

				// Handle other key presses for list navigation etc. if not Enter
				// This part might already be handled by the general list update below
			} else if m.state == stateClassInput { // Note: stateClassInput is for creating a new class name
				// This was previously stateClassInput, ensure it's distinct from student input state
				m.className = m.classInput.Value()

				// Check if this is from the main menu's Create Class option
				if m.currentMenu == "Create Class" {
					err := createClass(m.className)
					if err != nil {
						m.err = err
						return m, nil
					}
					m.output = fmt.Sprintf("Added class: %s\n\nPress Enter to return to main menu.", m.className)

					// Set up to return to main menu after showing output
					// Don't save class input state in history, so we go directly to main menu
					m.state = stateOutput
					// m.currentMenu will be set by stateOutput's logic or can be set here if needed
					// Clear the menu history to ensure we go back to main menu
					m.menuHistory = []int{stateMainMenu}
					return m, nil
				}

				// Handle the legacy code paths for other input scenarios if any (removed for this refactor)
				// Legacy code has been removed as part of refactoring
			} else if m.state == stateDeleteConfirmation { // Handle class deletion confirmation
				if msg.String() == "enter" {
					// Get the currently selected item from the list
					selected, ok := m.list.SelectedItem().(Item)
					if !ok {
						return m, nil
					}
					
					// Check which option was selected
					if selected.title == "Yes, delete class" {
						// User confirmed deletion - delete the class and all students
						err := deleteClassAndStudents(m.selectedClass)
						if err != nil {
							m.err = err
							return m, nil
						}
						
						// Return directly to class selection menu instead of showing output
						m.state = stateClassSelection
						m.currentMenu = "Select a Class"
						
						// Refresh class list in case it changed
						classes, err := getClasses()
						if err != nil {
							m.err = err
							return m, nil
						}
						m.classList = classes
						
						// Set up class selection menu
						classItems := make([]list.Item, len(m.classList)+1) // +1 for Back option
						for i, className := range m.classList {
							classItems[i] = Item{title: className, description: "Select to manage this class"}
						}
						classItems[len(m.classList)] = Item{title: "Back", description: "Return to main menu"}
						m.list.SetItems(classItems)
						return m, nil
						
					} else if selected.title == "No, cancel" {
						// User cancelled deletion - go back to class management menu
						if len(m.menuHistory) > 0 {
							lastIndex := len(m.menuHistory) - 1
							m.state = m.menuHistory[lastIndex]
							m.menuHistory = m.menuHistory[:lastIndex]
							m.currentMenu = "Class Management: " + m.selectedClass
							
							// Reset the menu items for class management
							manageItems := []list.Item{
								Item{title: "Manage Students", description: "Add or remove students"},
								Item{title: "Manage Repos", description: "Clone, pull, or clean repositories"},
								Item{title: "View GH Activity", description: "Check student GitHub activity"},
								Item{title: "Delete Class", description: "Delete this class and its data"},
								Item{title: "Back", description: "Return to main menu"},
							}
							m.list.SetItems(manageItems)
							m.selectedItem = 3 // Reset to the Delete Class option
							return m, nil
						}
					}
				} else if msg.String() == "esc" {
					// Escape cancels and returns to previous menu
					if len(m.menuHistory) > 0 {
						lastIndex := len(m.menuHistory) - 1
						m.state = m.menuHistory[lastIndex]
						m.menuHistory = m.menuHistory[:lastIndex]
						m.currentMenu = "Class Management: " + m.selectedClass
						
						// Reset the menu items for class management
						manageItems := []list.Item{
							Item{title: "Manage Students", description: "Add or remove students"},
							Item{title: "Manage Repos", description: "Clone, pull, or clean repositories"},
							Item{title: "View GH Activity", description: "Check student GitHub activity"},
							Item{title: "Delete Class", description: "Delete this class and its data"},
							Item{title: "Back", description: "Return to main menu"},
						}
						m.list.SetItems(manageItems)
						m.selectedItem = 3 // Reset to the Delete Class option
						return m, nil
					}
				}
				// Let the list handle all other key presses (like arrow keys)
				return m, nil
			} else if m.state == stateStudentInput {
				// Process student input
				studentNames := strings.Fields(m.studentInput.Value())
				if len(studentNames) == 0 {
					// No student names entered, go back to previous menu
					if len(m.menuHistory) > 0 {
						lastIndex := len(m.menuHistory) - 1
						m.state = m.menuHistory[lastIndex]
						m.menuHistory = m.menuHistory[:lastIndex]
						return m, nil
					}
				}	// Add students to the database
				for _, name := range studentNames {
					err := addStudent(m.selectedClass, name)
					if err != nil {
						m.err = err
						return m, nil
					}
				}

				m.output = fmt.Sprintf("Added %d students to class %s", len(studentNames), m.selectedClass)
				m.menuHistory = append(m.menuHistory, m.state)
				m.state = stateOutput
				return m, nil
			} else if m.state == stateOutput {
				// Return to previous state from output display
				if len(m.menuHistory) > 0 {
					// Pop the last state from history
					lastIndex := len(m.menuHistory) - 1
					previousState := m.menuHistory[lastIndex]
					m.menuHistory = m.menuHistory[:lastIndex]
					m.state = previousState // The state is now restored

					// If returning to student input, clear it and set focus
					if m.state == stateStudentInput {
						m.currentMenu = "Add Students to: " + m.selectedClass // Restore title
						m.studentInput.SetValue("")                           // Clear input
						m.studentInput.Focus()                                // Re-focus
						return m, textinput.Blink                             // Return blink command
					} else {
						// Restore appropriate menu title and selected item for other states
						switch m.state {
						case stateMainMenu:
							m.currentMenu = "Main Menu"
						case stateClassSelection:
							m.currentMenu = "Select a Class"
						case stateClassManagement:
							m.currentMenu = "Managing Class: " + m.selectedClass
						case stateManageStudents:
							m.currentMenu = "Manage Students: " + m.selectedClass
						}
						m.selectedItem = 0 // Reset selection for menu states
					}
					return m, nil
				}
				// If no history, go to main menu
				m.state = stateMainMenu
				return m, nil
			}
		}

	// Add these cases to handle new messages for loading state
	case spinner.TickMsg: // For animating the spinner
		var cmd tea.Cmd
		if m.state == stateLoading {
			m.spinner, cmd = m.spinner.Update(msg)
		}
		return m, cmd

	case operationResultMsg: // Custom message type for when repo operations complete
		if msg.err != nil {
			m.output = fmt.Sprintf("Operation failed: %v", msg.err)
		} else {
			m.output = formatLogs(msg.logs)
		}
		// Even if an error occurred, we transition to stateOutput to show the message.
		// The menuHistory should have been set before transitioning to stateLoading.
		m.state = stateOutput
		// m.currentMenu will be set by the View logic or stateOutput transition logic
		// No need to Tick spinner anymore, operation is done.
		return m, nil

		// Handle other message types like textinput.BlurMsg if necessary
		// case textinput.BlurMsg:
		//    if msg.ID == m.classInput.ID() { ... }
	}

	var cmd tea.Cmd
	// Handle text input updates only if an input field is focused
	switch m.state {
	case stateMainMenu, stateClassSelection, stateClassManagement, stateManageStudents, stateManageRepos, stateStudentSelectionForDelete, stateDeleteConfirmation:
		m.list, cmd = m.list.Update(msg)
	case stateClassInput:
		m.classInput, cmd = m.classInput.Update(msg)
	case stateStudentInput:
		m.studentInput, cmd = m.studentInput.Update(msg)
	}

	return m, cmd
}

func (m Model) View() string {
	// Handle errors first
	if m.err != nil {
		errorMsg := fmt.Sprintf("%s\n\n%s", errorStyle.Render("Error:"), m.err.Error())
		m.err = nil // Clear the error after displaying it
		return docStyle.Render(errorMsg + "\n\nPress any key to continue.")
	}

	// Update breadcrumb path based on current state
	breadcrumbPath := getBreadcrumbPath(m, m.currentMenu)
	
	// Create breadcrumb navigation
	breadcrumb := buildBreadcrumb(breadcrumbPath)

	switch m.state {
	case stateMainMenu:
		// Use our new simple menu display with the selected item highlighted
		return docStyle.Render(breadcrumb + "\n\n" + createSimpleMenuWithSelection("Student Code Viewer", m.menuItems, m.selectedItem))

	case stateClassSelection:
		// Convert class list to menu items and add a Back option
		classItems := make([]Item, len(m.classList)+1) // +1 for Back option
		for i, className := range m.classList {
			classItems[i] = Item{title: className, description: "Select to manage this class"}
		}
		// Add Back option as the last item
		classItems[len(m.classList)] = Item{title: "Back", description: "Return to main menu"}
		return docStyle.Render(breadcrumb + "\n\n" + createSimpleMenuWithSelection("Select a Class", classItems, m.selectedItem))

	case stateClassManagement:
		// Create class management menu items
		manageItems := []Item{
			{title: "Manage Students", description: "Add or remove students"},
			{title: "Manage Repos", description: "Clone, pull, or clean repositories"},
			{title: "View GH Activity", description: "Check student GitHub activity"},
			{title: "Delete Class", description: "Delete this class and its data"},
			{title: "Back", description: "Return to main menu"},
		}
		return docStyle.Render(breadcrumb + "\n\n" + createSimpleMenuWithSelection("Managing Class: "+m.selectedClass, manageItems, m.selectedItem))

	case stateManageStudents:
		// Create manage students menu items
		studentManageItems := []Item{
			{title: "Add Student(s)", description: "Add new students to this class"},
			{title: "Delete Student", description: "Remove a student from this class"},
			{title: "Back", description: "Return to class management menu"},
		}
		return docStyle.Render(breadcrumb + "\n\n" + createSimpleMenuWithSelection("Manage Students: "+m.selectedClass, studentManageItems, m.selectedItem))

	case stateStudentSelectionForDelete:
		// Convert student list to menu items and add a Back option
		studentDeleteItems := make([]Item, len(m.studentList)+1) // +1 for Back option
		for i, studentName := range m.studentList {
			studentDeleteItems[i] = Item{title: studentName, description: "Select to delete this student"}
		}
		// Add Back option as the last item
		studentDeleteItems[len(m.studentList)] = Item{title: "Back", description: "Return to manage students menu"}
		return docStyle.Render(breadcrumb + "\n\n" + createSimpleMenuWithSelection("Delete Student from: "+m.selectedClass, studentDeleteItems, m.selectedItem))

	case stateManageRepos:
		// Create manage repositories menu items
		repoManageItems := []Item{
			{title: "Clone All Repos", description: "Clone all student repositories for this class"},
			{title: "Pull All Repos", description: "Pull updates for all student repositories"},
			{title: "Clean All Repos", description: "Remove all cloned repositories for this class"},
			{title: "Back", description: "Return to class management menu"},
		}
		return docStyle.Render(breadcrumb + "\n\n" + createSimpleMenuWithSelection("Manage Repositories: "+m.selectedClass, repoManageItems, m.selectedItem))

	case stateClassInput:
		return docStyle.Render(
			breadcrumb + "\n\n" +
				titleStyle.Render("Enter Class Name") + "\n\n" +
				m.classInput.View() + "\n\n" +
				helpStyle.Render("Press Enter to confirm or Esc to cancel"),
		)

	case stateStudentInput:
		return docStyle.Render(
			breadcrumb + "\n\n" +
				titleStyle.Render("Enter Student Usernames") + "\n" +
				"(Space-separated list of GitHub usernames)\n\n" +
				m.studentInput.View(),
		)

	case stateDeleteConfirmation:
		// Create confirmation menu items
		warningText := fmt.Sprintf(
			"Are you sure you want to delete class '%s' and ALL associated student data?\n" +
			"This action CANNOT be undone.",
			m.selectedClass,
		)
		
		// Set the list title to reflect the current operation
		m.list.Title = "Confirm Deletion"
		
		// Render using the list's built-in View method
		return docStyle.Render(
			breadcrumb + "\n\n" +
			errorStyle.Render(warningText) + "\n\n" +
			m.list.View(),
		)

	case stateLoading: // New view for loading state
		return docStyle.Render(fmt.Sprintf("%s\n\n%s %s\n\n%s",
			breadcrumb,
			m.spinner.View(),
			m.loadingMessage,
			secondaryTextStyle.Render("(Press Esc to attempt to cancel or Ctrl+C to quit)"), // Note: Cancellation isn't implemented yet
		))

	case stateOutput:
		// Check if the output is one of our special styled tables
		latestActivityMarker := "✨ GitHub Users & Their Last Commits ✨"
		weekViewActivityMarker := "✨ GitHub Commit Activity (Last 7 Work Days) ✨"

		if strings.Contains(m.output, latestActivityMarker) || strings.Contains(m.output, weekViewActivityMarker) {
			// For styled tables, display them directly without additional formatting
			return docStyle.Render(
				breadcrumb + "\n" +
				m.output + "\n" +
				helpStyle.Render("Press Enter to continue."),
			)
		} else {
			// For regular output, apply our styling
			return docStyle.Render(
				breadcrumb + "\n" +
				titleStyle.Render("Output") + "\n\n" +
				m.output + "\n\n" +
				helpStyle.Render("Press Enter to continue."),
			)
		}

	case stateViewGHActivity:
		// Create GitHub activity menu items
		ghActivityItems := []Item{
			{title: "Week View", description: "View student activity for the past week"},
			{title: "Check Latest Activity", description: "Display the latest commit time for each student"},
			{title: "Back", description: "Return to class management menu"},
		}
		return docStyle.Render(breadcrumb + "\n\n" + createSimpleMenuWithSelection("GitHub Activity for Class: "+m.selectedClass, ghActivityItems, m.selectedItem))

	default:
		return docStyle.Render(breadcrumb + "\n\nUnknown state")
	}
}

func formatDurationAgo(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(math.Mod(d.Hours(), 24))
	minutes := int(math.Mod(d.Minutes(), 60))

	if d < 0 {
		d = -d // Consider events in the future as 'from now'
		// For this app, 'ago' makes more sense, so negative durations might indicate an error or future timestamp
		// However, for dummy data, we ensure 'ago' by setting referenceNow after commit times.
		// If real data could be in the future, this logic might need adjustment or return "in X..."
	}

	if days > 0 {
		if hours > 0 {
			return fmt.Sprintf("%dd %dh ago", days, hours)
		}
		return fmt.Sprintf("%dd ago", days)
	}
	if hours > 0 {
		if minutes > 0 {
			return fmt.Sprintf("%dh %dm ago", hours, minutes)
		}
		return fmt.Sprintf("%dh ago", hours)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm ago", minutes)
	}
	seconds := int(math.Mod(d.Seconds(), 60))
	if seconds > 0 {
		return fmt.Sprintf("%ds ago", seconds)
	}
	return "just now"
}

func getStudentsLatestCommitActivity(className string, studentUsernames []string) (string, error) {
	startTime := time.Now()

	// Define styles inspired by the screenshot
	// Main container for the table content (not the overall app docStyle)
	tableContainerStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")) // Purple border

	// Header above the table: "Fetching GitHub commit data..."
	fetchingHeaderStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#E06BFF")). // Bright pink/purple
		PaddingBottom(1)

	// Title inside the table: "✨ GitHub Users & Their Last Commits ✨"
	tableTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFD700")). // Gold for sparkles and text
		Background(lipgloss.Color("#3A2D4C")). // Dark purple background
		Padding(0, 1).
		Width(50). // Approximate width to center text
		Align(lipgloss.Center).
		MarginBottom(1)

	// Column headers: "User", "Last Commit"
	columnHeaderStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF85C0")). // Pinkish
		Padding(0, 1)

	// Usernames
	userStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#87CEFA")) // Light sky blue

	// Timestamps
	timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#32CD32")) // Lime green

	// Error style
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")) // Red

	// Footer below the table: "Done! Fetched commit data..."
	doneFooterStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#32CD32")). // Lime green for checkmark and text
		PaddingTop(1)

	// Icons - just strings for now
	sparkleIcon := "✨"
	userIconPlaceholder := "❖"
	timeIcon := "🕒"

	// Actual icons from screenshot
	userIcons := []string{"🐙", "🚀", "✨", "🐱", "💎"}

	var sb strings.Builder

	// Header line
	sb.WriteString(fetchingHeaderStyle.Render(fmt.Sprintf("%s ║ Fetching GitHub commit data with sparkles ║ %s", sparkleIcon, sparkleIcon)))
	sb.WriteString("\n") // Newline after fetching header

	// Table content construction
	var tableContentSb strings.Builder
	tableContentSb.WriteString(tableTitleStyle.Render("✨ GitHub Users & Their Last Commits ✨"))
	tableContentSb.WriteString("\n")

	// Column Headers Row
	headerRow := lipgloss.JoinHorizontal(lipgloss.Top,
		columnHeaderStyle.Copy().Width(15).Render(userIconPlaceholder+" User"),
		columnHeaderStyle.Copy().Width(25).Render("Last Commit"),
	)
	tableContentSb.WriteString(headerRow)
	tableContentSb.WriteString("\n")
	// Divider line (simple one)
	tableContentSb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Render(strings.Repeat("─", 50)))
	tableContentSb.WriteString("\n")

	// Data Rows - Fetch real GitHub data
	if len(studentUsernames) == 0 {
		tableContentSb.WriteString(errorStyle.Render("No students found in class"))
		tableContentSb.WriteString("\n")
	} else {
		for i, studentName := range studentUsernames {
			icon := userIcons[i%len(userIcons)] // Cycle through icons
			
			// Try to fetch real commit data from .github.io repository
			repoName := studentName + ".github.io"
			latestCommit, err := getLatestCommitForRepo(studentName, repoName)
			
			var formattedTime string
			if err != nil {
				// If .github.io repo doesn't exist or has no commits, show error
				formattedTime = errorStyle.Render("No .github.io repo")
			} else {
				// Calculate time ago from the actual commit date
				durationAgo := time.Now().Sub(latestCommit.Commit.Author.Date)
				formattedTime = timeStyle.Render("Last push " + formatDurationAgo(durationAgo))
			}

			userDisplay := userStyle.Render(studentName)
			row := lipgloss.JoinHorizontal(lipgloss.Top,
				lipgloss.NewStyle().Width(15).Render(fmt.Sprintf("%s %s", icon, userDisplay)),
				lipgloss.NewStyle().Width(25).Render(fmt.Sprintf("%s %s", timeIcon, formattedTime)),
			)
			tableContentSb.WriteString(row)
			tableContentSb.WriteString("\n")
			if i < len(studentUsernames)-1 {
				tableContentSb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#444444")).Render(strings.Repeat("┈", 50)) + "\n")
			}
		}
	}

	// Table Footer (charm-github v1.0.0)
	tableFooterStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#AAAAAA")).
		Background(lipgloss.Color("#302640")). // Slightly lighter than table title bg
		Padding(0,1).
		Width(50).
		Align(lipgloss.Center).
		MarginTop(1)
	tableContentSb.WriteString(tableFooterStyle.Render("❖: *❖ charm-github v1.0.0 ❖*:❖"))
	tableContentSb.WriteString("\n")

	sb.WriteString(tableContainerStyle.Render(tableContentSb.String()))
	sb.WriteString("\n") // Newline after table container

	// Done Footer line
	elapsedTime := time.Since(startTime).Seconds()
	sb.WriteString(doneFooterStyle.Render(fmt.Sprintf("✓ Done! Fetched commit data with extra cuteness in %.2fs ✓", elapsedTime)))

	return sb.String(), nil
}

func getStudentsCommitWeekViewActivity(className string, studentUsernames []string) (string, error) {
	startTime := time.Now()

	// --- Styles ---
	// Table border style
	tableBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#000000"))
		
	// Title style for the table
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#000000")).
		Align(lipgloss.Center)

	// We're not using a fetching header in the minimalist design

	// Column widths
	usernameColWidth := 15
	dateColWidth := 10

	// Header cell style
	headerCellStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#000000")).
		Align(lipgloss.Center).
		Width(dateColWidth).
		Border(lipgloss.NormalBorder(), false, false, false, true). // Right border only
		BorderForeground(lipgloss.Color("#FFFFFF"))

	// Username header cell style (first column)
	usernameHeaderStyle := headerCellStyle.Copy().
		Width(usernameColWidth).
		Align(lipgloss.Center)

	// Regular cell style
	cellStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#000000")).
		Align(lipgloss.Center).
		Width(dateColWidth).
		Border(lipgloss.NormalBorder(), false, false, false, true). // Right border only
		BorderForeground(lipgloss.Color("#FFFFFF"))

	// Username cell style (first column)
	usernameCellStyle := cellStyle.Copy().
		Width(usernameColWidth).
		Align(lipgloss.Left).
		Padding(0, 1, 0, 0)

	// Row divider style
	rowDividerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#000000"))

	// Footer style for "Done!" message
	doneFooterStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#32CD32")). // Green text
		PaddingTop(1)

	// --- Data Generation ---
	// Use real student usernames or fallback to empty if none provided
	if len(studentUsernames) == 0 {
		studentUsernames = []string{} // No dummy data, show empty state
	}
	
	// Generate dates for the last 5 workdays
	now := time.Now()
	workDates := []time.Time{}
	daysBack := 0
	for len(workDates) < 5 {
		workDate := now.AddDate(0, 0, -daysBack)
		if workDate.Weekday() != time.Saturday && workDate.Weekday() != time.Sunday {
			workDates = append(workDates, workDate)
		}
		daysBack++
	}

	// Reverse the workDates slice to get the dates in chronological order
	for i, j := 0, len(workDates)-1; i < j; i, j = i+1, j-1 {
		workDates[i], workDates[j] = workDates[j], workDates[i]
	}

	// Fetch real commit data from GitHub
	activityData := make(map[string]map[string]bool) // username -> date -> has activity
	
	for _, username := range studentUsernames {
		activityData[username] = make(map[string]bool)
		repoName := username + ".github.io"
		
		// Get the date range for the week (start of first day to end of last day)
		since := workDates[0].Truncate(24 * time.Hour)
		until := workDates[len(workDates)-1].Add(24 * time.Hour).Truncate(24 * time.Hour)
		
		// Fetch commits for this date range
		commits, err := getCommitsInDateRange(username, repoName, since, until)
		if err != nil {
			// If repo doesn't exist or API fails, mark all days as false
			for _, date := range workDates {
				activityData[username][date.Format("2006-01-02")] = false
			}
			continue
		}
		
		// Create a set of dates that have commits
		commitDates := make(map[string]bool)
		for _, commit := range commits {
			commitDate := commit.Commit.Author.Date.Format("2006-01-02")
			commitDates[commitDate] = true
		}
		
		// Mark activity for each work day
		for _, date := range workDates {
			dateKey := date.Format("2006-01-02")
			activityData[username][dateKey] = commitDates[dateKey]
		}
	}

	// --- Assembly ---
	var sb strings.Builder

	// Create title with class name
	title := "GitHub Commit Activity - Last 5 Work Days"
	if className != "" {
		title = fmt.Sprintf("Class: %s - GitHub Commit Activity - Last 5 Work Days", className)
	}

	// Table header row
	headerRow := strings.Builder{}
	headerRow.WriteString(usernameHeaderStyle.Render("Username"))
	for _, date := range workDates {
		// Format as "Mon 05/12" to match reference image
		dateStr := date.Format("Mon") + " " + fmt.Sprintf("%02d/%02d", date.Month(), date.Day())
		headerRow.WriteString(headerCellStyle.Render(dateStr))
	}

	// Start building table rows with the title and header
	rows := []string{
		titleStyle.Render(title),
		"", // Empty line after title
		headerRow.String(),
	}

	// Add a row divider after the header
	divider := strings.Repeat("─", usernameColWidth + (dateColWidth * len(workDates)) + len(workDates))
	rows = append(rows, rowDividerStyle.Render(divider))

	// Data rows
	for _, username := range studentUsernames {
		if len(rows) > 20 { // Limit number of rows for display
			break
		}
		
		row := strings.Builder{}
		row.WriteString(usernameCellStyle.Render(username))
		
		for _, date := range workDates {
			dateKey := date.Format("2006-01-02")
			hasActivity := activityData[username][dateKey]
			
			var symbol string
			var symbolColor lipgloss.Style
			
			if hasActivity {
				symbol = "✓" // Check mark
				symbolColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")) // Green
			} else {
				symbol = "✗" // X mark
				symbolColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")) // Red
			}
			
			row.WriteString(cellStyle.Render(symbolColor.Render(symbol)))
		}
		
		rows = append(rows, row.String())
	}

	// Join all rows with newlines
	tableContent := strings.Join(rows, "\n")

	// Wrap in border
	sb.WriteString(tableBorderStyle.Render(tableContent))
	sb.WriteString("\n")
	sb.WriteString(doneFooterStyle.Render(fmt.Sprintf("✓ Done! Fetched weekly commit data in %.2fs ✓", time.Since(startTime).Seconds())))

	return sb.String(), nil
}

func deleteClassAndStudents(className string) error {
	// Get all students in the class first
	students, err := getStudents(className)
	if err != nil {
		return fmt.Errorf("failed to get students for class %s: %w", className, err)
	}

	// Delete each student from the class
	for _, student := range students {
		if err := deleteStudent(className, student); err != nil {
			return fmt.Errorf("failed to delete student %s from class %s: %w", student, className, err)
		}
	}

	// Now that all students are deleted, delete the class itself
	if err := deleteClass(className); err != nil {
		return fmt.Errorf("failed to delete class %s: %w", className, err)
	}

	return nil
}
