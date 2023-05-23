package main


func main() {
	// Create instances of the necessary dependencies (taskCreator, fileStorage, etc.)
	taskCreator := // Initialize your task creator
	fileStorage := // Initialize your file storage
	targetURL := // Set your target URL

	// Create an instance of OrchestratorAPI
	orchestratorAPI := New(taskCreator, fileStorage, targetURL)

	// Create a Gin router
	router := gin.Default()

	// Define your API routes
	router.GET("scraper/meta", orchestratorAPI.Meta)
	router.GET("scraper/site", orchestratorAPI.Site)

	// Run the server
	if err := router.Run(":8080"); err != nil {
		fmt.Println("Failed to start server:", err)
	}
}