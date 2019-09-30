




from bs4 import BeautifulSoup
import requests
from pprint import pprint

url = "https://flyingrobot.co//collections/latest?page=33"


r = requests.get(url)
# r = requests.get(url)
raw_html = r.content
soup = BeautifulSoup(raw_html, 'html.parser')

soup = soup.findAll('a', {'class': 'product-card'})
links = []
for s in soup:
    print(s['href'])
print(links)
