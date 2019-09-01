
from bs4 import BeautifulSoup
import requests
from pprint import pprint

url = "https://thehighco.co.za/product/mothership-faberge-egg/"

r = requests.get(url)
raw_html = r.content
soup = BeautifulSoup(raw_html, 'html.parser')
product_stock__ = soup.find("input", {"class": "input-text qty text"})
print(product_stock__)

# print(product_name, product_price, product_stock, product_short_disc, product_category_link, product_category)