package handler

import (
	"jira-clone/internal/model"
	"jira-clone/internal/store"
	"log"
	"net/http"
	"strconv"
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
	router.GET("/task/details/:id", AuthRequired(), h.ShowTaskForm)
	router.POST("/task/new", AuthRequired(), h.CreateNewTask)
	router.GET("/task/showAll", AuthRequired(), h.ShowTasks)
	router.POST("/submit-task", AuthRequired(), h.submitTask)

}

func (h *Handler) ShowLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "login.templ", nil)
}

func (h *Handler) HandleLogin(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	user, err := h.store.VerifyUser(username, password)
	if err != nil {
		// Handle different kinds of errors appropriately
		c.HTML(http.StatusInternalServerError, "login.templ", gin.H{"Error": "Server error"})
		return
	}

	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Save()

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
		if userID == nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
		} else {
			c.Next()
		}
	}
}

func (h *Handler) ShowDashboard(c *gin.Context) {
	tasks, err := h.store.GetTasks()
	if err != nil {
		// Log the error and handle it appropriately
		log.Printf("Error retrieving tasks: %v", err)
		// Render an error message on the dashboard, or redirect to an error page
		c.HTML(http.StatusInternalServerError, "error.templ", gin.H{"Error": "Unable to retrieve tasks"})
		return
	}

	// If no error, render the dashboard with the tasks
	c.HTML(http.StatusOK, "dashboard.templ", gin.H{"tasks": tasks})
}

func (h *Handler) CreateNewTask(c *gin.Context) {

	c.HTML(http.StatusOK, "taskform.templ", nil)
}

func (h *Handler) ShowTaskForm(c *gin.Context) {

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
	c.HTML(http.StatusOK, "taskform.templ", gin.H{"Task": task})
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
		AssignedTo:  userID.(int), // Assuming task is initially assigned to the creator
	}
	h.store.CreateTask(task)

	// Redirect to the dashboard or another appropriate page
	c.Redirect(http.StatusFound, "/dashboard")
}

func (h *Handler) renderTaskList(c *gin.Context) {
	tasks, err := h.store.GetTasks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting tasks: " + err.Error()})
		return
	}

	if len(tasks) == 0 {
		c.HTML(http.StatusOK, "taskList.templ", gin.H{"message": "No tasks found."})
	} else {
		c.HTML(http.StatusOK, "taskList.templ", gin.H{"tasks": tasks})
	}
}

func (h *Handler) ShowTasks(c *gin.Context) {
	h.renderTaskList(c)
}

func (h *Handler) submitTask(c *gin.Context) {
	var task model.Task

	// Log specific form values
	log.Printf("Birthdate from form: %s", c.PostForm("birth_date"))
	log.Printf("Title from form: %s", c.PostForm("title"))

	if err := c.ShouldBind(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("Task Data: %+v", task)

	// Validate birth_date is not empty and is a valid date
	if _, err := time.Parse("2006-01-02", task.BirthDate); task.BirthDate != "" && err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid birth date"})
		return
	}

	if err := h.store.CreateTask(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create task: " + err.Error()})
		return
	}

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
