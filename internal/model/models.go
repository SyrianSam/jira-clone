package model

import "time"

type User struct {
	ID       int
	Username string
	Password string
	Role     string
}

type Task struct {
	ID                        int       `form:"taskID"`
	Title                     string    `form:"title"`
	Description               string    `form:"description"`
	AssignedTo                int       `form:"assigned_to"`
	State                     string    `form:"etat"`
	FirstName                 string    `form:"first_name"`
	LastName                  string    `form:"last_name"`
	BirthDate                 string    `form:"birth_date"`
	Email                     string    `form:"email"`
	PostalCode                string    `form:"postal_code"`
	City                      string    `form:"city"`
	RegulatoryComplianceCheck bool      `form:"regulatory_compliance_check"`
	ContractCompliance        bool      `form:"contract_compliance"`
	TaskCreator               string    `form:"task_creator"`
	TaskResponsible           string    `form:"task_responsible"`
	Comments                  string    `form:"comments"`
	CreatedAt                 time.Time // No form tag needed as this will be set on the server-side
}
