from requests import Session

session = Session()

# HEAD requests ask for *just* the headers, which is all you need to grab the
# session cookie
# session.head('https://www.shelflife.co.za/products/Montana-Limited-Edition-ENTES-Can')

import requests
import json

url = 'https://www.shelflife.co.za/ajax/prod_qty_dropdown.php'

# Adding empty header as parameters are being sent in payload
payload = {
    'prod': '7809',
    'size': 'L',
    'qty': '0'
}

#
#
response = session.post(url, data=payload)

print(response.text)