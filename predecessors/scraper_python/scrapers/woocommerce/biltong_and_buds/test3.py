from bs4 import BeautifulSoup
import requests
import json
from pprint import pprint

url = "https://www.biltongandbudz.co.za/product/orange-bud-feminized-1-seed-multiple/"
r = requests.get(url)
raw_html = r.content
soup = BeautifulSoup(raw_html, 'html.parser')

product_name = soup.find("h1", {"class": "product-title product_title entry-title"}).getText().strip()

variations = soup.find('table', {'class': 'variations'})
if variations is not None:
    json_data = json.loads(str(soup.find('form', {'class': 'variations_form cart'})['data-product_variations']))

    for data in json_data:
        variation_id = data['variation_id']
        pack_size = data['attributes']['attribute_pa_pack-size'].replace("-", "").replace("seeds", "").strip()
        stock = data['availability_html'].replace('<p class="stock in-stock">', '').replace('</p>', '').strip().replace('in', '').replace('stock', '').strip()
        price = data['display_price']
        print(variation_id, pack_size, stock, price)
