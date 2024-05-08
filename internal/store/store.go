package store

import (
	"database/sql"
	"fmt"
	"jira-clone/internal/model"
	"log"
	"math/rand"

	"github.com/tealeg/xlsx"

	_ "github.com/lib/pq" // PostgreSQL driver
)

type Store struct {
	db *sql.DB // Include this to hold the database connection
}

func New(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

func NewDB(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("gira", dataSourceName)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func (s *Store) VerifyUser(username, password string) (*model.User, error) {
	var user model.User
	err := s.db.QueryRow("SELECT id, username, password , heirarchy FROM users WHERE username = $1", username).Scan(&user.ID, &user.Username, &user.Password, &user.Role)
	fmt.Println(user.ID)
	fmt.Println(user.Password)
	fmt.Println(user.Username)
	fmt.Println(user.Role)
	if err != nil {
		return nil, err
	}
	// Compare hashed password
	// err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	// if err != nil {
	// 	return nil, err
	// }
	return &user, nil
}

func (s *Store) CreateUser(username string, password string) error {

	log.Printf(username)
	_, err := s.db.Exec("INSERT INTO users (username, password, heirarchy) VALUES ($1, $2, 1)", username, password)
	if err != nil {
		return err
	}
	return nil
}
func (s *Store) GetUserById(userID int) (*model.User, error) {
	var user model.User
	err := s.db.QueryRow("SELECT id, username, password, heirarchy FROM users WHERE id = $1", userID).Scan(&user.ID, &user.Username, &user.Password, &user.Role)
	if err != nil {
		log.Printf("%s", err.Error())
		return nil, err
	}

	log.Printf(user.Role)
	log.Printf(user.Username)
	return &user, nil
}

func (s *Store) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	err := s.db.QueryRow("SELECT id, username, password, heirarchy FROM users WHERE username = $1", username).Scan(&user.ID, &user.Username, &user.Password, &user.Role)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Store) FindUsers(query string) ([]model.User, error) {
	var users []model.User
	query = "%" + query + "%" // Add % wildcards
	rows, err := s.db.Query(
		"SELECT id, username FROM users WHERE username ILIKE $1", query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var user model.User
		err := rows.Scan(&user.ID, &user.Username)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (s *Store) GetTasks() ([]model.Task, error) {
	var tasks []model.Task
	rows, err := s.db.Query("SELECT id, title, description, assigned_to, state, first_name, last_name, birth_date, email, postal_code, city, regulatory_compliance_check, contract_compliance, task_creator, task_responsible, comments, archived FROM tasks ")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var task model.Task
		if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.AssignedTo, &task.State, &task.FirstName, &task.LastName, &task.BirthDate, &task.Email, &task.PostalCode, &task.City, &task.RegulatoryComplianceCheck, &task.ContractCompliance, &task.TaskCreator, &task.TaskResponsible, &task.Comments, &task.Archived); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (s *Store) GetTaskByID(id int) (*model.Task, error) {
	var task model.Task
	err := s.db.QueryRow(`
		SELECT id, title, description, assigned_to, state, first_name, last_name, birth_date, email, postal_code, city, regulatory_compliance_check, contract_compliance, task_creator, task_responsible, comments, priority, credit_card, created_at, COALESCE(updated_at, ''), bank_account_number
		FROM tasks WHERE id = $1`, id).Scan(&task.ID, &task.Title, &task.Description, &task.AssignedTo, &task.State, &task.FirstName, &task.LastName, &task.BirthDate, &task.Email, &task.PostalCode, &task.City, &task.RegulatoryComplianceCheck, &task.ContractCompliance, &task.TaskCreator, &task.TaskResponsible, &task.Comments, &task.Priority, &task.CreditCard, &task.CreatedAt, &task.UpdatedAt, &task.BankAccountNumber)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *Store) FindTasksByName(name string) ([]*model.Task, error) {

	// Prepare a slice to hold the tasks
	var tasks []*model.Task
	// Use SQL LIKE operator to match task names that contain the provided name anywhere in the title
	// Use '%' wildcards to match any sequence of characters before and after the name
	rows, err := s.db.Query(`
        SELECT id, title, description, assigned_to, state, first_name, last_name, birth_date, email, postal_code, city, regulatory_compliance_check, contract_compliance, task_creator, task_responsible, comments, priority, credit_card, created_at, COALESCE(updated_at, ''), bank_account_number
        FROM tasks WHERE title LIKE '%' || $1 || '%' AND archived != '1'`, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate through all returned rows
	for rows.Next() {
		var task model.Task
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.AssignedTo, &task.State, &task.FirstName, &task.LastName, &task.BirthDate, &task.Email, &task.PostalCode, &task.City, &task.RegulatoryComplianceCheck, &task.ContractCompliance, &task.TaskCreator, &task.TaskResponsible, &task.Comments, &task.Priority, &task.CreditCard, &task.CreatedAt, &task.UpdatedAt, &task.BankAccountNumber)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, &task)
	}

	// Check for errors from iterating over rows
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *Store) SaveFieldOrder(userID int, fieldOrder string) error {
	// SQL statement that inserts a new row or updates it if there's a conflict on the user_id column
	query := `
    INSERT INTO user_preferences (user_id, field_order)
    VALUES ($1, $2)
    ON CONFLICT (user_id) DO UPDATE
    SET field_order = EXCLUDED.field_order
    `
	_, err := s.db.Exec(query, userID, fieldOrder)
	return err
}

func (s *Store) GetFieldOrder(userID int) (string, error) {
	defaultFieldOrderString := "title-section,state-section,credit-card-section,rib-section,contract-compliance-section,first-name-section,last-name-section,regulatory-check-section,bank-account-section,assigned-to-section,city-section,email-section,postal-code-section,priority-section,birth-date-section,created-at-section,last-modification-section"

	var fieldOrder string
	query := "SELECT field_order FROM user_preferences WHERE user_id = $1"

	err := s.db.QueryRow(query, userID).Scan(&fieldOrder)
	if err != nil {
		// If there's no record, we could consider returning a default order
		if err == sql.ErrNoRows {
			// Return default order if no specific order is stored
			return defaultFieldOrderString, nil
		}
		return "", err
	}

	return fieldOrder, nil
}

func (s *Store) CreateTask(task model.Task) error {
	log.Printf("task.BankAccountNumber")
	log.Printf(task.BankAccountNumber)
	query := `
        INSERT INTO tasks (
            title, description, assigned_to, state, first_name, last_name, 
            birth_date, email, postal_code, city, regulatory_compliance_check, 
            contract_compliance, task_creator, task_responsible, comments, 
            priority, credit_card, created_at, bank_account_number, archived
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
        RETURNING id`

	err := s.db.QueryRow(query, task.Title, task.Description, task.AssignedTo, task.State,
		task.FirstName, task.LastName, task.BirthDate, task.Email, task.PostalCode,
		task.City, task.RegulatoryComplianceCheck, task.ContractCompliance,
		task.TaskCreator, task.TaskResponsible, task.Comments, task.Priority,
		task.CreditCard, task.CreatedAt, task.BankAccountNumber, 0).Scan(&task.ID)

	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateTask(task *model.Task) error {
	log.Printf("update_at: %s", task.UpdatedAt)
	log.Printf("Title: %s", task.Title)
	log.Printf("ID: %d", task.ID)
	log.Printf("lastName: %s", task.LastName)
	query := `UPDATE tasks SET
        title=$1, description=$2, state=$3, first_name=$4, last_name=$5,
        birth_date=$6, email=$7, postal_code=$8, city=$9, 
        regulatory_compliance_check=$10, contract_compliance=$11, 
        task_creator=$12, task_responsible=$13, comments=$14, 
        priority=$15, credit_card=$16, updated_at=$17, bank_account_number=$18 WHERE id=$19`

	_, err := s.db.Exec(query, task.Title, task.Description, task.State,
		task.FirstName, task.LastName, task.BirthDate, task.Email,
		task.PostalCode, task.City, task.RegulatoryComplianceCheck,
		task.ContractCompliance, task.TaskCreator, task.TaskResponsible,
		task.Comments, task.Priority, task.CreditCard, task.UpdatedAt, task.BankAccountNumber, task.ID)

	if err != nil {
		log.Printf("the error in update Task: %v", err.Error())
		return err
	}

	return nil
}

func (s *Store) VerifyRegularity(firstName, lastName string) (bool, error) {
	// Load the XLSX file
	wb, err := xlsx.OpenFile("NOK.xlsx")
	if err != nil {
		return false, err // Handle error if the file cannot be opened
	}

	// Assume the relevant data is in the first sheet
	sheet := wb.Sheets[0]

	// Iterate through rows, assuming first row is the header
	for i, row := range sheet.Rows {
		if i == 0 {
			continue // Skip header row
		}
		if len(row.Cells) < 3 {
			continue // Ensure there are enough cells
		}

		// Read cells; error handling omitted for brevity
		prenom := row.Cells[0].String()
		nom := row.Cells[1].String()
		// statut := row.Cells[2].String()

		// Check if the row matches the provided first name and last name
		if prenom == firstName && nom == lastName {
			return false, nil
		}
	}

	// Return not found or default status if no match is found
	return true, nil
}
func (s *Store) GenerateTitle() string {
	var id int
	err := s.db.QueryRow("SELECT id FROM tasks ORDER BY id DESC LIMIT 1").Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			id = 0
		} else {
			return ""
		}
	}
	id++

	log.Printf("here is the generated ID: %d", id)
	return fmt.Sprintf("KBB-%d", id)
}

// GenerateBankAccountNumber generates a bank account number with a random
// country code prefix, which is one of "A", "B", "C", or "D".
func (s *Store) GenerateBankAccountNumber() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const n = 10
	countryCode := []byte{'A', 'B', 'C', 'D'}[rand.Intn(4)]
	b := make([]byte, n+1)
	b[0] = countryCode
	for i := 1; i < len(b); i++ {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func (s *Store) DeleteTask(id int) error {
	_, err := s.db.Exec("DELETE FROM tasks WHERE id = $1", id)
	return err
}

func (s *Store) ArchiveTask(id int) error {
	_, err := s.db.Exec("UPDATE tasks SET archived = 1 WHERE id = $1", id)
	return err
}
