import requests

class PaystackSubscriptionManager:
    def __init__(self, secret_key):
        self.secret_key = secret_key
        self.base_url = 'https://api.paystack.co/'

    def create_subscription(self, customer_id, plan_code):
        url = self.base_url + 'subscription'
        headers = {
            'Authorization': 'Bearer ' + self.secret_key,
            'Content-Type': 'application/json'
        }
        data = {
            'customer': customer_id,
            'plan': plan_code
        }
        response = requests.post(url, headers=headers, json=data)
        response_data = response.json()
        subscription_code = response_data.get('data', {}).get('subscription_code')
        return subscription_code

    def update_subscription(self, subscription_code, customer_id=None, plan_code=None):
        url = self.base_url + f'subscription/{subscription_code}'
        headers = {
            'Authorization': 'Bearer ' + self.secret_key,
            'Content-Type': 'application/json'
        }
        data = {}
        if customer_id:
            data['customer'] = customer_id
        if plan_code:
            data['plan'] = plan_code

        response = requests.put(url, headers=headers, json=data)
        return response.ok

    def cancel_subscription(self, subscription_code):
        url = self.base_url + f'subscription/{subscription_code}/disable'
        headers = {
            'Authorization': 'Bearer ' + self.secret_key
        }

        response = requests.post(url, headers=headers)
        return response.ok
