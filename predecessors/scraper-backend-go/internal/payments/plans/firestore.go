package plans

import (
	"context"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type FirestorePlanManager struct {
	db *firestore.Client
}

func NewFirestorePlanManager(client *firestore.Client) *FirestorePlanManager {
	return &FirestorePlanManager{
		db: client,
	}
}

func (manager *FirestorePlanManager) Create(p Plan) (string, error) {
	docRef, _, err := manager.db.Collection("plans").Add(context.Background(), p)
	if err != nil {
		return "", err
	}

	return docRef.ID, nil
}

func (manager *FirestorePlanManager) Get(planID string) (*Plan, error) {
	docRef := manager.db.Collection("plans").Doc(planID)
	doc, err := docRef.Get(context.Background())
	if err != nil {
		if doc == nil {
			return nil, nil
		}
		return nil, err
	}

	var plan Plan
	err = doc.DataTo(&plan)
	if err != nil {
		return nil, err
	}

	plan.ID = docRef.ID
	return &plan, nil
}

func (manager *FirestorePlanManager) Update(planID string, p Plan) error {
	docRef := manager.db.Collection("plans").Doc(planID)
	_, err := docRef.Set(context.Background(), p, firestore.MergeAll)
	return err
}

func (manager *FirestorePlanManager) Delete(planID string) error {
	docRef := manager.db.Collection("plans").Doc(planID)
	_, err := docRef.Delete(context.Background())
	return err
}

func (manager *FirestorePlanManager) GetAll() ([]Plan, error) {
	iter := manager.db.Collection("plans").Documents(context.Background())
	var plans []Plan
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var plan Plan
		err = doc.DataTo(&plan)
		if err != nil {
			return nil, err
		}
		plan.ID = doc.Ref.ID
		plans = append(plans, plan)
	}
	return plans, nil
}
