package model

type User struct {
	ID       int
	Username string
	Password string
	Role     string
}

type Task struct {
	ID          int
	Title       string
	Description string
	AssignedTo  int
}
