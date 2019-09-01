import asyncio
import aiohttp
import tqdm
from motor import motor_asyncio
import datetime
from bs4 import BeautifulSoup

from scrapers.aiohttp_exception import ignore_aiohttp_ssl_eror

class TrophySeeds(object):
    def __init__(self, client):
        self.url = 'https://www.trophyseeds.com/shop/'
        self.db = client.trophyseeds

    async def get_product_data(self, session, url):
        async with session.get(url) as resp:
            text = await resp.read()

        soup = BeautifulSoup(text.decode('utf-8'), 'html5lib')
        soup = soup.find("div", {"class": "summary entry-summary"})

        product_name = soup.find("h1", {"class": "product_title entry-title"}).getText()
        product_price = soup.find("span", {"class": "woocommerce-Price-amount amount"}).getText().replace("R", "")
        product_stock_str = soup.find("p", {"class": "stock in-stock"}).getText()\
                                                                       .replace("in", "")\
                                                                       .replace("stock", "")\
                                                                       .strip()
        product_stock = int(product_stock_str) if int(product_stock_str.isdigit()) else 0
        product_short_disc = soup.find("div", {"class": "woocommerce-product-details__short-description"}).getText()
        product_category_ = soup.find("span", {"class": "posted_in"}).find("a")
        product_category_link = product_category_['href']
        product_category = product_category_.getText()

        result = await self.db.data.insert_one({
            "name": product_name,
            "price": product_price,
            "stock": product_stock,
            "shortDisc": product_short_disc,
            "category": product_category,
            "categoryLink": product_category_link,
            "url": url,
            "date": datetime.datetime.utcnow()
        })
        # tqdm.tqdm.write(f'result {result.inserted_id}')


    async def get_products_on_page(self, session, url):
        async with session.get(url) as resp:
            text = await resp.read()

        soup = BeautifulSoup(text.decode('utf-8'), 'html5lib')
        soup = soup.find('main', {'id': 'main'})\
                   .find('ul', {'class': 'products columns-3'})\
                   .findAll('li')
        if len(soup) is 0:
            return None
        product_links = (str(links.find('a')['href']) for links in soup)
        return product_links

    async def get_all_products(self, session, url):
        max_pages = await self.get_max_pages(session, url)
        tasks = [self.get_products_on_page(session, "".join([url, "page/", str(i)])) for i in range(1, int(max_pages + 1))]
        responses = [await f for f in tqdm.tqdm(asyncio.as_completed(tasks), total=len(tasks), desc="{Trophy Seeds (Product List)}", position=1)]
        return [i for resp in responses for i in resp]

    async def get_all_product_data(self, session, products):
        tasks = [self.get_product_data(session, prod) for prod in products]
        responses = [await f for f in tqdm.tqdm(asyncio.as_completed(tasks), total=len(tasks), desc="{Trophy Seeds (Product Data)}", position=1)]
        return [resp for resp in responses]

    async def get_all_categories(self, session, url):
        async with session.get(url) as resp:
            text = await resp.read()

        soup = BeautifulSoup(text.decode('utf-8'), 'html5lib')
        soup = soup.find("body").findAll("div", {"class": "container-fluid header_bottom"})[1].find('div', {
            'role': 'navigation'
        }).findAll("li")
        cats = []
        for s in soup:
            if "Search" in str(s):
                continue
            cats.append("".join([url, s.find("a")["href"]]))
        return cats

    async def main(self):
        ignore_aiohttp_ssl_eror(asyncio.get_running_loop())
        async with aiohttp.ClientSession() as session:
            categories = await self.get_all_categories(session, self.url)



if __name__ == "__main__":
    addr = "localhost"
    port = 27017
    client = motor_asyncio.AsyncIOMotorClient(addr, port)
    prog = TrophySeeds(client)

    loop = asyncio.get_event_loop()
    try:
        loop.create_task(prog.main())
        loop.run_forever()
    except KeyboardInterrupt:
        pass
    finally:
        tqdm.tqdm.write('step: loop.close()')
        loop.close()
