from bs4 import BeautifulSoup
import requests
from pprint import pprint

url = "https://www.diyelectronics.co.za/store/dfrobot/2134-steam-sensor-module-by-dfrobot.html"
r = requests.get(url)
raw_html = r.content
soup = BeautifulSoup(raw_html, 'html.parser')

a = soup.find('p', {'id': 'pQuantityAvailable'})
print(a)

