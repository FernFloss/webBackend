package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

// RunMigrations reads a SQL migration file and executes all statements against the database
// It handles multi-statement scripts, comments, and DO $$ blocks correctly
// The migration is idempotent (safe to run multiple times)
func RunMigrations(db *sql.DB, filePath string) error {
	log.Printf("Reading migration file: %s", filePath)

	// Read the entire SQL file
	sqlBytes, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read migration file %s: %w", filePath, err)
	}

	sqlContent := string(sqlBytes)
	log.Printf("Migration file read successfully (%d bytes)", len(sqlBytes))

	// Remove comments before processing
	sqlContent = removeComments(sqlContent)

	// Split SQL into individual statements
	statements := splitSQLStatements(sqlContent)

	if len(statements) == 0 {
		log.Println("No SQL statements found in migration file")
		return nil
	}

	log.Printf("Found %d SQL statements to execute", len(statements))

	// Execute each statement
	for i, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		// Log statement preview (first 100 chars)
		preview := stmt
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		log.Printf("Executing statement %d/%d: %s", i+1, len(statements), preview)

		// Execute the statement
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute statement %d/%d: %w\nStatement: %s", i+1, len(statements), err, preview)
		}
	}

	log.Println("All migration statements executed successfully")
	return nil
}

// removeComments removes SQL comments from the content
// Handles both -- single-line comments and /* */ multi-line comments
func removeComments(content string) string {
	// Remove /* */ style comments (multi-line)
	// This regex handles nested comments and comments spanning multiple lines
	multiLineCommentRegex := regexp.MustCompile(`/\*[\s\S]*?\*/`)
	content = multiLineCommentRegex.ReplaceAllString(content, "")

	// Remove -- style comments (single-line)
	// Split by newlines, process each line
	lines := strings.Split(content, "\n")
	var cleanedLines []string

	for _, line := range lines {
		// Find -- that's not inside a string literal
		// Simple approach: find first -- that's not inside quotes
		inSingleQuote := false
		inDoubleQuote := false
		commentStart := -1

		for i := 0; i < len(line); i++ {
			char := line[i]
			prevChar := byte(0)
			if i > 0 {
				prevChar = line[i-1]
			}

			// Handle escaped quotes
			if prevChar == '\\' {
				continue
			}

			if char == '\'' && !inDoubleQuote {
				inSingleQuote = !inSingleQuote
			} else if char == '"' && !inSingleQuote {
				inDoubleQuote = !inDoubleQuote
			} else if char == '-' && i < len(line)-1 && line[i+1] == '-' && !inSingleQuote && !inDoubleQuote {
				commentStart = i
				break
			}
		}

		if commentStart >= 0 {
			line = line[:commentStart]
		}

		cleanedLines = append(cleanedLines, line)
	}

	return strings.Join(cleanedLines, "\n")
}

// splitSQLStatements splits SQL content into individual statements
// Handles DO $$ blocks correctly (they contain semicolons but shouldn't be split)
func splitSQLStatements(content string) []string {
	var statements []string
	var current strings.Builder
	inDoBlock := false
	doBlockDelimiter := ""

	// Process character by character to handle DO blocks correctly
	contentRunes := []rune(content)
	i := 0

	for i < len(contentRunes) {
		char := contentRunes[i]

		// Check for DO block start
		if !inDoBlock && i+2 < len(contentRunes) {
			remaining := string(contentRunes[i:])
			if len(remaining) > 10 && strings.HasPrefix(strings.ToUpper(remaining), "DO") {
				// Check if followed by whitespace and delimiter
				doBlockRegex := regexp.MustCompile(`(?i)^DO\s+(\$\$|\$[a-zA-Z_][a-zA-Z0-9_]*\$|\$\w+\$)`)
				matches := doBlockRegex.FindStringSubmatch(remaining)
				if len(matches) > 1 {
					inDoBlock = true
					doBlockDelimiter = matches[1]
					// Write "DO " and delimiter to current
					current.WriteString(matches[0])
					i += len(matches[0])
					continue
				}
			}
		}

		// Check for DO block end
		if inDoBlock {
			remaining := string(contentRunes[i:])
			// Look for delimiter followed by semicolon (with optional whitespace)
			delimiterPattern := regexp.QuoteMeta(doBlockDelimiter) + `\s*;`
			delimiterRegex := regexp.MustCompile(`^` + delimiterPattern)
			if delimiterRegex.MatchString(remaining) {
				// Found end of DO block
				match := delimiterRegex.FindString(remaining)
				current.WriteString(match)
				i += len(match)
				inDoBlock = false
				doBlockDelimiter = ""
				// Continue to check if this ends a statement
				continue
			}
		}

		// Write current character
		current.WriteRune(char)

		// Check for statement end (semicolon) only if not in DO block
		if !inDoBlock && char == ';' {
			// Check if this semicolon is not inside a string
			stmt := strings.TrimSpace(current.String())
			if stmt != "" && stmt != ";" {
				statements = append(statements, stmt)
			}
			current.Reset()
		}

		i++
	}

	// Add any remaining statement
	stmt := strings.TrimSpace(current.String())
	if stmt != "" && stmt != ";" {
		statements = append(statements, stmt)
	}

	return statements
}

