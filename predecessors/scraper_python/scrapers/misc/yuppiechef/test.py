from bs4 import BeautifulSoup
import requests
from pprint import pprint
base_url = "https://www.yuppiechef.com/"
url = "https://www.yuppiechef.com/local-gin.htm"
r = requests.get(url)
raw_html = r.content
soup = BeautifulSoup(raw_html, 'html.parser')
last_page = soup.find("nav", {"class": "pagination-wrapper"})\
    .find("p", {"class": "yc-paragraph"}).find("a", {"class": "text-link"})['href'].split("=")[1].split("&")[0]
last_page = int(last_page)

pagination_str = "?page="

product_links = []

for page in range(1, last_page+1):
    url_ = "".join([url, pagination_str, str(page)])
    print(f"Url for Page: {url_}")

    r = requests.get(url_)
    if not r.status_code < 400:
        break

    raw_html = r.content
    soup = BeautifulSoup(raw_html, 'html.parser')

    a = soup.find('div', {'class': 'flex-grid-content'})
    b = a.find('div', {'class': 'flex-grid-group'})

    c = b.findAll('div', {"class": "flex-grid-block"})
    if len(b) is 0:
        break

    for d in c:
        e = d.find("a")['href']
        e = f"{base_url}/{e}"
        product_links.append(e)
        print(e)

    print(f"Completed Page: {page}")
    page += 1

with open("./text.txt", "w") as f:
    for p in product_links:
        f.write(f"{p} \n")


