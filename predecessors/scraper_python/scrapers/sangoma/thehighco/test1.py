from bs4 import BeautifulSoup
import requests
from pprint import pprint

url = "https://thehighco.co.za/shop/"
r = requests.get(url)
raw_html = r.content
soup = BeautifulSoup(raw_html, 'html.parser')
soup = soup.find("div", {"class": "woocommerce-container"})
soup = soup.find("ul", {"class": "products clearfix products-3"})

soup = soup.findAll("li", recursive=False)

for i in soup:

    print(i.find('a'))
    print()









