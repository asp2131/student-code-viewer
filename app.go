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
					// Placeholder: create the actual menu for ViewGHActivity later
					// For now, just set items or it will panic if list is empty on view
					m.list = createViewGHActivityMenu(m.selectedClass) // This function needs to be created
					return m, nil
				case 3: // Delete Class
					// TODO: Implement Delete Class confirmation and logic
					m.output = "Delete Class not yet implemented."
					m.menuHistory = append(m.menuHistory, m.state)
					m.state = stateOutput
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
					selectedOptionIndex := m.list.Index()

					// Options: 0: Week View, 1: Check Specific Activity, 2: Back
					switch selectedOptionIndex {
					case 0: // Week View
						m.output = fmt.Sprintf("Week View for '%s' not yet implemented.", m.selectedClass)
						m.menuHistory = append(m.menuHistory, m.state)
						m.state = stateOutput
						return m, nil
					case 1: // Check Specific Activity
						m.menuHistory = append(m.menuHistory, m.state)
						m.state = stateLoading
						m.loadingMessage = fmt.Sprintf("Fetching GitHub activity for class '%s'...", m.selectedClass)
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
				i, _ := m.list.SelectedItem().(Item)

				switch i.title {
				case "Add Students": //This case might be legacy/unreachable after refactor.
					m.menuHistory = append(m.menuHistory, m.state)
					m.studentInput.Focus()
					m.state = stateStudentInput
					return m, nil
				}
			} else if m.state == stateStudentInput {
				// Process student input
				studentNames := strings.Fields(m.studentInput.Value())
				if len(studentNames) == 0 {
					m.output = "No student names entered."
					m.menuHistory = append(m.menuHistory, m.state)
					m.state = stateOutput
					return m, nil
				}

				// Add students to the database
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
	case stateClassSelection, stateClassManagement, stateManageStudents, stateStudentSelectionForDelete, stateManageRepos:
		m.list, cmd = m.list.Update(msg)
	case stateClassInput:
		m.classInput, cmd = m.classInput.Update(msg)
	case stateStudentInput:
		m.studentInput, cmd = m.studentInput.Update(msg)
	case stateViewGHActivity:
		m.list, cmd = m.list.Update(msg)
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

	// Create breadcrumb navigation
	breadcrumb := breadcrumbStyle.Render(m.currentMenu)

	switch m.state {
	case stateMainMenu:
		// Use our new simple menu display with the selected item highlighted
		return docStyle.Render(createSimpleMenuWithSelection("Student Code Viewer", m.menuItems, m.selectedItem))

	case stateClassSelection:
		// Convert class list to menu items and add a Back option
		classItems := make([]Item, len(m.classList)+1) // +1 for Back option
		for i, className := range m.classList {
			classItems[i] = Item{title: className, description: "Select to manage this class"}
		}
		// Add Back option as the last item
		classItems[len(m.classList)] = Item{title: "Back", description: "Return to main menu"}
		return docStyle.Render(createSimpleMenuWithSelection("Select a Class", classItems, m.selectedItem))

	case stateClassManagement:
		// Create class management menu items
		manageItems := []Item{
			{title: "Manage Students", description: "Add or remove students"},
			{title: "Manage Repos", description: "Clone, pull, or clean repositories"},
			{title: "View GH Activity", description: "Check student GitHub activity"},
			{title: "Delete Class", description: "Delete this class and its data"},
			{title: "Back", description: "Return to main menu"},
		}
		return docStyle.Render(createSimpleMenuWithSelection("Managing Class: "+m.selectedClass, manageItems, m.selectedItem))

	case stateManageStudents:
		// Create manage students menu items
		studentManageItems := []Item{
			{title: "Add Student(s)", description: "Add new students to this class"},
			{title: "Delete Student", description: "Remove a student from this class"},
			{title: "Back", description: "Return to class management menu"},
		}
		return docStyle.Render(createSimpleMenuWithSelection("Manage Students: "+m.selectedClass, studentManageItems, m.selectedItem))

	case stateStudentSelectionForDelete:
		// Convert student list to menu items and add a Back option
		studentDeleteItems := make([]Item, len(m.studentList)+1) // +1 for Back option
		for i, studentName := range m.studentList {
			studentDeleteItems[i] = Item{title: studentName, description: "Select to delete this student"}
		}
		// Add Back option as the last item
		studentDeleteItems[len(m.studentList)] = Item{title: "Back", description: "Return to manage students menu"}
		return docStyle.Render(createSimpleMenuWithSelection("Delete Student from: "+m.selectedClass, studentDeleteItems, m.selectedItem))

	case stateManageRepos:
		// Create manage repositories menu items
		repoManageItems := []Item{
			{title: "Clone All Repos", description: "Clone all student repositories for this class"},
			{title: "Pull All Repos", description: "Pull updates for all student repositories"},
			{title: "Clean All Repos", description: "Remove all cloned repositories for this class"},
			{title: "Back", description: "Return to class management menu"},
		}
		return docStyle.Render(createSimpleMenuWithSelection("Manage Repositories: "+m.selectedClass, repoManageItems, m.selectedItem))

	case stateClassInput:
		return docStyle.Render(
			breadcrumb + "\n" +
				titleStyle.Render("Enter Class Name") + "\n\n" +
				m.classInput.View() + "\n\n" +
				helpStyle.Render("Press Enter to confirm or Esc to cancel"),
		)

	case stateStudentInput:
		return docStyle.Render(
			breadcrumb + "\n" +
				titleStyle.Render("Enter Student Usernames") + "\n" +
				"(Space-separated list of GitHub usernames)\n\n" +
				m.studentInput.View(),
		)

	case stateLoading: // New view for loading state
		breadcrumb := breadcrumbStyle.Render(m.currentMenu)
		return docStyle.Render(fmt.Sprintf("%s\n\n%s %s\n\n%s",
			breadcrumb,
			m.spinner.View(),
			m.loadingMessage,
			secondaryTextStyle.Render("(Press Esc to attempt to cancel or Ctrl+C to quit)"), // Note: Cancellation isn't implemented yet
		))

	case stateOutput:
		// Check if the output is our special styled table
		// The marker string is "✨ GitHub Users & Their Last Commits ✨"
		if strings.Contains(m.output, "✨ GitHub Users & Their Last Commits ✨") {
			// If it's the styled table, render it directly AFTER the breadcrumb, then the confirmation.
			return docStyle.Render(
				breadcrumbStyle.Render(m.currentMenu) + "\n" +
				m.output + // This is already fully styled
				"\n\n" + helpStyle.Render("Press Enter to continue."),
			)
		}
		// Otherwise, use the standard output box
		return docStyle.Render(
			breadcrumbStyle.Render(m.currentMenu) + "\n" +
				outputBoxStyle.Render(m.output+"\n\nPress Enter to continue."),
		)

	case stateViewGHActivity:
		// m.list is populated by createViewGHActivityMenu and contains the correct items and title.
		// m.list.Update(msg) in the Update function handles navigation (k/j, up/down).
		return docStyle.Render(m.list.View())

	default:
		return "Loading..."
	}
}

