package model

type User struct {
	ID       int
	Username string
	Password string
	Role     string
}

type Task struct {
	ID                        string `form:"taskID"`
	Title                     string `form:"title"`
	Description               string `form:"description"`
	AssignedTo                string `form:"assigned_to"`
	State                     string `form:"state"`
	FirstName                 string `form:"first_name"`
	LastName                  string `form:"last_name"`
	BirthDate                 string `form:"birth_date"`
	Email                     string `form:"email"`
	PostalCode                string `form:"postal_code"`
	City                      string `form:"city"`
	RegulatoryComplianceCheck string `form:"regulatory_compliance_check"`
	ContractCompliance        string `form:"contract_compliance"`
	TaskCreator               string `form:"task_creator"`
	TaskResponsible           string `form:"task_responsible"`
	Comments                  string `form:"comments"`
	Priority                  string `form:"priority"`
	CreditCard                string `form:"credit_card"`
	EstimantOrigine           string `form:"origin_estimator"`
	Project                   string `form:"project"`
	CreatedAt                 string // No form tag needed as this will be set on the server-side
	UpdatedAt                 string `form:"update_date"`
}
