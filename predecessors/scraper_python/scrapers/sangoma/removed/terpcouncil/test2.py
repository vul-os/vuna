
from bs4 import BeautifulSoup
import requests
from pprint import pprint

url = "https://www.terpcouncil.com/categories/breeders/barneys/"
r = requests.get(url)
raw_html = r.content
soup = BeautifulSoup(raw_html, 'html.parser')
soup = soup.find("div", {"class": "shop-container"})\
           .find("div", {
                "class": "products row row-small "
                         "row-masonry has-packery "
                         "large-columns-3 "
                         "medium-columns-3 "
                         "small-columns-2"
            })

soup = soup.findAll("div", {"class": "product-small"})

for s in soup:
    print(s)
