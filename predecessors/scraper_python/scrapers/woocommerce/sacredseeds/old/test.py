from bs4 import BeautifulSoup
import requests
from pprint import pprint

url = "https://sacredseeds.co.za/"
r = requests.get(url)
raw_html = r.content
soup = BeautifulSoup(raw_html, 'html.parser')
soup = soup.find('main', {'id': 'main'}).find("ul", {"class": "products columns-3"}).findAll("li")

ret_list = []
links = []
for s in soup:
    name = s.find("h2").getText().strip()
    link = s.find("a")['href']
    ret_list.append({
        "name": name,
        "link": link
    })
    links.append(link)

print(ret_list)
print(links)