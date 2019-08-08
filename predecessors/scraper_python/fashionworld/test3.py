from bs4 import BeautifulSoup
import requests

links = ['https://www.fashionworld.co.za/just-arrived', 'https://www.fashionworld.co.za/women', 'https://www.fashionworld.co.za/kids', 'https://www.fashionworld.co.za/get-the-look', 'https://www.fashionworld.co.za/on-sale']

page_url = "?page="
page = 9999

url = "".join([links[0], "/", page_url, str(1)])
print(url)
r = requests.get(url)
raw_html = r.content
soup = BeautifulSoup(raw_html, 'html.parser')
a = soup.find("div", {"data-ajax-content": ""}).findAll("div", {"class": "columns block product"})

prod_links = []

for c in a:
    d = c.find('a')['href']
    prod_links.append(d)










