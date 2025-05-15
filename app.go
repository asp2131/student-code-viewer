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
			// Allow escaping from class input back to main menu
			if m.state == stateClassInput && m.currentMenu == "Create Class" {
				// Return to main menu
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
				// Handle class management menu selection
				managementOptions := []string{
					"Manage Students",
					"Manage Repos",
					"View GH Activity",
					"Delete Class",
					"Back",
				}
				
				if m.selectedItem < len(managementOptions) {
					selectedOption := managementOptions[m.selectedItem]
					switch selectedOption {
					case "Back":
						if len(m.menuHistory) > 0 {
							// Pop the last state from history
							lastIndex := len(m.menuHistory) - 1
							m.state = m.menuHistory[lastIndex]
							m.menuHistory = m.menuHistory[:lastIndex]

							// Reset to main menu
							items := []Item{
								{title: "Manage Classes", description: "Select and manage an existing class"},
								{title: "Create Class", description: "Create a new class"},
								{title: "Quit", description: "Exit the application"},
							}
							m.menuItems = items
							m.selectedItem = 0
							m.currentMenu = "Main Menu"
						}
						return m, nil

					// Implement other class management options here
					// For now, we'll just show a placeholder message
					default:
						m.output = fmt.Sprintf("Selected option: %s\nThis feature is coming soon!", selectedOption)
						m.menuHistory = append(m.menuHistory, m.state)
						m.state = stateOutput
						return m, nil
					}
				}
			} else if m.state == stateClassInput {
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
					m.state = m.menuHistory[lastIndex]
					m.menuHistory = m.menuHistory[:lastIndex]
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
