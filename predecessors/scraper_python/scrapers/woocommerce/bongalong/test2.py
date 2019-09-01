
from bs4 import BeautifulSoup
import requests
from pprint import pprint
url = "https://bongalong.co.za/shop-2/blossom/"

r = requests.get(url)
raw_html = r.content
soup = BeautifulSoup(raw_html, 'html.parser')

a = soup.find("div", {"class": "summary entry-summary"})

print(a)

product_name = a.find("h1", {"class": "product_title entry-title"}).getText()
product_price = a.find("span", {"class": "woocommerce-Price-amount amount"}).getText()
product_stock = a.find("p", {"class": "stock in-stock"}).getText()

print(product_name, product_price, product_stock)