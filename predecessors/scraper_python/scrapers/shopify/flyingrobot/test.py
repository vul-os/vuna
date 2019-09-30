


from bs4 import BeautifulSoup
import requests
from pprint import pprint

url = "https://flyingrobot.co/collections/speed-controllers/products/dys-elf-4in1-10a-esc"


r = requests.get(url)
# r = requests.get(url)
raw_html = r.content
soup = BeautifulSoup(raw_html, 'html.parser')
soup = soup.find("span", {"class": "product-single__stock"})
print(soup)
