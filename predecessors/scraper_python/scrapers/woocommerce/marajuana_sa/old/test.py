
from bs4 import BeautifulSoup
import requests
from pprint import pprint

url = "https://marijuanasa.co.za/"
headers = {
    "User-Agent": "my web scraping program. contact me at admin@domain.com"
}
r = requests.get(url, headers=headers)
# r = requests.get(url)
raw_html = r.content
soup = BeautifulSoup(raw_html, 'html.parser')
print(soup)
soup = soup.find("ul", {"class": "product-categories"})
print(soup)
soup = soup.findAll("li")

print(soup)
