from bs4 import BeautifulSoup
import requests
from pprint import pprint
base_url = "https://www.shelflife.co.za/"
url = "https://www.shelflife.co.za/Online-store/headwear"
r = requests.get(url)

pagination_str = "?&page="
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

    a = soup.find('div', {'class': 'row push_both push_top push_bottom light_row'})

    c = soup.findAll('div', {"class": "col-xs-6 col-sm-3"})
    if len(soup) is 0:
        break

    if len(c) < 17:
        break

    for d in c:
        e = d.find("a")['href']
        if "products/" in str(e):
            product_links.append("".join([base_url, e]))
            print("".join([base_url, e]))

    print(f"Completed Page: {page}")
    page += 1

with open("./text.txt", "w") as f:
    for p in product_links:
        f.write(f"{p} \n")


