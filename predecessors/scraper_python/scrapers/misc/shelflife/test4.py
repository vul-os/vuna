from bs4 import BeautifulSoup
import requests
from requests import Session

from pprint import pprint

q_url = 'https://www.shelflife.co.za/ajax/prod_qty_dropdown.php'

url = "https://www.shelflife.co.za/products/SUICOKE-MOTO-Cab-Sandals-Black"
r = requests.get(url)

raw_html = r.content
soup = BeautifulSoup(raw_html, 'html.parser')

sku_id = soup.find("input", {"name": "prod"})['value']
soup = soup.find("select", {"id": "size"}).findAll("option")

options = []

for s in soup:
    if s['value'] is not "":
        options.append(s['value'])

stock = {}


for opt in options:

    session = Session()

    # HEAD requests ask for *just* the headers, which is all you need to grab the
    # session cookie
    # session.head('https://www.shelflife.co.za/products/Montana-Limited-Edition-ENTES-Can')

    import requests
    import json


    # Adding empty header as parameters are being sent in payload
    payload = {
        'prod': sku_id,
        'size': opt,
        'qty': '0'
    }

    #
    #
    response = session.post(q_url, data=payload)
    soup = BeautifulSoup(response.text, 'html.parser')

    stock[str(opt)] = soup.findAll('option')[-1]['value']

print(stock)