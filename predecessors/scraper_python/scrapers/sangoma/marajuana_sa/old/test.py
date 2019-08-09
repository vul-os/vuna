
from bs4 import BeautifulSoup
import requests
from pprint import pprint

url = "https://marijuanasa.co.za/"

r = requests.get(url)
raw_html = r.content
soup = BeautifulSoup(raw_html, 'html.parser')
print(soup)
soup = soup.find("ul", {"class": "product-categories"})
print(soup)
soup = soup.findAll("li")

print(soup)
