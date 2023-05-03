package orchestrator

import (
	"context"
	"fmt"
	"net/url"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	taskspb "cloud.google.com/go/cloudtasks/apiv2/cloudtaskspb"
)

// createTask creates a new task in your App Engine queue.
func createTask(projectID, locationID, queueID string, product orchestrator.Product) (*taskspb.Task, error) {
	// Create a new Cloud Tasks client instance.
	// See https://godoc.org/cloud.google.com/go/cloudtasks/apiv2
	ctx := context.Background()
	client, err := cloudtasks.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("NewClient: %v", err)
	}
	defer client.Close()

	// Build the Task queue path.
	queuePath := fmt.Sprintf("projects/%s/locations/%s/queues/%s", projectID, locationID, queueID)

	technology := url.QueryEscape(product.Technology)
	urlEncoded := url.QueryEscape(product.URL)
	urlStr := fmt.Sprintf("https://mycloudrundomain.com/?technology=%s&url=%s", technology, urlEncoded)

	// Build the Task payload.
	// https://godoc.org/google.golang.org/genproto/googleapis/cloud/tasks/v2#CreateTaskRequest
	req := &taskspb.CreateTaskRequest{
		Parent: queuePath,
		Task: &taskspb.Task{
			// https://godoc.org/google.golang.org/genproto/googleapis/cloud/tasks/v2#AppEngineHttpRequest
			MessageType: &taskspb.Task_AppEngineHttpRequest{
				AppEngineHttpRequest: &taskspb.AppEngineHttpRequest{
					HttpMethod: taskspb.HttpMethod_POST,
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					RelativeUri: urlStr,
				},
			},
			Name: fmt.Sprintf("product-task-%d", product.ID),
		},
	}

	// // Add a payload message if one is present.
	// req.Task.GetAppEngineHttpRequest().Body = []byte(message)

	createdTask, err := client.CreateTask(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("cloudtasks.CreateTask: %v", err)
	}

	return createdTask, nil
}
