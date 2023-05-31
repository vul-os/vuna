class FirestoreManagerSubscription():
    def __init__(self, db):
        self.db = db
        
    def create(self, subscription_data):
        # Create a new subscription document in Firestore
        doc_ref = self.db.collection('subscriptions').document()
        doc_ref.set(subscription_data)
        return doc_ref.id

    def get(self, subscription_id):
        # Retrieve subscription data from Firestore
        doc_ref = self.db.collection('subscriptions').document(subscription_id)
        doc = doc_ref.get()

        if doc.exists:
            return doc.to_dict()

        return None

    def update(self, subscription_id, subscription_data):
        # Update an existing subscription document in Firestore
        doc_ref = self.db.collection('subscriptions').document(subscription_id)
        doc_ref.update(subscription_data)

    def delete(self, subscription_id):
        # Delete a subscription document from Firestore
        doc_ref = self.db.collection('subscriptions').document(subscription_id)
        doc_ref.delete()