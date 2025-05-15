package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// initDB initializes the database connection and creates tables if they don't exist
func initDB() error {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Create the .scv directory if it doesn't exist
	scvDir := filepath.Join(homeDir, ".scv")
	if err := os.MkdirAll(scvDir, 0755); err != nil {
		return fmt.Errorf("failed to create .scv directory: %w", err)
	}

	// Open the database file
	dbPath := filepath.Join(scvDir, "scv.db")
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Create tables if they don't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS classes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE
		);
		CREATE TABLE IF NOT EXISTS students (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			class_id INTEGER,
			username TEXT,
			FOREIGN KEY (class_id) REFERENCES classes (id) ON DELETE CASCADE,
			UNIQUE(class_id, username)
		);
	`)
	if err != nil {
		db.Close()
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

// getClasses returns a list of all class names
func getClasses() ([]string, error) {
	rows, err := db.Query("SELECT name FROM classes ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("failed to query classes: %w", err)
	}
	defer rows.Close()

	var classes []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to scan class name: %w", err)
		}
		classes = append(classes, name)
	}

	return classes, nil
}

// createClass creates a new class with the given name
func createClass(name string) error {
	_, err := db.Exec("INSERT INTO classes (name) VALUES (?)", name)
	if err != nil {
		return fmt.Errorf("failed to create class: %w", err)
	}
	return nil
}

// deleteClass deletes a class by name
func deleteClass(name string) error {
	_, err := db.Exec("DELETE FROM classes WHERE name = ?", name)
	if err != nil {
		return fmt.Errorf("failed to delete class: %w", err)
	}
	return nil
}

// getStudents returns a list of student usernames for a given class
func getStudents(className string) ([]string, error) {
	rows, err := db.Query(`
		SELECT s.username FROM students s
		JOIN classes c ON s.class_id = c.id
		WHERE c.name = ?
		ORDER BY s.username
	`, className)
	if err != nil {
		return nil, fmt.Errorf("failed to query students: %w", err)
	}
	defer rows.Close()

	var students []string
	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			return nil, fmt.Errorf("failed to scan student username: %w", err)
		}
		students = append(students, username)
	}

	return students, nil
}

// addStudent adds a student to a class
func addStudent(className, username string) error {
	_, err := db.Exec(`
		INSERT INTO students (class_id, username)
		SELECT id, ? FROM classes WHERE name = ?
	`, username, className)
	if err != nil {
		return fmt.Errorf("failed to add student: %w", err)
	}
	return nil
}

// deleteStudent removes a student from a class
func deleteStudent(className, username string) error {
	_, err := db.Exec(`
		DELETE FROM students
		WHERE username = ? AND class_id = (
			SELECT id FROM classes WHERE name = ?
		)
	`, username, className)
	if err != nil {
		return fmt.Errorf("failed to delete student: %w", err)
	}
	return nil
}
