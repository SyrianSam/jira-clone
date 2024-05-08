package handler

import (
	"jira-clone/internal/model"
	"jira-clone/internal/store"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	store *store.Store
}

func NewHandler(store *store.Store) *Handler {
	return &Handler{
		store: store,
	}
}

func (h *Handler) SetupRoutes(router *gin.Engine) {
	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))

	// Set the root route
	router.GET("/", func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		if userID != nil {
			c.Redirect(http.StatusFound, "/dashboard")
		} else {
			c.Redirect(http.StatusFound, "/login")
		}
	})

	router.GET("/login", h.ShowLogin)
	router.POST("/login", h.HandleLogin)
	router.GET("/logout", h.HandleLogout)
	router.GET("/dashboard", AuthRequired(), h.ShowDashboard)
	router.GET("/task/details/:id", AuthRequired(), h.ShowTaskFormDynamic)
	// router.GET("/task/details/:id", AuthRequired(), h.ShowTaskForm)
	router.POST("/task/details/:id", AuthRequired(), h.UpdateTask)
	router.DELETE("/task/delete/:id", AuthRequired(), h.DeleteTask)
	router.POST("/task/archive/:id", AuthRequired(), h.ArchiveTask)
	router.POST("/task/new", AuthRequired(), h.CreateNewTask)
	router.GET("/task/showAll", AuthRequired(), h.ShowTasks)
	router.POST("/modify-task", AuthRequired(), h.ModifyTask)
	router.POST("/create-task", AuthRequired(), h.InsertTask)
	router.GET("/users", h.SearchUsers)
	router.GET("/create-account", h.CreateAccountRoute)
	router.POST("/create-account", h.CreateAccount)
	router.GET("/task/search", h.SearchTasks)
	router.POST("/bankaccount/generate", h.generateBankAccountNumber)
	router.POST("/verify-names", h.VerifyNames)
	router.POST("/save-field-order", h.SaveFieldOrder)
}
func (h *Handler) SaveFieldOrder(c *gin.Context) {
	session := sessions.Default(c)
	userID, ok := session.Get("user_id").(int)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found in session"})
		return
	}

	// Define the default field order
	defaultFieldOrderString := "title-section,state-section,credit-card-section,rib-section,contract-compliance-section,first-name-section,last-name-section,regulatory-check-section,bank-account-section,assigned-to-section,city-section,email-section,postal-code-section,priority-section,birth-date-section,created-at-section,last-modification-section"
	defaultOrderSlice := strings.Split(defaultFieldOrderString, ",")

	// Retrieve the current pinned fields from the database.
	pinnedFields, err := h.store.GetFieldOrder(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current field order"})
		return
	}

	// Convert pinned fields string to slice.
	pinnedFieldsSlice := strings.Split(pinnedFields, ",")

	// Get checked pin ID from the form.
	checkedPin := c.PostForm("divId")
	log.Printf("Checked divID: %v", checkedPin)

	// Determine whether the checked pin is already in the pinned fields.
	found := false
	for i, pin := range pinnedFieldsSlice {
		if pin == checkedPin {
			// If found, remove it from the slice.
			pinnedFieldsSlice = append(pinnedFieldsSlice[:i], pinnedFieldsSlice[i+1:]...)
			found = true
			break
		}
	}

	// If the checked pin is not found in the list, append it.
	if !found && checkedPin != "" {
		pinnedFieldsSlice = append(pinnedFieldsSlice, checkedPin)
	}

	// Update the pinned fields in the database with possibly updated list.
	updatedPinnedFields := strings.Join(pinnedFieldsSlice, ",")
	err = h.store.SaveFieldOrder(userID, updatedPinnedFields)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update field order"})
		return
	}

	// Determine remaining fields by excluding pinned fields from the default fields
	remainingFields := make([]string, 0)
	pinnedFieldsMap := sliceToMap(pinnedFieldsSlice)
	for _, field := range defaultOrderSlice {
		if !pinnedFieldsMap[field] {
			remainingFields = append(remainingFields, field)
		}
	}

	log.Printf("Pinned fields for display: %v", pinnedFields)
	log.Printf("Remaining fields for display: %v", remainingFields)

	// Retrieve task ID from the session and handle errors
	taskID, ok := session.Get("task_id").(int)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task ID not found in session"})
		return
	}

	// Load task details and handle errors
	task, err := h.store.GetTaskByID(taskID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"message": "Failed to load task"})
		return
	}

	// Rerender the task form with the updated field lists.
	c.HTML(http.StatusOK, "taskformDynamic.templ", gin.H{
		"Task":            task,
		"PinnedFields":    pinnedFieldsSlice,
		"RemainingFields": remainingFields,
	})
}

