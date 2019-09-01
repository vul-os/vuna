
from bs4 import BeautifulSoup
import requests
from pprint import pprint
url = "https://www.yuppiechef.com//geometric-gin.htm?id=34301&name=Geometric-Gin-Gin-750ml&PHPSESSID=01t3buecrgqnut0a319dv9gapn/"

r = requests.get(url)
raw_html = r.content
soup = BeautifulSoup(raw_html, 'html.parser')

a = soup.find("div", {"id": "product"})

product_category = a.find("div", {"class": "product_heading"}).find("a", {"class": "category"}).getText().strip()
product_name = a.find("div", {"class": "product_heading"}).find("h1").getText().strip()
product_price = a.find("span", {"itemprop": "price"}).getText().strip()

product_stock = a.find("div", {"class": "feedback__message"}).find("div", {"class": "feedback-title"}).getText().strip()


print(product_name, product_category, product_price, product_stock)