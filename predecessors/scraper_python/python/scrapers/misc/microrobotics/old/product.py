from bs4 import BeautifulSoup
import requests


url = "https://www.robotics.org.za/list-all-products/6970622931072"
r = requests.get(url)
raw_html = r.content
soup = BeautifulSoup(raw_html, 'html.parser')

name = soup.find("h1", {"class": "mr-custom-product-heading"}).getText()
price = soup.find("h2").getText()

stock_table = soup.find("table", {"class": "table table-bordered table-striped"}).findAll('tr')
stock = {}
for t in stock_table:
    _t = t.findAll("th")
    if len(_t) > 0:
        continue
    _t_ = t.findAll("td")
    stock[str(_t_[0].getText())] = _t_[1].getText()

data = {
    "url": url,
    "name": name,
    "price": price,
    "stock": stock
}

print(data)