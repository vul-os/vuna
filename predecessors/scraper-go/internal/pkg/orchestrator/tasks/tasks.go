package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	cloudtaskspb "cloud.google.com/go/cloudtasks/apiv2/cloudtaskspb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

type TaskCreatorDetails struct {
	TargetUrl string
	ProjectID string
	Location  string
	QueueID   string
}

type TaskCreator struct {
	Client     *cloudtasks.Client
	DetailsMap map[string]TaskCreatorDetails
}

func New(detailsMap map[string]TaskCreatorDetails) (*TaskCreator, error) {
	ctx := context.Background()
	client, err := cloudtasks.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloud Tasks client: %w", err)
	}

	return &TaskCreator{
		Client:     client,
		DetailsMap: detailsMap,
	}, nil
}

func (t *TaskCreator) createTask(url string, key string) error {
	details := t.DetailsMap[key]

	parent := fmt.Sprintf("projects/%s/locations/%s/queues/%s",
		details.ProjectID, details.Location, details.QueueID)

	task := &cloudtaskspb.Task{
		MessageType: &cloudtaskspb.Task_HttpRequest{
			HttpRequest: &cloudtaskspb.HttpRequest{
				HttpMethod: cloudtaskspb.HttpMethod_GET,
				Url:        url,
			},
		},
	}

	req := &cloudtaskspb.CreateTaskRequest{
		Parent: parent,
		Task:   task,
	}

	if _, err := t.Client.CreateTask(context.Background(), req); err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	fmt.Println("Task created")
	return nil
}

func (t *TaskCreator) CreateTaskScrapeSite(url string) error {
	return t.createTask(fmt.Sprintf("%s/scraper/site/%s", t.DetailsMap["site"].TargetUrl, url), "site")
}

func (t *TaskCreator) CreateTaskScrapeMeta(url string) error {
	return t.createTask(fmt.Sprintf("%s/scraper/meta/%s", t.DetailsMap["meta"].TargetUrl, url), "meta")
}

func (t *TaskCreator) CreateTaskOrchestrateProduct(file string) error {
	return t.createTask(fmt.Sprintf("%s/orchestrator/product/%s", t.DetailsMap["orchestrateProduct"].TargetUrl, file), "orchestrateProduct")
}

func (t *TaskCreator) CreateTaskScrapeProduct(url string, scraper string, scheduledTime time.Time) error {
	taskURL := fmt.Sprintf("%s/scraper/product", t.DetailsMap["product"].TargetUrl)

	payload := map[string]interface{}{
		"url":         url,
		"scraper":     scraper,
		"full_scrape": "true",
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	details := t.DetailsMap["product"]

	parent := fmt.Sprintf("projects/%s/locations/%s/queues/%s",
		details.ProjectID, details.Location, details.QueueID)

	task := &cloudtaskspb.Task{
		MessageType: &cloudtaskspb.Task_HttpRequest{
			HttpRequest: &cloudtaskspb.HttpRequest{
				HttpMethod: cloudtaskspb.HttpMethod_POST,
				Url:        taskURL,
				Headers:    map[string]string{"Content-Type": "application/json"},
				Body:       jsonPayload,
			},
		},
		ScheduleTime: timestamppb.New(scheduledTime),
	}

	req := &cloudtaskspb.CreateTaskRequest{
		Parent: parent,
		Task:   task,
	}

	if _, err := t.Client.CreateTask(context.Background(), req); err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	fmt.Println("Task created")
	return nil
}
