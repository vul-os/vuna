package main

import (
	"log"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"

	"context"
)

func main() {
	// Create Firestore client
	ctx := context.Background()
	projectID := "your-project-id"
	credFilePath := "path/to/your/credentials.json"
	client, err := firestore.NewClient(ctx, projectID, option.WithCredentialsFile(credFilePath))
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	defer client.Close()

	// Create Paystack plan manager
	paystackManager := NewPaystackPlanManager()

	// Create Firestore plan manager
	firestoreManager := NewFirestorePlanManager(client)

	// Create Gin router
	router := gin.Default()

	// Plan API endpoints
	planAPI := PlanAPI{
		paystackManager:  paystackManager,
		firestoreManager: firestoreManager,
	}
	router.POST("/plans", planAPI.CreatePlan)
	router.GET("/plans/:planID", planAPI.GetPlan)
	router.PUT("/plans/:planID", planAPI.UpdatePlan)
	router.DELETE("/plans/:planID", planAPI.DeletePlan)

	// Subscription API endpoints
	subscriptionAPI := SubscriptionAPI{
		paystackManager:  paystackManager,
		firestoreManager: firestoreManager,
	}
	router.POST("/subscriptions", subscriptionAPI.CreateSubscription)
	router.GET("/subscriptions/:subscriptionID", subscriptionAPI.GetSubscription)
	router.PUT("/subscriptions/:subscriptionID", subscriptionAPI.UpdateSubscription)
	router.DELETE("/subscriptions/:subscriptionID", subscriptionAPI.DeleteSubscription)

	// Run the server
	port := ":8080"
	log.Printf("Server running on port %s", port)
	err = router.Run(port)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