// Helper function to check if a slice contains a string
func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}

// Helper function to convert a slice to a map for quick lookup
func sliceToMap(slice []string) map[string]bool {
	result := make(map[string]bool)
	for _, item := range slice {
		result[item] = true
	}
	return result
}

func (h *Handler) VerifyNames(c *gin.Context) {
	firstName := c.PostForm("first_name")
	lastName := c.PostForm("last_name")

	verified, err := h.store.VerifyRegularity(firstName, lastName) // Assume this function checks the names
	if err != nil {
		// c.HTML(http.StatusBadRequest, "error.html", gin.H{"message": "Verification failed", "details": err.Error()})
		log.Printf("error verifying regularity: %v", err.Error())
		return
	}

	c.HTML(http.StatusOK, "compliance_check.templ", gin.H{"verified": verified})
}

func (h *Handler) SearchUsers(c *gin.Context) {
	q := c.Param("assigned_to")
	users, err := h.store.FindUsers(q)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to search users")
		return
	}
	c.HTML(http.StatusOK, "assigned_to_list.templ", gin.H{"users": users})
}

func (h *Handler) SearchTasks(c *gin.Context) {

	q := c.Query("searchTerm")

	q = strings.ToUpper(q)
	tasks, err := h.store.FindTasksByName(q)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to search tasks")
		return
	}

	c.HTML(http.StatusOK, "taskList.templ", gin.H{"tasks": tasks})

}

func (h *Handler) ShowLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "login.templ", nil)
}

func (h *Handler) HandleLogin(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	user, err := h.store.VerifyUser(username, password)
	if err != nil {
		log.Printf("error verifying user: %v", err)
		// Handle different kinds of errors appropriately
		c.HTML(http.StatusInternalServerError, "login.templ", gin.H{"Error": "Incorrect username or password"})
		return
	}
	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Set("userRole", user.Role)
	session.Save()

	log.Printf("user data: %v", user)

	c.Redirect(http.StatusFound, "/dashboard")
}

func (h *Handler) HandleLogout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.Redirect(http.StatusFound, "/login")
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		userRole := session.Get("userRole")
		log.Printf("authrequired userRole: %v", userRole)
		log.Printf("authrequired: %v", userID)
		if userID == nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
		} else {
			c.Next()
		}
	}
}

