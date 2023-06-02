package subscriptions

import (
	"context"

	"cloud.google.com/go/firestore"
)

type FirestoreSubscriptionManager struct {
	db *firestore.Client
}

func NewFirestoreSubscriptionManager(client *firestore.Client) *FirestoreSubscriptionManager {
	return &FirestoreSubscriptionManager{
		db: client,
	}
}

func (manager *FirestoreSubscriptionManager) Create(subscription Subscription) (string, error) {
	docRef, _, err := manager.db.Collection("subscriptions").Add(context.Background(), subscription)
	if err != nil {
		return "", err
	}

	return docRef.ID, nil
}

func (manager *FirestoreSubscriptionManager) Get(subscriptionID string) (*Subscription, error) {
	docRef := manager.db.Collection("subscriptions").Doc(subscriptionID)
	doc, err := docRef.Get(context.Background())
	if err != nil {
		if doc == nil {
			return nil, nil
		}
		return nil, err
	}

	var subscription Subscription
	err = doc.DataTo(&subscription)
	if err != nil {
		return nil, err
	}

	subscription.ID = docRef.ID
	return &subscription, nil
}

func (manager *FirestoreSubscriptionManager) Update(subscriptionID string, subscription Subscription) error {
	docRef := manager.db.Collection("subscriptions").Doc(subscriptionID)
	_, err := docRef.Set(context.Background(), subscription, firestore.MergeAll)
	return err
}

func (manager *FirestoreSubscriptionManager) Delete(subscriptionID string) error {
	docRef := manager.db.Collection("subscriptions").Doc(subscriptionID)
	_, err := docRef.Delete(context.Background())
	return err
}
