from bs4 import BeautifulSoup
import requests
from pprint import pprint

url = "https://www.netram.co.za/4483-palette-2-pro-multi-material-filament-system.html"
r = requests.get(url)
raw_html = r.content
soup = BeautifulSoup(raw_html, 'html.parser')

a = soup.find('p', {'id': 'pQuantityAvailable'})
product_price = soup.find("span", {"id": "our_price_display"})['content']

print(float(product_price))

