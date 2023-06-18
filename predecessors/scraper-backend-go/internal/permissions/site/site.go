package site

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"cloud.google.com/go/bigquery"
	scraperAuth "github.com/exolutionza/scraper-backend-go/internal/auth"
	"github.com/exolutionza/scraper-backend-go/internal/permissions/plan"
	"google.golang.org/api/iterator"
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
	user, ok := r.Context().Value("user").(scraperAuth.User)
	if !ok {
		http.Error(w, "Failed to retrieve user from context", http.StatusInternalServerError)
		return
	}
	userID := user.ID
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
	maxProducts, err := s.userPlanService.GetMaxProductsForUser(user.Email)
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
		SELECT
		COUNT(DISTINCT dp.ProductIdentifier) as Count
		FROM
		scrapers.datapoint_partitioned dp
		INNER JOIN
		scrapers.product_unique pu
		ON
		dp.ProductIdentifier = pu.ProductIdentifier
		WHERE
		pu.SiteIdentifier IN UNNEST(@siteIdentifiers)
		AND dp.MaxQty > 0
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

	type CountResult struct {
		Count int64 `bigquery:"count"`
	}

	for {
		var result CountResult
		err := it.Next(&result)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return 0, fmt.Errorf("failed to iterate query results: %w", err)
		}
		fmt.Println(result)
		productCount += int(result.Count)
	}

	return productCount, nil
}

func (s *SitePermissionService) updateSitePermissionsTable(userID string, siteIdentifiers []string) error {
	ctx := context.Background()

	// Prepare DELETE statement to delete existing rows of the user
	deleteSql := fmt.Sprintf(`
		DELETE FROM scrapers.site_permissions
		WHERE userid = "%s"`,
		userID)
	deleteQuery := s.client.Query(deleteSql)
	if _, err := deleteQuery.Run(ctx); err != nil {
		return fmt.Errorf("failed to delete old site permissions: %w", err)
	}

	// Prepare and run multiple INSERT statements to insert new rows for the user
	for _, siteID := range siteIdentifiers {
		insertSql := fmt.Sprintf(`
			INSERT INTO scrapers.site_permissions (userid, siteid)
			VALUES ("%s", "%s")`,
			userID, siteID)
		insertQuery := s.client.Query(insertSql)
		fmt.Println(insertQuery)
		if _, err := insertQuery.Run(ctx); err != nil {
			return fmt.Errorf("failed to insert new site permission: %w", err)
		}
	}

	return nil
}
