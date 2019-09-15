from bs4 import BeautifulSoup
import requests
from pprint import pprint

url = "https://www.biltongandbudz.co.za/product/droewors-wheels-100g/"
r = requests.get(url)
raw_html = r.content
soup = BeautifulSoup(raw_html, 'html.parser')
soup = soup.find("p", {"class": "stock in-stock"})
print(r.request.headers)
print(soup)




