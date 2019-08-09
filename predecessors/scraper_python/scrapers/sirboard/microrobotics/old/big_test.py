from bs4 import BeautifulSoup
from multiprocessing import Pool
import requests
from tqdm import tqdm

import pickle

with open("/home/imran/Documents/personal/projects/scraper/text.txt") as f:
    content = f.readlines()
content = [x.strip() for x in content]



def scrape(url):
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
    return data

p = Pool(4)
p.map(scrape, content)
p.terminate()
p.join()

# for url in tqdm(content):
#     r = requests.get(url)
#     raw_html = r.content
#     soup = BeautifulSoup(raw_html, 'html.parser')
#
#     name = soup.find("h1", {"class": "mr-custom-product-heading"}).getText()
#     price = soup.find("h2").getText()
#
#     stock_table = soup.find("table", {"class": "table table-bordered table-striped"}).findAll('tr')
#     stock = {}
#     for t in stock_table:
#         _t = t.findAll("th")
#         if len(_t) > 0:
#             continue
#         _t_ = t.findAll("td")
#         stock[str(_t_[0].getText())] = _t_[1].getText()
#
#     data = {
#         "url": url,
#         "name": name,
#         "price": price,
#         "stock": stock
#     }
#
#     print(data)

    # # Now we "sync" our database
    # with open("./out.pkl", 'wb') as wfp:
    #     pickle.dump(scores, wfp)