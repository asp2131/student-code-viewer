package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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
	studentInput.Placeholder = "Enter student GitHub usernames (space-separated)"
	studentInput.Focus()
	studentInput.CharLimit = 200
	studentInput.Width = 40

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
		menuHistory:   []int{},
		classList:     []string{},
		selectedClass: "",
		currentMenu:   "Main Menu",
		menuItems:     menuItems,
		studentList:   []string{}, // Initialize studentList
		selectedItem:  0,
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
					// TODO: Transition to stateManageRepos
					m.output = "Manage Repos not yet implemented."
					m.menuHistory = append(m.menuHistory, m.state)
					m.state = stateOutput
					return m, nil
				case 2: // View GH Activity
					// TODO: Transition to stateViewActivity
					m.output = "View GH Activity not yet implemented."
					m.menuHistory = append(m.menuHistory, m.state)
					m.state = stateOutput
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
						m.selectedItem = 0             // Reset selection
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

	// Handle window resize
	case tea.WindowSizeMsg:
		// Update list height
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	// Update the list model or text inputs
	var cmd tea.Cmd
	switch m.state {
	case stateClassSelection, stateClassManagement:
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
		return docStyle.Render(errorMsg + "\n\nPress Enter/Esc to continue.")
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

	case stateOutput:
		return docStyle.Render(
			breadcrumbStyle.Render(m.currentMenu) + "\n" +
				outputBoxStyle.Render(m.output + "\n\nPress Enter to continue."),
		)

	default:
		return "Loading..."
	}
}
