package bi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"

	"cloud.google.com/go/bigquery"
	"github.com/gin-gonic/gin"
)

type BigQueryProcessor struct {
	Client *bigquery.Client
}

func NewBigQueryProcessor(client *bigquery.Client) *BigQueryProcessor {
	return &BigQueryProcessor{
		Client: client,
	}
}

func (bp *BigQueryProcessor) TemplateAndExecuteOne(c *gin.Context) {
	ctx := c.Request.Context()
	// Parse the request body
	var data map[string]interface{}
	err := json.NewDecoder(c.Request.Body).Decode(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse request body"})
		return
	}

	// Extract the name and template_dict from the data
	name, _ := data["name"].(string)
	templateDict, _ := data["template_dict"].(map[string]interface{})

	// Process the file contents
	fileContents := ProcessFile(name)
	if fileContents == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "File does not exist"})
		return
	}

	// Apply the template substitution
	tmpl, err := template.New("query").Parse(fileContents)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse query template"})
		return
	}

	var queryBuilder strings.Builder
	err = tmpl.Execute(&queryBuilder, templateDict)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute query template"})
		return
	}

	query := queryBuilder.String()
	fmt.Println(query)
	q := bp.Client.Query(query)
	// Run the query and print results when the query job is completed.
	job, err := q.Run(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Bigquery job run"})
		return
	}
	status, err := job.Wait(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Bigquery error run wait %s", status)})
		return
	}
	it, err := job.Read(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Bigquery job run"})
		return
	}
	data, columns, err := BqSQLToJSON(it)
	fmt.Println(data, columns)
	type ExecuteOneResponse struct {
		Columns []string    `json:"columns"`
		Data    interface{} `json:"data"`
		Error   interface{} `json:"error"`
	}
	response := ExecuteOneResponse{
		Data:    data,
		Columns: columns,
	}
	// Set the response headers
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, response)
}
