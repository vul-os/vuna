import requests

class PaystackSubscriptionManager:
    def __init__(self, secret_key):
        self.secret_key = secret_key
        self.base_url = 'https://api.paystack.co'

    def create_subscription(self, customer_id, plan_id):
        url = f'{self.base_url}/subscription'
        headers = {
            'Authorization': f'Bearer {self.secret_key}',
            'Content-Type': 'application/json'
        }
        data = {
            'customer': customer_id,
            'plan': plan_id
        }

        response = requests.post(url, headers=headers, json=data)
        if response.status_code == 201:
            subscription_data = response.json()
            return subscription_data['data']['subscription_code']
        else:
            return None

    def update_subscription(self, subscription_code, customer_id=None, plan_id=None):
        url = f'{self.base_url}/subscription/{subscription_code}'
        headers = {
            'Authorization': f'Bearer {self.secret_key}',
            'Content-Type': 'application/json'
        }
        data = {}
        if customer_id:
            data['customer'] = customer_id
        if plan_id:
            data['plan'] = plan_id

        response = requests.put(url, headers=headers, json=data)
        return response.status_code == 200

    def delete_subscription(self, subscription_code):
        url = f'{self.base_url}/subscription/{subscription_code}'
        headers = {
            'Authorization': f'Bearer {self.secret_key}'
        }

        response = requests.delete(url, headers=headers)
        return response.status_code == 204
