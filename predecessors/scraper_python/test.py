
from lxml import html
import requests

page = requests.get('https://www.fashionworld.co.za/products?page=30')

tree = html.fromstring(page.content)
# This will create a list of buyers:
products = tree.xpath('//div[data-component=listedProduct]>a[href]')
print(products)


import requests

