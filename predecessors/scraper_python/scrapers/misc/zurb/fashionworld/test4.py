from bs4 import BeautifulSoup
import requests

url = "https://www.fashionworld.co.za/products/lds-ethnic-weave-platform-block-heel-p-toe"
r = requests.get(url)
raw_html = r.content
soup = BeautifulSoup(raw_html, 'html.parser')
a = soup.find("div", {"data-ajax-content": ""}).findAll("div", {"class": "columns block product"})










