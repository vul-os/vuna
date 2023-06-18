package site

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
	"github.com/exolutionza/scraper-backend-go/internal/permissions/plan"

	"firebase.google.com/go/auth"
)

type SitePermissionService struct {
	client          *bigquery.Client
	userPlanService *plan.UserPlanService
}

func NewSitePermissionService(client *bigquery.Client, 
		userPlanService *plan.UserPlanService) *SitePermissionService {
	return &SitePermissionService{
		client:          client,
		userPlanService: userPlanService,
	}
}

func (s *SitePermissionService) UpdateSitePermissions(w http.ResponseWriter, 
		r *http.Request) {
	// Get the user ID from the authenticated user
	user := r.Context().Value("user").(*auth.Token)
	userID := user.UID

	// Parse request body
	var req struct {
		SiteIdentifiers []string `json:"site_identifiers"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	// Perform site permission updates
	maxProducts, err := s.userPlanService.GetMaxProductsForUser(userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get max products for user: %v", err), 
		http.StatusInternalServerError)
		return
	}

	productCount, err := s.getProductCountForSites(req.SiteIdentifiers)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get product count for sites: %v", err), 
			http.StatusInternalServerError)
		return
	}

	if productCount > maxProducts {
		http.Error(w, fmt.Sprintf("Product count %d exceeds max allowed products %d", 
			productCount, maxProducts), http.StatusBadRequest)
		return
	}

	if err := s.updateSitePermissionsTable(userID, req.SiteIdentifiers); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update site permissions table: %v", err),
			 http.StatusInternalServerError)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Site permissions updated successfully")
}

func (s *SitePermissionService) getProductCountForSites(siteIdentifiers []string) (int, error) {
	ctx := context.Background()

	query := s.client.Query(`
		SELECT COUNT(DISTINCT ProductID)
		FROM datapoint
		WHERE SiteIdentifier IN UNNEST(@siteIdentifiers)
			AND MaxQty > 0
	`)
	query.Parameters = []bigquery.QueryParameter{
		{
			Name:  "siteIdentifiers",
			Value: siteIdentifiers,
		},
	}

	it, err := query.Read(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to execute BigQuery query: %w", err)
	}

	var productCount int

	for {
		var count int
		err := it.Next(&count)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return 0, fmt.Errorf("failed to iterate query results: %w", err)
		}
		productCount += count
	}

	return productCount, nil
}

func (s *SitePermissionService) updateSitePermissionsTable(userID string, siteIdentifiers []string) error {
	ctx := context.Background()

	siteIds := strings.Join(siteIdentifiers, `","`)
	// Prepare SQL statement
	sql := fmt.Sprintf(`
		INSERT INTO site_permissions (user_id, site_identifiers)
		VALUES ("%s", ["%s"])
		ON DUPLICATE KEY UPDATE site_identifiers = ["%s"]
	`, userID, siteIds, siteIds)

	// Execute the SQL statement
	query := s.client.Query(sql)
	_, err := query.Run(ctx)
	if err != nil {
		return fmt.Errorf("failed to update site permissions table: %w", err)
	}

	return nil
}

