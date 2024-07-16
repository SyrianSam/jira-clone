package model

type User struct {
	ID       int
	Username string
	Password string
	Role     string
}

type UserPreference struct {
	UserID     int
	FieldOrder string
}

// Field represents a single form field with potential options for dropdowns, if applicable.
type Field struct {
	ID           string
	Label        string
	Type         string
	Options      []Option
	Value        string
	Placeholder  string
	Required     bool
	ReadOnly     bool
	ExtraHTML    string // To include additional HTML attributes like data attributes or JavaScript functions
	Action       string // For form actions, typically used with buttons
	Include      string // Specifies fields to include in an action, used with HTMX or AJAX
	Target       string // Defines where to place the response of an action
	Trigger      string // Event that triggers the action
	List         string
	HTMXGet      string // Assuming you're using a custom struct to handle HTMX attributes
	HTMXTrigger  string
	HTMXTarget   string
	AutoComplete string
}

type Option struct {
	Value    string
	Display  string
	Selected bool
}

type Task struct {
	ID                        int    `form:"taskID"`
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
	Rib                       string `form:"rib"`
	BankAccountNumber         string `form:"bank_account_number"`
	EstimantOrigine           string `form:"origin_estimator"`
	Project                   string `form:"project"`
	CreatedAt                 string `form:"create_date"` // No form tag needed as this will be set on the server-side
	UpdatedAt                 string `form:"update_date"`
	Archived                  string
	ParentTask                string
	Subtasks                  string
}
