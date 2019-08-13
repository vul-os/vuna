from bs4 import BeautifulSoup
import requests
from pprint import pprint

url = "https://thehighco.co.za/product-category/dabbing"
r = requests.get(url)
raw_html = r.content
soup = BeautifulSoup(raw_html, 'html.parser')
soup = soup.find("nav", {"class": "woocommerce-pagination"})\
           .findAll("a", {"class": "page-numbers"})

pagnation_str = 'page'
pages = 1

page_list = []
for i in soup:
    if isinstance(i['class'], list):
        if 'next' in i['class'][0]:
            continue
    page_list.append(i)
pages = int(page_list[-1].getText())

cat_ret = []
for page in range(1, pages):
    r = requests.get("/".join([url, pagnation_str, str(page)]))
    raw_html = r.content
    soup = BeautifulSoup(raw_html, 'html.parser')
    soup = soup.find("div", {"class": "woocommerce-container"}).find("ul", {"class": "products clearfix products-3"}).findAll("li")
    page_ret = []
    for s in soup:
        temp_ret = {
            "name": s.find("a")['aria-label'],
            "link": s.find("a")['href']
        }
        page_ret.append(temp_ret)

    cat_ret.extend(page_ret)

print(len(cat_ret))
print(cat_ret)
# soup = soup.find("ul", {"class": "products clearfix products-3"})
# print(soup)









