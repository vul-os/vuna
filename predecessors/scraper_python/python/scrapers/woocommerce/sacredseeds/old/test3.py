from bs4 import BeautifulSoup
import requests
from pprint import pprint

url = "https://sacredseeds.co.za/product/dutch-passion-auto-mazar-3-pack/"
pagenator_str = "page"
r = requests.get(url)
raw_html = r.content
soup = BeautifulSoup(raw_html, 'html.parser')
soup = soup.find("div", {"id": "content"}).find("div", {"class": "summary entry-summary"})

name = soup.find("h1", {"class": "product_title entry-title"}).getText()
id = soup.find("span", {"class": "sku_wrapper"}).find("span", {"class": "sku"}).getText()
price = soup.find("p", {"class": "price"}).find("span", {"class": "woocommerce-Price-amount amount"}).getText()
stock = soup.find("div", {"class": "quantity"}).find("input")['max']

print(name, id, price, stock)