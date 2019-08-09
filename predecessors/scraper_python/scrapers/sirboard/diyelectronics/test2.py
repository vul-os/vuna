from bs4 import BeautifulSoup
import requests
from pprint import pprint

url = "https://www.diyelectronics.co.za/store/199-electronics"
r = requests.get(url)
raw_html = r.content
soup = BeautifulSoup(raw_html, 'html.parser')
a = soup.find('ul', {'class': 'pagination'})
b = a.findAll('li')
c = b[-2]
d = c.find('span').getText()
print(d)