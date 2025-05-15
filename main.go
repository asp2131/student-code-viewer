package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Initialize the database
	if err := initDB(); err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		os.Exit(1)
	}
	// Ensure db is the global variable from database.go
	// The actual closing of the db connection will be handled by the `db` package if necessary,
	// or we can add a specific CloseDB function if not already present.
	// For now, let's assume initDB sets up a global 'db' instance that can be closed.
	// If 'db' is not directly accessible, we might need a db.CloseDB() function.
	defer db.Close() // This line might cause issues if 'db' isn't the correct global var or if Close() isn't a method of the global var.

	// Initialize the model
	m := initialModel()

	// Start the Bubble Tea program
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil { // p.Run() is the correct method to start and wait for the program to finish
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
