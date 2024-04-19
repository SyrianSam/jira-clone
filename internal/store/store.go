package store

import (
	"database/sql"
	"fmt"
	"jira-clone/internal/model"

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
	err := s.db.QueryRow("SELECT id, username, password FROM users WHERE username = $1", username).Scan(&user.ID, &user.Username, &user.Password)
	fmt.Println(user.ID)
	fmt.Println(user.Password)
	fmt.Println(user.Username)
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

func (s *Store) CreateUser(user model.User) error {
	_, err := s.db.Exec("INSERT INTO users (username, password, role) VALUES ($1, $2, $3)", user.Username, user.Password, user.Role)
	return err
}

func (s *Store) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	err := s.db.QueryRow("SELECT id, username, password, role FROM users WHERE username = $1", username).Scan(&user.ID, &user.Username, &user.Password, &user.Role)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Store) GetTasks() ([]model.Task, error) {
	var tasks []model.Task
	rows, err := s.db.Query("SELECT id, title, description, assigned_to, state, first_name, last_name, birth_date, email, postal_code, city, regulatory_compliance_check, contract_compliance, task_creator, task_responsible, comments FROM tasks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var task model.Task
		if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.AssignedTo, &task.State, &task.FirstName, &task.LastName, &task.BirthDate, &task.Email, &task.PostalCode, &task.City, &task.RegulatoryComplianceCheck, &task.ContractCompliance, &task.TaskCreator, &task.TaskResponsible, &task.Comments); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (s *Store) GetTaskByID(id int) (*model.Task, error) {
	var task model.Task
	err := s.db.QueryRow(`
		SELECT id, title, description, assigned_to, state, first_name, last_name, birth_date, email, postal_code, city, regulatory_compliance_check, contract_compliance, task_creator, task_responsible, comments 
		FROM tasks WHERE id = $1`, id).Scan(&task.ID, &task.Title, &task.Description, &task.AssignedTo, &task.State, &task.FirstName, &task.LastName, &task.BirthDate, &task.Email, &task.PostalCode, &task.City, &task.RegulatoryComplianceCheck, &task.ContractCompliance, &task.TaskCreator, &task.TaskResponsible, &task.Comments)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *Store) CreateTask(task model.Task) error {
	_, err := s.db.Exec(`
        INSERT INTO tasks (
            title, description, assigned_to, state, first_name, last_name, 
            birth_date, email, postal_code, city, regulatory_compliance_check, 
            contract_compliance, task_creator, task_responsible, comments, 
            priority, credit_card, estimant_origine, project, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)`,
		task.Title, task.Description, task.AssignedTo, task.State, task.FirstName, task.LastName,
		task.BirthDate, task.Email, task.PostalCode, task.City, task.RegulatoryComplianceCheck,
		task.ContractCompliance, task.TaskCreator, task.TaskResponsible, task.Comments,
		task.Priority, task.CreditCard, task.EstimantOrigine, task.Project, task.CreatedAt, task.UpdatedAt)
	return err
}

func (s *Store) UpdateTask(task *model.Task) error {
	query := `UPDATE tasks SET title=$1, description=$2, state=$3 WHERE id=$4`

	_, err := s.db.Exec(query, task.Title, task.Description, task.State, task.ID)
	return err
}

// func (s *Store) UpdateTask2(task *model.Task) error {
// 	// Starting the query construction
// 	query := "UPDATE tasks SET "
// 	params := []interface{}{}
// 	paramID := 1

// 	// Append each field to the query only if it's not the zero value of the field
// 	if task.Title != "" {
// 		query += fmt.Sprintf("title = $%d, ", paramID)
// 		params = append(params, task.Title)
// 		paramID++
// 	}
// 	if task.Description != "" {
// 		query += fmt.Sprintf("description = $%d, ", paramID)
// 		params = append(params, task.Description)
// 		paramID++
// 	}
// 	if task.State != "" {
// 		query += fmt.Sprintf("state = $%d, ", paramID)
// 		params = append(params, task.State)
// 		paramID++
// 	}
// 	// Continue for other fields...
// 	if task.FirstName != "" {
// 		query += fmt.Sprintf("first_name = $%d, ", paramID)
// 		params = append(params, task.FirstName)
// 		paramID++
// 	}
// 	if task.LastName != "" {
// 		query += fmt.Sprintf("last_name = $%d, ", paramID)
// 		params = append(params, task.LastName)
// 		paramID++
// 	}
// 	// Additional fields would be handled here similarly...

// 	// Handle boolean and date fields (only update if needed)
// 	// This assumes the frontend or service ensures correct values or uses defaults
// 	if task.RegulatoryComplianceCheck { // Assuming we want to update even if it's false
// 		query += fmt.Sprintf("regulatory_compliance_check = $%d, ", paramID)
// 		params = append(params, task.RegulatoryComplianceCheck)
// 		paramID++
// 	}
// 	if task.ContractCompliance { // Same assumption as above
// 		query += fmt.Sprintf("contract_compliance = $%d, ", paramID)
// 		params = append(params, task.ContractCompliance)
// 		paramID++
// 	}

// 	// Always update 'updated_at' field to current timestamp
// 	query += fmt.Sprintf("updated_at = $%d ", paramID)
// 	params = append(params, time.Now())
// 	paramID++

// 	// Ensure the query ends correctly with the WHERE clause
// 	query = strings.TrimSuffix(query, ", ") // Remove the last comma from the last added field
// 	query += fmt.Sprintf(" WHERE id = $%d", paramID)
// 	params = append(params, task.ID)
// 	log.Printf("%v", query)
// 	log.Printf("%v", params)
// 	// Execute the query
// 	_, err := s.db.Exec(query, params...)
// 	return err
// }

func (s *Store) DeleteTask(id int) error {
	_, err := s.db.Exec("DELETE FROM tasks WHERE id = $1", id)
	return err
}
