import requests
from bs4 import BeautifulSoup
import json

main_url = "https://www.property24.com/for-sale/cape-town/western-cape/432#P109464913"

re = requests.get(main_url)
soup = BeautifulSoup(re.text, "html.parser")
soup = soup.find('ul', {'class': 'pagination'})
soup = soup.findAll('li')
max_pages = soup[-1].find('a')['data-pagenumber']
print(max_pages)
# soup = soup.find('div', {'class': 'col-xs-9'})
# soup = soup.findAll('a', {'class': ''})
# for a in soup:
#     if a.find('span', {'class': 'js_listingTileImageHolder p24_image'}):
#         print(f"https://www.property24.com{a['href']}")
# print(soup)
