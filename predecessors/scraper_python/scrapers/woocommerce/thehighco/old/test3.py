from bs4 import BeautifulSoup
import requests
from pprint import pprint

url = "https://thehighco.co.za/product/mothership-faberge-egg/"
r = requests.get(url)
raw_html = r.content
soup = BeautifulSoup(raw_html, 'html.parser')
soup = soup.find("input", {"class": "input-text qty text"})
print(soup)










