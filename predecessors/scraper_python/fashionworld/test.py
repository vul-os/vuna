from bs4 import BeautifulSoup
import requests


url = "https://www.fashionworld.co.za"
r = requests.get(url)
raw_html = r.content
soup = BeautifulSoup(raw_html, 'html.parser')

a = soup.find("div", {"class": "menu-section"})\
    .find("div", {"class": "row"})\
    .find("nav", {"class": "main-menu"})

b = a.find_all("div", {"class": "is-child"})
for z in b:
    z.decompose()

c = a.findAll('li')
b = a.findAll("a", {"class": "disable-link"})

for j in b:
    j.decompose()

links = []
for g in c:
    f = g.find("a")
    if f is not None:
        ret = "".join([url, f['href']])
        links.append(ret)
print(links)



    # .find("ul", {"class": "menu"})\
    # .findAll('li', recursive=False)






# print(a)