func (h *Handler) ShowDashboard(c *gin.Context) {

	log.Printf("What the hell is going on?")
	session := sessions.Default(c)
	userID := session.Get("user_id")
	userRole := session.Get("userRole")
	userIDVal, ok := session.Get("user_id").(int)
	if !ok {
		log.Printf("ok: %v", ok)
		return
	}

	log.Printf("userID val: %v", userIDVal)
	// userProfile, err := h.store.GetUserById(1)
	log.Printf("userRole.id: %v", userRole)

	tasks, err := h.store.GetTasks()
	if err != nil {
		// Log the error and handle it appropriately
		log.Printf("Error retrieving tasks: %v", err)
		// Render an error message on the dashboard, or redirect to an error page
		c.HTML(http.StatusInternalServerError, "error.templ", gin.H{"Error": "Unable to retrieve tasks"})
		return
	}

	filteredTasks, _ := h.filterTasklist(c, tasks, "all")
	// If no error, render the dashboard with the tasks
	c.HTML(http.StatusOK, "dashboard.templ", gin.H{"tasks": filteredTasks, "filter": "all", "userID": userID, "userRole": userRole})
}
func (h *Handler) filterTasklist(c *gin.Context, tasks []model.Task, filter string) ([]model.Task, error) {

	session := sessions.Default(c)
	userID := session.Get("user_id")

	var filteredTasks []model.Task
	for _, task := range tasks {
		log.Printf("task.AssignedTo")
		log.Printf(task.AssignedTo)
		log.Printf(task.Archived)
		switch filter {
		case "my":
			// Assuming there's a method c.User() to get current user and Task struct has an AssignedTo attribute
			assignedToId, _ := strconv.Atoi(strings.Split(task.AssignedTo, " - ")[0])
			if assignedToId == userID.(int) && task.Archived == "0" {
				filteredTasks = append(filteredTasks, task)
			}
		case "archived":
			// Assuming Task struct has an Archived attribute
			if task.Archived == "1" {
				filteredTasks = append(filteredTasks, task)
			}
		case "name":
			// You'll need to handle this differently, perhaps via another parameter for name filtering
		case "all":
			// No filtering needed, just append
			if task.Archived == "0" {
				filteredTasks = append(filteredTasks, task)
			}
		}
	}

	return filteredTasks, nil
}
func (h *Handler) renderTaskList(c *gin.Context, filter string) {

	tasks, err := h.store.GetTasks() // Assuming this fetches all tasks
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting tasks: " + err.Error()})
		return
	}

	var filteredTasks []model.Task

	session := sessions.Default(c)
	userID := session.Get("user_id")

	log.Printf("%v", userID)
	// Apply filtering based on the filter parameter
	for _, task := range tasks {
		log.Printf("task.AssignedTo")
		log.Printf(task.AssignedTo)
		log.Printf(task.Archived)
		switch filter {
		case "my":
			// Assuming there's a method c.User() to get current user and Task struct has an AssignedTo attribute
			assignedToId, _ := strconv.Atoi(strings.Split(task.AssignedTo, " - ")[0])
			if assignedToId == userID.(int) && task.Archived == "0" {
				filteredTasks = append(filteredTasks, task)
			}
		case "archived":
			// Assuming Task struct has an Archived attribute
			if task.Archived == "1" {
				filteredTasks = append(filteredTasks, task)
			}
		case "name":
			// You'll need to handle this differently, perhaps via another parameter for name filtering
		case "all":
			// No filtering needed, just append
			if task.Archived == "0" {
				filteredTasks = append(filteredTasks, task)
			}
		}
	}

	if len(filteredTasks) == 0 {
		c.HTML(http.StatusOK, "taskList.templ", gin.H{"message": "No tasks found."})
	} else {
		c.HTML(http.StatusOK, "taskList.templ", gin.H{"tasks": filteredTasks, "filter": filter, "userID": userID})
	}
}
func (h *Handler) CreateAccount(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	// log.Printf(username, password)

	err := h.store.CreateUser(username, password)
	if err != nil {
		// Handle different kinds of errors appropriately
		c.HTML(http.StatusInternalServerError, "login.templ", gin.H{"Error": "Server error"})
		return
	}

	c.Redirect(http.StatusFound, "/login")
}

func (h *Handler) CreateAccountRoute(c *gin.Context) {
	c.HTML(http.StatusOK, "create_account.templ", nil)

}

func (h *Handler) CreateNewTask(c *gin.Context) {

	c.HTML(http.StatusOK, "taskform_new.templ", nil)
}

func (h *Handler) ArchiveTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		log.Printf("Task ID not provided")
		c.HTML(http.StatusBadRequest, "error.templ", gin.H{"Error": "No task ID provided"})
		return
	}
	id, err := strconv.Atoi(taskID)
	if err != nil {
		log.Printf("Invalid task ID format: %v", err)
		c.HTML(http.StatusBadRequest, "error.templ", gin.H{"Error": "Invalid task ID format"})
		return
	}
	err = h.store.ArchiveTask(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to archive task"})
		return
	}
	c.Redirect(http.StatusSeeOther, "/task/showAll")
}

