class PaystackPlanManager:
    def __init__(self, secret_key):
        self.secret_key = secret_key
        self.base_url = 'https://api.paystack.co/'

    def create_plan(self, name, interval, amount):
        url = self.base_url + 'plan'
        headers = {
            'Authorization': 'Bearer ' + self.secret_key,
            'Content-Type': 'application/json'
        }
        data = {
            'name': name,
            'interval': interval,
            'amount': amount
        }
        response = requests.post(url, headers=headers, json=data)
        response_data = response.json()
        plan_code = response_data.get('data', {}).get('plan_code')
        return plan_code

    def update_plan(self, plan_code, name=None, interval=None, amount=None):
        url = self.base_url + f'plan/{plan_code}'
        headers = {
            'Authorization': 'Bearer ' + self.secret_key,
            'Content-Type': 'application/json'
        }
        data = {}
        if name:
            data['name'] = name
        if interval:
            data['interval'] = interval
        if amount:
            data['amount'] = amount

        response = requests.put(url, headers=headers, json=data)
        return response.ok

    def delete_plan(self, plan_code):
        url = self.base_url + f'plan/{plan_code}'
        headers = {
            'Authorization': 'Bearer ' + self.secret_key
        }

        response = requests.delete(url, headers=headers)
        return response.ok
