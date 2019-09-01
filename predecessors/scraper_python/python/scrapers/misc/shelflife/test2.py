from bs4 import BeautifulSoup
import requests
from pprint import pprint

url = "https://www.shelflife.co.za/"
r = requests.get(url)

raw_html = r.content
soup = BeautifulSoup(raw_html, 'html.parser')
soup = soup.find("body").findAll("div", {"class": "container-fluid header_bottom"})[1].find('div', {'role': 'navigation'}).findAll("li")
cats = []
for s in soup:
    if "Search" in str(s):
        continue
    cats.append("".join([url, s.find("a")["href"]]))

print(cats)
