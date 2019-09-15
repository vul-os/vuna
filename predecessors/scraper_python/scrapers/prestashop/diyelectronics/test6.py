from bs4 import BeautifulSoup
import requests
from pprint import pprint

url = "https://www.diyelectronics.co.za/store/7-3d-printing?id_category=7&n=817"

r = requests.get(url)
raw_html = r.content
soup = BeautifulSoup(raw_html, 'html.parser')

a = soup.find('div', {'id': 'center_column'})
b = a.find('ul', {"class": "product_list grid row"})
c = b.findAll('li')
product_links = []
for d in c:
    data = d.find('a', {'class': 'product-name'})
    if data is not None:
        link = data['href'].strip()
        name = data.getText().strip()
        print(link)
        product_links.append(link)

print(len(product_links))