// formatDurationAgo converts a time.Duration into a string like "Xd Yh ago" or "Ym ago".
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

// getStudentsLatestCommitActivity fetches (dummy) GitHub activity for students in a class
// and returns a fully styled string ready for display.
//revive:disable-next-line:unused-parameter // className is for future use with actual API calls
func getStudentsLatestCommitActivity(className string, studentUsernames []string) (string, error) {
	startTime := time.Now()
	const timeLayout = "2006-01-02 15:04:05"                                     // Layout for parsing string dates
	referenceNow := time.Date(2023, 5, 16, 10, 0, 0, 0, time.UTC) // Fixed point for 'ago' calculation

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

	// Dummy data
	dummyCommits := []struct {
		user string
		icon string
		time string // Keep as string for initial definition, parse later
	}{
		{user: "octocat", icon: userIcons[0], time: "2023-05-12 14:32:21"},
		{user: "defunkt", icon: userIcons[1], time: "2023-05-14 09:17:43"},
		{user: "mojombo", icon: userIcons[2], time: "2023-05-10 22:01:53"},
		{user: "wycats", icon: userIcons[3], time: "2023-05-13 16:45:09"},
		{user: "dhh", icon: userIcons[4], time: "2023-05-15 11:23:37"},
	}

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

	// Data Rows
	if len(studentUsernames) == 0 { // If no actual students, use all dummy data
		for i, commitData := range dummyCommits {
			commitTime, err := time.Parse(timeLayout, commitData.time)
			var formattedTime string
			if err != nil {
				formattedTime = timeStyle.Render("invalid date")
			} else {
				durationAgo := referenceNow.Sub(commitTime)
				formattedTime = timeStyle.Render("Last push " + formatDurationAgo(durationAgo))
			}

			userDisplay := userStyle.Render(commitData.user)
			row := lipgloss.JoinHorizontal(lipgloss.Top,
				lipgloss.NewStyle().Width(15).Render(fmt.Sprintf("%s %s", commitData.icon, userDisplay)),
				lipgloss.NewStyle().Width(25).Render(fmt.Sprintf("%s %s", timeIcon, formattedTime)),
			)
			tableContentSb.WriteString(row)
			tableContentSb.WriteString("\n")
			// Add a faint divider between rows, except for the last one
			if i < len(dummyCommits)-1 {
				tableContentSb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#444444")).Render(strings.Repeat("┈", 50)) + "\n")
			}
		}
	} else { // Use actual student names with dummy icons/times if available
		for i, studentName := range studentUsernames {
			icon := userIcons[i%len(userIcons)] // Cycle through icons
			// Generate a dummy commit time relative to referenceNow
			dummyCommitTime := referenceNow.Add(-time.Duration((i*24)+(i*5)+2) * time.Hour).Add(-time.Duration(i*15) * time.Minute)
			durationAgo := referenceNow.Sub(dummyCommitTime)
			formattedTime := timeStyle.Render("Last push " + formatDurationAgo(durationAgo))

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