func (h *Handler) DeleteTask(c *gin.Context) {
	taskID := c.Param("id") // Get task ID from the URL parameter
	if taskID == "" {
		log.Printf("Task ID not provided")
		c.HTML(http.StatusBadRequest, "error.templ", gin.H{"Error": "No task ID provided"})
		return
	}
	// Convert the taskID to an integer
	id, err := strconv.Atoi(taskID)
	if err != nil {
		log.Printf("Invalid task ID format: %v", err)
		c.HTML(http.StatusBadRequest, "error.templ", gin.H{"Error": "Invalid task ID format"})
		return
	}
	err = h.store.DeleteTask(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to delete task"})
		return
	}
	// c.JSON(http.StatusOK, gin.H{"message": "Task deleted"})

	c.Redirect(http.StatusSeeOther, "/task/showAll")

}

func (h *Handler) ShowTaskFormDynamic(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id").(int)
	taskIDParam := c.Param("id") // Assumes task ID is a route parameter

	taskID, err := strconv.Atoi(taskIDParam)
	if err != nil {
		log.Printf("Invalid task ID format: %v", err)
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"message": "Invalid task ID format"})
		return
	}
	session.Set("task_id", taskID)
	session.Save()

	task, err := h.store.GetTaskByID(taskID) // Retrieve the task details
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"message": "Failed to load task"})
		return
	}

	order, err := h.store.GetFieldOrder(userID) // Retrieve field order for the user
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"message": "Failed to load field order"})
		return
	}

	defaultFieldOrderString := "title-section,state-section,credit-card-section,rib-section,contract-compliance-section,first-name-section,last-name-section,regulatory-check-section,bank-account-section,assigned-to-section,city-section,email-section,postal-code-section,priority-section,birth-date-section,created-at-section,last-modification-section"
	defaultFields := strings.Split(defaultFieldOrderString, ",")

	pinnedFields := strings.Split(order, ",")
	pinnedFieldsMap := make(map[string]bool)
	for _, field := range pinnedFields {
		pinnedFieldsMap[field] = true
	}

	remainingFields := make([]string, 0)
	for _, field := range defaultFields {
		if !pinnedFieldsMap[field] {
			remainingFields = append(remainingFields, field)
		}
	}

	log.Printf("Pinned fields for display: %v", pinnedFields)
	log.Printf("Remaining fields for display: %v", remainingFields)

	c.HTML(http.StatusOK, "taskformDynamic.templ", gin.H{
		"Task":            task,
		"PinnedFields":    pinnedFields,
		"RemainingFields": remainingFields,
	})
}

func (h *Handler) ShowTaskForm(c *gin.Context) {

	session := sessions.Default(c)
	userRole := session.Get("userRole")
	// Retrieve the task ID from the route parameter
	taskID := c.Param("id")
	if taskID == "" {
		log.Printf("Task ID not provided")
		c.HTML(http.StatusBadRequest, "error.templ", gin.H{"Error": "No task ID provided"})
		return
	}

	// Convert the taskID to an integer
	id, err := strconv.Atoi(taskID)
	if err != nil {
		log.Printf("Invalid task ID format: %v", err)
		c.HTML(http.StatusBadRequest, "error.templ", gin.H{"Error": "Invalid task ID format"})
		return
	}

	// Retrieve the task from the store by its ID
	task, err := h.store.GetTaskByID(id)
	if err != nil {
		// Log the error and handle it appropriately
		log.Printf("Error retrieving the task: %v", err)
		// Render an error message on the dashboard, or redirect to an error page
		c.HTML(http.StatusInternalServerError, "error.templ", gin.H{"Error": "Unable to retrieve task"})
		return
	}
	log.Printf("%s", task.State)
	// Render the task form with the task data
	c.HTML(http.StatusOK, "taskform.templ", gin.H{"Task": task, "UserRole": userRole})
}

func (h *Handler) CreateTask(c *gin.Context) {

	session := sessions.Default(c)
	userID := session.Get("user_id")
	if userID == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	title := c.PostForm("title")
	description := c.PostForm("description")

	// Validate input data
	if title == "" || description == "" {
		c.HTML(http.StatusBadRequest, "taskform.templ", gin.H{
			"Error":       "Title and description cannot be empty",
			"Title":       title,
			"Description": description,
		})
		return
	}

	// Create the task using the store
	task := model.Task{
		Title:       title,
		Description: description,
		AssignedTo:  userID.(string), // Assuming task is initially assigned to the creator
	}

	log.Printf(task.State)
	h.store.CreateTask(task)

	// Redirect to the dashboard or another appropriate page
	c.Redirect(http.StatusFound, "/dashboard")
}

