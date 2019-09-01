
from bs4 import BeautifulSoup
import requests
from pprint import pprint
import json
import math
import time


pages = 70

headers = {
    "User-Agent": "my web scraping program. contact me at admin@domain.com"
}

def load_page(url):
    r = requests.get(url, headers=headers)
    raw_html = r.content.decode("utf-8")

    str_list = [
        "/**/",
        "typeof",
        "BCSfFilterCallback === 'function' && BCSfFilterCallback("
    ]

    for st in str_list:
        raw_html = raw_html.replace(st, "")

    raw_html = raw_html[:-2]
    d = json.loads(raw_html)
    return d

def get_max_pages(url):
    page = load_page(url)
    num_products = int(page['total_product'])
    num_pages = math.ceil(num_products / pages)

    return num_pages

def scrape_page(page):
    for product in page:
        var = product["variants"][0]
        name = var['sku']
        price = var['price']
        avail = var['available']
        stock = var['inventory_quantity']

        print(name, price, avail, stock)


time_now = int(time.time())
url = f"https://services.mybcapps.com/bc-sf-filter/filter?t={str(1534885434)}&q=Internet+of+things&shop=communica-south-africa.myshopify.com&page=1&limit={pages}&sort=best-selling&display=grid&collection_scope=&product_available=false&variant_available=false&build_filter_tree=false&check_cache=false&sort_first=available&callback=BCSfFilterCallback&event_type=page"
print(url)
exit()

max_pg = get_max_pages(url)
for pg in range(1, max_pg+1):
    url1 = f"https://services.mybcapps.com/bc-sf-filter/filter?t={str(1534885434)}&q=Internet+of+things&shop=communica-south-africa.myshopify.com&page={pg}&limit={pages}&sort=best-selling&display=grid&collection_scope=&product_available=false&variant_available=false&build_filter_tree=false&check_cache=false&sort_first=available&callback=BCSfFilterCallback&event_type=page"
    data = load_page(url1)
    scrape_page(data['products'])
    # print(len(data['products']))





# soup = BeautifulSoup(raw_html, 'html.parser')
# soup = soup.find("div", {"class": "list-collections"})
# soup = soup.findAll("div")
#
# links = []
#
# for i in soup:
#     link = i.find("a", {"class": "hidden-product-link"})
#     if link is not None:
#         links.append(link['href'])
# #
# print(links)
