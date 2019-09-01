from bs4 import BeautifulSoup
import requests
from pprint import pprint

url = "https://www.trophyseeds.com/shop/"
r = requests.get(url)

pagination_str = "/page/"
page = 1

product_links = []

while True:
    url_ = "".join([url, pagination_str, str(page)])
    print(f"Url for Page: {url_}")

    r = requests.get(url_)
    if not r.status_code < 400:
        break

    raw_html = r.content
    soup = BeautifulSoup(raw_html, 'html.parser')

    a = soup.find('main', {'id': 'main'})
    b = a.find('ul', {'class': 'products columns-3'})

    c = b.findAll('li')
    if len(b) is 0:
        break

    for d in c:
        e = d.find("a")['href']
        product_links.append(e)
        print(e)

    print(f"Completed Page: {page}")
    page += 1

with open("./text.txt", "w") as f:
    for p in product_links:
        f.write(f"{p} \n")


