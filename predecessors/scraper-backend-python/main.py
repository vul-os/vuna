from flask import Flask
from src.bi.api import bi_bp
# from src.payments.plan.api import plan_bp
# from src.payments.subscription.api import subscriptions_bp

# cred = credentials.Certificate('path/to/service-account-key.json')  # Replace with your service account key path
# firebase_admin.initialize_app(cred)


# firestore_db = firestore.client()
# paystack_secret_key = 'YOUR_PAYSTACK_SECRET_KEY'  # Replace with your actual Paystack secret key


app = Flask(__name__)

# Register your blueprints
app.register_blueprint(bi_bp)
# app.register_blueprint(plan_bp)
# app.register_blueprint(subscription_bp)

# Add more blueprints if needed
# app.register_blueprint(another_bp)

def main(request):
    # Handle the incoming request using Flask app
    return app(request)


if __name__ == '__main__':
    # For local development
    app.run(host='localhost', port=8080, debug=True)