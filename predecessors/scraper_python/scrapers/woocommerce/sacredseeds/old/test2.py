from bs4 import BeautifulSoup
import requests
from pprint import pprint

url = "https://sacredseeds.co.za/product-category/dutch-passion"
pagenator_str = "page"
r = requests.get(url)
raw_html = r.content
soup = BeautifulSoup(raw_html, 'html.parser')

pages = 1
soup = soup.find("nav", {"class": "woocommerce-pagination"})
if soup is not None:
    pages = int(soup.find("ul", {"class": "page-numbers"}).findAll("li")[-2].find("a").getText().strip())

ret_list = []
links = []
for page in range(1, pages):
    r = requests.get("/".join([url, pagenator_str, str(page)]))
    raw_html = r.content
    soup = BeautifulSoup(raw_html, 'html.parser')
    soup = soup.find('main', {'id': 'main'}).find("ul", {"class": "products columns-3"}).findAll("li")


    for s in soup:
        name = s.find("h2").getText().strip()
        link = s.find("a")['href']
        ret_list.append({
            "name": name,
            "link": link
        })
        links.append(link)

print(ret_list)
print(links)