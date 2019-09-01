from bs4 import BeautifulSoup
import requests
from pprint import pprint

url = "https://www.cybercellar.com/spirits/gin"
r = requests.get(url)

pagination_str = "?p="
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

    a = soup.find('div', {'class': 'col-main'})
    b = a.find('div', {'class': 'category-products'})

    if b is None:
        break

    b = b.find("ul")
    c = b.findAll('li')

    if len(b) is 0:
        break

    for d in c:
        product_info = d.find("div", {"class": "product-info"})
        if product_info is not None:
            product_name = product_info.find("h2", {"class": "product-name"}).getText().strip()
            brand_name = product_info.find("span", {"class": "brand-name"}).getText().strip()
            vintage = product_info.find("span", {"class": "vintage"}).getText().strip()

            box_price = product_info.find("div", {"class": "price-box"})
            case_price = product_info.find("div", {"class": "case-price"})
            qty = product_info.find("div", {"class": "qty-available-alt"})

            if box_price is not None:
                if box_price.find("span") is not None:
                    box_price = box_price.find("span").getText().strip()
            if case_price is not None:
                if case_price.find("span") is not None:
                    case_price = case_price.find("span").getText().strip()
                else:
                    case_price = case_price.getText().strip()
            if qty is not None:
                qty = qty.getText().strip()

            link = d.find("a", {"class": "product-image"})["href"]

            _ret = {
                "brand_name": brand_name,
                "vintage": vintage,
                "box_price": box_price,
                "case_price": case_price,
                "stock": qty,
                "link": link
            }

            product_links.append(_ret)
            print(_ret)

    print(f"Completed Page: {page}")
    page += 1

with open("./text.txt", "w") as f:
    for p in product_links:
        f.write(f"{p} \n")


