package plans

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type PlanAPI struct {
	firestoreManager FirestorePlanManager
	paystackManager  PaystackPlanManager
}

func NewPlanAPI(
		firestoreManager FirestorePlanManager, 
		paystackManager PaystackPlanManager,
	) *PlanAPI {
	return &PlanAPI{
		firestoreManager: firestoreManager,
		paystackManager:  paystackManager,
	}
}

func (api *PlanAPI) CreatePlan(c *gin.Context) {
	var plan Plan
	err := c.BindJSON(&plan)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse request body"})
		return
	}

	// Create plan in Paystack
	paystackPlan, err := api.paystackManager.CreatePlan(plan)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create plan in Paystack"})
		return
	}

	// Save plan in Firestore
	planID, err := api.firestoreManager.Create(plan)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save plan in Firestore"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"plan_id": planID, "paystack_plan": paystackPlan})
}

func (api *PlanAPI) GetPlan(c *gin.Context) {
	planID := c.Param("plan_id")

	// Retrieve plan from Firestore
	plan, err := api.firestoreManager.Get(planID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve plan from Firestore"})
		return
	}
	if plan == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plan not found"})
		return
	}

	c.JSON(http.StatusOK, plan)
}

func (api *PlanAPI) UpdatePlan(c *gin.Context) {
	planID := c.Param("plan_id")

	var plan Plan
	err := c.BindJSON(&plan)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse request body"})
		return
	}

	// Retrieve plan from Firestore
	existingPlan, err := api.firestoreManager.Get(planID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve plan from Firestore"})
		return
	}
	if existingPlan == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plan not found"})
		return
	}

	// Update plan in Paystack
	err = api.paystackManager.UpdatePlan(planID, plan)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update plan in Paystack"})
		return
	}

	// Update plan in Firestore
	err = api.firestoreManager.Update(planID, plan)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update plan in Firestore"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Plan updated successfully"})
}

func (api *PlanAPI) DeletePlan(c *gin.Context) {
	planID := c.Param("plan_id")

	// Retrieve plan from Firestore
	existingPlan, err := api.firestoreManager.Get(planID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve plan from Firestore"})
		return
	}
	if existingPlan == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plan not found"})
		return
	}

	// Delete plan in Paystack
	err = api.paystackManager.DeletePlan(planID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete plan in Paystack"})
		return
	}

	// Delete plan in Firestore
	err = api.firestoreManager.Delete(planID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete plan in Firestore"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Plan deleted successfully"})
}
