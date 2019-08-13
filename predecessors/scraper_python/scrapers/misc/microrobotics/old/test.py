from bs4 import BeautifulSoup
import requests


# url = "https://www.robotics.org.za"
# r = requests.get(url)
# raw_html = r.content
# soup = BeautifulSoup(raw_html, 'html.parser')
# a = soup.find("div", {"id": "yumenu-1"})
# b = a.find('div', {'class': 'yum-am'})
# c = b.find('ul')
#
# children = c.find_all("li")
# link_list = []
# for c in children:
#     d = c.find('a')
#     link_list.append(d['href'])
#
# print(link_list)

# or just https://www.robotics.org.za/list-all-products'

url = "https://www.robotics.org.za/list-all-products"
r = requests.get(url)

pagination_str = "?page="
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

    z = soup.find("div", {"class": "row product-layout-row"})
    b = soup.select("div[class^=product-layout]")
    if len(b) is 0:
        break

    for c in b:
        d = c.find("div", {"class": "image"}).find('a')['href']
        product_links.append(d)

    print(f"Completed Page: {page}")
    page += 1

with open("./text.txt", "w") as f:
    for p in product_links:
        f.write(f"{p} \n")

