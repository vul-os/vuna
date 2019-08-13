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

max_pages = int(d)

pagination_str = "?p="
page = 1

product_links = []

for i in range(page, max_pages):
    url_ = "".join([url, pagination_str, str(page)])
    print(f"Url for Page: {url_}")

    r = requests.get(url_)
    raw_html = r.content
    soup = BeautifulSoup(raw_html, 'html.parser')

    a = soup.find('div', {'id': 'center_column'})
    b = a.find('ul', {"class": "product_list grid row"})
    c = b.findAll('li')
    for d in c:
        data = d.find('a', {'class': 'product-name'})
        link = data['href'].strip()
        name = data.getText().strip()
        print(link)
        product_links.append(link)

    print(f"Completed Page: {page}")
    page += 1

with open("./text.txt", "w") as f:
    for p in product_links:
        f.write(f"{p} \n")

