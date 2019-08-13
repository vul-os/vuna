from bs4 import BeautifulSoup
import requests
from pprint import pprint

url = "https://www.terpcouncil.com/shop/"
r = requests.get(url)
raw_html = r.content
soup = BeautifulSoup(raw_html, 'html.parser')
soup = soup.find("div", {"id": "shop-sidebar"}).find("ul", {"class": "product-categories"})

def get_data(soup):
    cat = soup.find('a')
    cat_name = cat.getText()
    link = cat['href']
    return {cat_name: link}


def parse_ul(elem):
    result = {}
    for sub in elem.find_all('li', recursive=False):
        if sub.ul is None:
            continue
        data = get_data(sub)
        if sub.ul is not None:
            # recurse down
            data['children'] = parse_ul(sub.ul)
        result[sub.a.get_text(strip=True)] = data
    return result

print(parse_ul(soup))