package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	cloudtaskspb "cloud.google.com/go/cloudtasks/apiv2/cloudtaskspb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

type TaskCreator struct {
	Client    *cloudtasks.Client
	Parent    string
	TargetURL string
}

func New(
	projectID string,
	location string,
	queueID string,
	targetUrl string,
) (*TaskCreator, error) {
	ctx := context.Background()
	client, err := cloudtasks.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloud Tasks client: %v", err)
	}
	parent := fmt.Sprintf("projects/%s/locations/%s/queues/%s", projectID, location, queueID)
	return &TaskCreator{
		Client:    client,
		Parent:    parent,
		TargetURL: targetUrl,
	}, nil
}

func (t *TaskCreator) CreateTask(task *cloudtaskspb.Task) error {
	ctx := context.Background()
	req := &cloudtaskspb.CreateTaskRequest{
		Parent: t.Parent,
		Task:   task,
	}
	_, err := t.Client.CreateTask(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create task: %v", err)
	}
	fmt.Println("Task created")
	return nil
}

func (t *TaskCreator) CreateTaskScrapeSite(url string) error {
	task := &cloudtaskspb.Task{
		MessageType: &cloudtaskspb.Task_HttpRequest{
			HttpRequest: &cloudtaskspb.HttpRequest{
				HttpMethod: cloudtaskspb.HttpMethod_GET,
				Url:        fmt.Sprintf("%s/scraper/site/%s", t.TargetURL, url),
			},
		},
	}
	return t.CreateTask(task)
}

func (t *TaskCreator) CreateTaskScrapeMeta(url string) error {
	task := &cloudtaskspb.Task{
		MessageType: &cloudtaskspb.Task_HttpRequest{
			HttpRequest: &cloudtaskspb.HttpRequest{
				HttpMethod: cloudtaskspb.HttpMethod_GET,
				Url:        fmt.Sprintf("%s/scraper/meta/%s", t.TargetURL, url),
			},
		},
	}
	return t.CreateTask(task)
}

func (t *TaskCreator) CreateTaskOrchestrateProduct(file string) error {
	task := &cloudtaskspb.Task{
		MessageType: &cloudtaskspb.Task_HttpRequest{
			HttpRequest: &cloudtaskspb.HttpRequest{
				HttpMethod: cloudtaskspb.HttpMethod_GET,
				Url:        fmt.Sprintf("%s/orchestrator/product/%s", t.TargetURL, file),
			},
		},
	}
	return t.CreateTask(task)
}

func (t *TaskCreator) CreateTaskScrapeProduct(url string, scraper string, scheduledTime time.Time) error {
	tUrl := strings.ReplaceAll(t.TargetURL, "https://", "")
	taskURL := fmt.Sprintf("%s/scraper/product/", tUrl)

	payload := map[string]interface{}{
		"url":     url,
		"scraper": scraper,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

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

	return t.CreateTask(task)
}