func (h *Handler) ShowTasks(c *gin.Context) {
	// Fetch the 'filter' query parameter, default to "all" if not specified
	filter := c.DefaultQuery("filter", "all")

	// Log the received filter parameter for debugging
	log.Printf("Filter parameter received: %s", filter)

	// Call the renderTaskList function with the determined filter
	h.renderTaskList(c, filter)
}
func (h *Handler) generateBankAccountNumber(c *gin.Context) {
	// Generate the bank account number using the store
	accountNumber := h.store.GenerateBankAccountNumber()

	// Set the generated value as the value of the input field
	c.HTML(http.StatusOK, "ban.templ", gin.H{
		"BankAccountNumber": accountNumber,
	})
}
func (h *Handler) InsertTask(c *gin.Context) {
	var task model.Task

	// Log specific form values for debugging
	log.Printf("in InsertTask:")
	log.Printf("Credit card from form: %s", c.PostForm("credit_card"))
	log.Printf("Title from form: %s", c.PostForm("title"))
	log.Printf("Context keys: %v", c.Request.Header)
	for key, value := range c.Request.Header {
		log.Printf("%s: %s", key, value)
	}
	log.Printf("Form data: %v", c.Request.Form)
	for key, value := range c.Request.Form {
		log.Printf("%s: %s", key, value)
	}

	// Bind form data to task struct
	if err := c.ShouldBind(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error binding data: " + err.Error()})
		return
	}
	now := time.Now()
	task.CreatedAt = now.Format("2006-01-02 15:04:05")
	task.Title = h.store.GenerateTitle()
	// task.BankAccountNumber = h.store.GenerateBankAccountNumber()

	parts := strings.Split(task.AssignedTo, "-")
	if len(parts) != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid format for AssignedTo"})
		return
	}
	// task.AssignedTo = strings.TrimSpace(parts[0])

	log.Printf("Task Data: %+v", task)

	// Attempt to create the task using the data provided
	if err := h.store.CreateTask(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create task: " + err.Error()})
		return
	}

	// Redirect to a specific page after successful creation
	c.Redirect(http.StatusSeeOther, "/task/showAll")
}

func (h *Handler) UpdateTask(c *gin.Context) {
	var updates model.Task

	if err := c.ShouldBind(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error binding data: " + err.Error()})
		return
	}
	updates.ID, _ = strconv.Atoi(c.Param("id"))
	now := time.Now()
	updates.UpdatedAt = now.Format("2006-01-02 15:04:05")

	// Update the task in the database
	if err := h.store.UpdateTask(&updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to update task", "details": err.Error()})
		return
	}

	c.Redirect(http.StatusSeeOther, "/task/showAll")
}

func (h *Handler) ModifyTask(c *gin.Context) {
	var task model.Task

	// Log specific form values for debugging
	log.Printf("Credit card from form: %s", c.PostForm("credit_card"))
	log.Printf("Title from form: %s", c.PostForm("title"))

	// Bind form data to task struct
	if err := c.ShouldBind(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error binding data: " + err.Error()})
		return
	}
	// Attempt to create the task using the data provided
	if err := h.store.UpdateTask(&task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create task: " + err.Error()})
		return
	}

	// Redirect to a specific page after successful creation
	c.Redirect(http.StatusSeeOther, "/task/showAll")
}

func (h *Handler) EditTaskForm(c *gin.Context) {
	taskID := c.Param("id")
	taskIDint, _ := strconv.Atoi(taskID)
	task, err := h.store.GetTaskByID(taskIDint)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.templ", gin.H{"Error": "Unable to retrieve task"})
		return
	}
	c.HTML(http.StatusOK, "taskform.templ", gin.H{"Task": task})
}
