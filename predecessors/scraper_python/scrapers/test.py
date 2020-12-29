import requests
from bs4 import BeautifulSoup

main_url = "http://www.3dprintingstore.co.za/3d-printers-upgrades/creality-cr-6-se-carborundum-glass/"

# Getting individual cities url
re = requests.get(main_url)
soup = BeautifulSoup(re.text, "html.parser")
breadcrumbs = soup.find('div', {'id': 'ProductBreadcrumb'})
cats = {}
for cat in breadcrumbs.findAll('ul'):
    crumbs = [cat.findAll('li')]
    if len(crumbs) == 0:
        continue
    cats[crumbs[0][1].find('a').text] = {
        'url': crumbs[0][1].find('a')['href'],
        'subcat': {
            k.find('a').text: k.find('a')['href'] for i, k in enumerate(crumbs[0][2:-1])
        } if len(crumbs[0]) > 2 else None
    }

for k, v in cats.items():
    print(k, v)
    print()