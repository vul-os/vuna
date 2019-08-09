
from bs4 import BeautifulSoup
import requests
from pprint import pprint

url = "https://www.trophyseeds.com/product/ethos-genetics-zsweet-inzanity/"

r = requests.get(url)
raw_html = r.content
soup = BeautifulSoup(raw_html, 'html.parser')

a = soup.find("div", {"class": "summary entry-summary"})

product_name = a.find("h1", {"class": "product_title entry-title"}).getText()
product_price = a.find("span", {"class": "woocommerce-Price-amount amount"}).getText()
product_stock = a.find("p", {"class": "stock in-stock"}).getText()
product_short_disc = a.find("div", {"class": "woocommerce-product-details__short-description"}).getText()
product_category_ = a.find("span", {"class": "posted_in"}).find("a")
product_category_link = product_category_['href']
product_category = product_category_.getText()

print(product_name, product_price, product_stock, product_short_disc, product_category_link, product_category)