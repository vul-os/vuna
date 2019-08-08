from bs4 import BeautifulSoup
import requests
from pprint import pprint

url = "https://www.diyelectronics.co.za/store/199-electronics"
r = requests.get(url)
raw_html = r.content
soup = BeautifulSoup(raw_html, 'html.parser')
a = soup.find('div', {'id': 'center_column'})
b = a.find('ul', {"class": "product_list grid row"})
c = b.findAll('li')
for d in c:
    data = d.find('a', {'class': 'product-name'})
    link = data['href'].strip()
    name = data.getText().strip()

    print(link, name)
  
# print(b)