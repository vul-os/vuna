from flask import Blueprint, request, jsonify, current_app
from src.payments.subscription.paystack import PaystackSubscriptionManager
from src.payments.subscription.firestore import FirestoreSubscriptionManager

subscriptions_bp = Blueprint('subscriptions', __name__)

firestore_db = current_app.config['CLIENT']
paystack_secret_key = current_app.config['PAYSTACK_SECRET_KEY']

firestore_manager = FirestoreSubscriptionManager(firestore_db)
paystack_manager = PaystackSubscriptionManager(paystack_secret_key)


@subscriptions_bp.route('/', methods=['POST'])
def create_subscription():
    data = request.get_json()
    customer_id = data.get('customer_id')
    plan_id = data.get('plan_id')

    # Create subscription in Paystack
    subscription_code = paystack_manager.create_subscription(customer_id, plan_id)

    # Save subscription in Firestore
    subscription_data = {
        'customer_id': customer_id,
        'plan_id': plan_id,
        'subscription_code': subscription_code
    }
    subscription_id = firestore_manager.create(subscription_data)

    return jsonify({'subscription_id': subscription_id, 'subscription_code': subscription_code}), 201

@subscriptions_bp.route('/<subscription_id>', methods=['GET'])
def get_subscription(subscription_id):
    # Retrieve subscription from Firestore
    subscription_data = firestore_manager.get(subscription_id)

    if subscription_data:
        return jsonify(subscription_data), 200
    else:
        return jsonify({'message': 'Subscription not found'}), 404

@subscriptions_bp.route('/<subscription_id>', methods=['PUT'])
def update_subscription(subscription_id):
    data = request.get_json()

    # Retrieve subscription from Firestore
    subscription_data = firestore_manager.get(subscription_id)

    if subscription_data:
        # Update subscription in Paystack
        paystack_manager.update_subscription(subscription_data['subscription_code'], customer_id=data.get('customer_id'), plan_id=data.get('plan_id'))

        # Update subscription in Firestore
        firestore_manager.update(subscription_id, data)

        return jsonify({'message': 'Subscription updated successfully'}), 200
    else:
        return jsonify({'message': 'Subscription not found'}), 404

@subscriptions_bp.route('/<subscription_id>', methods=['DELETE'])
def delete_subscription(subscription_id):
    # Retrieve subscription from Firestore
    subscription_data = firestore_manager.get(subscription_id)

    if subscription_data:
        # Delete subscription in Paystack
        paystack_manager.delete_subscription(subscription_data['subscription_code'])

        # Delete subscription in Firestore
        firestore_manager.delete(subscription_id)

        return jsonify({'message': 'Subscription deleted successfully'}), 200
    else:
        return jsonify({'message': 'Subscription not found'}), 404
