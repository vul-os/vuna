from flask import Flask, request, jsonify, Blueprint, current_app
from src.payments.plan.paystack import PaystackPlanManager
from src.payments.plan.firestore import FirestorePlanManager

plan_bp = Blueprint('plan_api', __name__)

firestore_db = current_app.config['CLIENT']
paystack_secret_key = current_app.config['PAYSTACK_SECRET_KEY']

firestore_manager = FirestorePlanManager(firestore_db)
paystack_manager = PaystackPlanManager(paystack_secret_key)


@plan_bp.route('/plans', methods=['POST'])
def create_plan():
    data = request.get_json()
    plan_name = data.get('name')
    plan_interval = data.get('interval')
    plan_amount = data.get('amount')

    # Create plan in Paystack
    plan_code = paystack_manager.create_plan(plan_name, plan_interval, plan_amount)

    # Save plan in Firestore
    plan_data = {
        'name': plan_name,
        'interval': plan_interval,
        'amount': plan_amount,
        'plan_code': plan_code
    }
    plan_id = firestore_manager.create(plan_data)

    return jsonify({'plan_id': plan_id, 'plan_code': plan_code}), 201

@plan_bp.route('/plans/<plan_id>', methods=['GET'])
def get_plan(plan_id):
    # Retrieve plan from Firestore
    plan_data = firestore_manager.get(plan_id)

    if plan_data:
        return jsonify(plan_data), 200
    else:
        return jsonify({'message': 'Plan not found'}), 404

@plan_bp.route('/plans/<plan_id>', methods=['PUT'])
def update_plan(plan_id):
    data = request.get_json()

    # Retrieve plan from Firestore
    plan_data = firestore_manager.get(plan_id)

    if plan_data:
        # Update plan in Paystack
        paystack_manager.update_plan(plan_data['plan_code'], name=data.get('name'), interval=data.get('interval'), amount=data.get('amount'))

        # Update plan in Firestore
        firestore_manager.update(plan_id, data)

        return jsonify({'message': 'Plan updated successfully'}), 200
    else:
        return jsonify({'message': 'Plan not found'}), 404

@plan_bp.route('/plans/<plan_id>', methods=['DELETE'])
def delete_plan(plan_id):
    # Retrieve plan from Firestore
    plan_data = firestore_manager.get(plan_id)

    if plan_data:
        # Delete plan in Paystack
        paystack_manager.delete_plan(plan_data['plan_code'])

        # Delete plan in Firestore
        firestore_manager.delete(plan_id)

        return jsonify({'message': 'Plan deleted successfully'}), 200
    else:
        return jsonify({'message': 'Plan not found'}), 404
