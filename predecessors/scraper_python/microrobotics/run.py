from bs4 import BeautifulSoup
import requests
from multiprocessing import Pool
from pymongo import MongoClient
import datetime

class MicroRobotics(object):
    def __init__(self, base_url, mongo_url="localhost"):
        self.url = base_url
        self.product_links = []

    def __call__(self, *args, **kwargs):
        self.product_links = self.get_all_products()
        return self.product_links

    def scrape(self, url):
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
            "stock": stock,
            "date": datetime.datetime.utcnow()
        }
        client = MongoClient("localhost")
        db = client.microrobotics

        db.data.insert_one(data)
        return data

    def get_list_categories(self):
        r = requests.get(self.url)
        raw_html = r.content
        soup = BeautifulSoup(raw_html, 'html.parser')
        a = soup.find("div", {"id": "yumenu-1"})
        b = a.find('div', {'class': 'yum-am'})
        c = b.find('ul')

        children = c.find_all("li")
        link_list = []
        for c in children:
            d = c.find('a')
            link_list.append(d['href'])

        return link_list

    @staticmethod
    def get_products_on_page(url):
        print(f"Url for Page: {url}")

        r = requests.get(url)
        if not r.status_code < 400:
            return False

        raw_html = r.content
        soup = BeautifulSoup(raw_html, 'html.parser')

        # a = soup.find("div", {"class": "row product-layout-row"})
        b = soup.select("div[class^=product-layout]")
        if len(b) is 0:
            return False
        listy = []
        for c in b:
            listy.append(
                c.find("div", {"class": "image"}).find('a')['href']
            )
        return listy

    def get_all_products(self, url="https://www.robotics.org.za/list-all-products"):
        pagination_str = "?page="
        page = 1
        product_links = []
        while True:
            url_ = "".join([url, pagination_str, str(page)])
            d = self.get_products_on_page(url_)
            if d is False or not len(d) > 0:
                break

            product_links.extend(d)
            print(f"Completed Page: {page}")
            page += 1

        return product_links

    @staticmethod
    def save_product_links(file, links):
        with open(file, "w") as f:
            for p in links:
                f.write(f"{p} \n")

if __name__ == "__main__":
    p = Pool(4)
    mr = MicroRobotics("https://www.robotics.org.za")
    links = mr()
    p.map(mr.scrape, links)
    p.terminate()
    p.join()
