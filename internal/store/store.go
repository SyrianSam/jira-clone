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

func (s *Store) CreateTask(task model.Task) error {
	_, err := s.db.Exec("INSERT INTO tasks (title, description, assigned_to) VALUES ($1, $2, $3)", task.Title, task.Description, task.AssignedTo)
	return err
}

func (s *Store) GetTasks() ([]model.Task, error) {
	var tasks []model.Task
	rows, err := s.db.Query("SELECT id, title, description, assigned_to FROM tasks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var task model.Task
		if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.AssignedTo); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (s *Store) GetTaskByID(id int) (*model.Task, error) {
	var task model.Task
	err := s.db.QueryRow("SELECT id, title, description, assigned_to FROM tasks WHERE id = $1", id).Scan(&task.ID, &task.Title, &task.Description, &task.AssignedTo)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *Store) UpdateTask(task model.Task) error {
	_, err := s.db.Exec("UPDATE tasks SET title = $1, description = $2, assigned_to = $3 WHERE id = $4", task.Title, task.Description, task.AssignedTo, task.ID)
	return err
}

func (s *Store) DeleteTask(id int) error {
	_, err := s.db.Exec("DELETE FROM tasks WHERE id = $1", id)
	return err
}
