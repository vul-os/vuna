package subscriptions

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type SubscriptionAPI struct {
	firestoreManager FirestoreSubscriptionManager
	paystackManager  PaystackSubscriptionManager
}

func NewSubscriptionAPI(
	firestoreManager FirestoreSubscriptionManager,
	paystackManager PaystackSubscriptionManager,
) *SubscriptionAPI {
	return &SubscriptionAPI{
		firestoreManager: firestoreManager,
		paystackManager:  paystackManager,
	}
}

func (api *SubscriptionAPI) CreateSubscription(c *gin.Context) {
	var subscription Subscription
	err := c.BindJSON(&subscription)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse request body"})
		return
	}

	// Create subscription in Paystack
	subscriptionCode, err := api.paystackManager.CreateSubscription(subscription)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subscription in Paystack"})
		return
	}

	// Save subscription in Firestore
	subscriptionID, err := api.firestoreManager.Create(subscription)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save subscription in Firestore"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"subscription_id": subscriptionID, "subscription_code": subscriptionCode})
}

func (api *SubscriptionAPI) GetSubscription(c *gin.Context) {
	subscriptionID := c.Param("subscription_id")

	// Retrieve subscription from Firestore
	subscription, err := api.firestoreManager.Get(subscriptionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve subscription from Firestore"})
		return
	}
	if subscription == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

func (api *SubscriptionAPI) UpdateSubscription(c *gin.Context) {
	subscriptionID := c.Param("subscription_id")

	var subscription Subscription
	err := c.BindJSON(&subscription)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse request body"})
		return
	}

	// Retrieve subscription from Firestore
	existingSubscription, err := api.firestoreManager.Get(subscriptionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve subscription from Firestore"})
		return
	}
	if existingSubscription == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	// Update subscription in Paystack
	err = api.paystackManager.UpdateSubscription(subscriptionID, subscription)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription in Paystack"})
		return
	}

	// Update subscription in Firestore
	err = api.firestoreManager.Update(subscriptionID, subscription)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription in Firestore"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscription updated successfully"})
}

func (api *SubscriptionAPI) DeleteSubscription(c *gin.Context) {
	subscriptionID := c.Param("subscription_id")

	// Retrieve subscription from Firestore
	existingSubscription, err := api.firestoreManager.Get(subscriptionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve subscription from Firestore"})
		return
	}
	if existingSubscription == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	// Delete subscription in Paystack
	// 	err = api.paystackManager.DeleteSubscription(existingSubscription.SubscriptionCode)

	err = api.paystackManager.DeleteSubscription(existingSubscription.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete subscription in Paystack"})
		return
	}

	// Delete subscription in Firestore
	err = api.firestoreManager.Delete(subscriptionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete subscription in Firestore"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscription deleted successfully"})
}
