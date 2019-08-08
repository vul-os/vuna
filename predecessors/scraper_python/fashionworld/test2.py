from bs4 import BeautifulSoup
import requests

links = ['https://www.fashionworld.co.za/just-arrived', 'https://www.fashionworld.co.za/women', 'https://www.fashionworld.co.za/kids', 'https://www.fashionworld.co.za/get-the-look', 'https://www.fashionworld.co.za/on-sale']

page_url = "?page="
page = 9999

url = "".join([links[0], "/", page_url, str(page)])
print(url)
r = requests.get(url)
raw_html = r.content
soup = BeautifulSoup(raw_html, 'html.parser')

a = soup.find("ul", {"class": "pagination float-right"})
b = a.find("li", {"class": "current"}).getText().replace("You're on page", "").strip()
max_page = int(b)

print(f"Max Pages: {max_page}")


for i in range(1, max_page):
    url = "".join([links[0], "/", page_url, str(i)])
    print(url)
    r = requests.get(url)
    raw_html = r.content
    soup = BeautifulSoup(raw_html, 'html.parser')

    z = soup.find("div", {"data-ajax-content": ""})
    print(z)

    exit()






