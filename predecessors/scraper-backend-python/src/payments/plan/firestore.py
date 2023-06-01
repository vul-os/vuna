class FirestorePlanManager():
    def __init__(self, db):
        self.db = db
        
    def create(self, plan_data):
        # Create a new plan document in Firestore
        doc_ref = self.db.collection('plans').document()
        doc_ref.set(plan_data)
        return doc_ref.id

    def get(self, plan_id):
        # Retrieve plan data from Firestore
        doc_ref = self.db.collection('plans').document(plan_id)
        doc = doc_ref.get()

        if doc.exists:
            return doc.to_dict()

        return None

    def update(self, plan_id, plan_data):
        # Update an existing plan document in Firestore
        doc_ref = self.db.collection('plans').document(plan_id)
        doc_ref.update(plan_data)

    def delete(self, plan_id):
        # Delete a plan document from Firestore
        doc_ref = self.db.collection('plans').document(plan_id)
        doc_ref.delete()


