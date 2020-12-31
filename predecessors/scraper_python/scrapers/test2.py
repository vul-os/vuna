import requests
from bs4 import BeautifulSoup
import json
import re

main_url = "https://www.biltongandbudz.co.za/shop/page/7"

req = requests.get(main_url)
soup = BeautifulSoup(req.text, "html.parser")

# soup = soup.find('ul', {'class': re.compile(r'products')})
# print(soup)
# soup = soup.find_all('li', {'class': re.compile(r'product')})
# product_links = [str(links.find('a')['href']) for links in soup]


soup = soup.select_one('.products')
soup = soup.select('.product')
product_links = [str(links.find('a')['href']) for links in soup]

# if len(soup) == 0:
#     return None

print(product_links)# print(soup)
# soup = soup.find('ul', {'class': 'products'}, partial=True) \
#     .find('li', {'class': 'product'}, partial=True) \
#     .findAll('a')
# if len(soup) == 0:
#     return None
# product_links = [s['href'] for s in soup]
#
# my_dictionary = "mat"