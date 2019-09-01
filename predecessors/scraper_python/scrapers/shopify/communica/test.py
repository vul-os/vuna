
from bs4 import BeautifulSoup
import requests
from pprint import pprint

url = "https://www.communica.co.za/shop"
base_url = "https://www.communica.co.za/"
headers = {
    "User-Agent": "my web scraping program. contact me at admin@domain.com"
}
r = requests.get(url, headers=headers)
# r = requests.get(url)
raw_html = r.content
soup = BeautifulSoup(raw_html, 'html.parser')
soup = soup.find("div", {"class": "list-collections"})
soup = soup.findAll("div")

links = []

for i in soup:
    link = i.find("a", {"class": "hidden-product-link"})
    if link is not None:
        links.append(link['href'])
#
print(links)
