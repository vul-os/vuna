import asyncio
import aiohttp
import html5lib

import tqdm
from motor import motor_asyncio
import datetime
from bs4 import BeautifulSoup

from python.scrapers.aiohttp_exception import ignore_aiohttp_ssl_eror

class BongAlong(object):
    def __init__(self, client):
        self.url = 'https://bongalong.co.za/shop/'
        self.db = client.trophyseeds

    async def get_product_data(self, session, url):
        async with session.get(url) as resp:
            text = await resp.read()

        soup = BeautifulSoup(text.decode('utf-8'), 'html5lib')
        soup = soup.find("div", {"class": "summary entry-summary"})

        product_name = soup.find("h1", {"class": "product_title entry-title"}).getText()
        product_price = soup.find("span", {"class": "woocommerce-Price-amount amount"}).getText().replace("R", "")
        product_stock_str = soup.find("p", {"class": "stock in-stock"}) #.getText()
        if product_stock_str is not None:
            product_stock_str = product_stock_str.getText()

        result = await self.db.data.insert_one({
            "name": product_name,
            "price": product_price,
            "stock": product_stock_str,
            "url": url,
            "date": datetime.datetime.utcnow()
        })

    async def get_max_pages(self, session, url):
        async with session.get(url) as resp:
            text = await resp.read()

        soup = BeautifulSoup(text.decode('utf-8'), 'html5lib')
        soup = soup.find("nav", {"class": "woocommerce-pagination"})
        max_pages = soup.findAll("li")[-2].find('a').getText()
        return max_pages

    async def get_products_on_page(self, session, url):
        async with session.get(url) as resp:
            text = await resp.read()

        soup = BeautifulSoup(text.decode('utf-8'), 'html5lib')
        soup = soup.find('ul', {'class': 'products columns-3'})\
                   .findAll('li')
        if len(soup) is 0:
            return None
        product_links = (str(links.find('a')['href']) for links in soup)
        return product_links

    async def get_all_products(self, session, url):
        max_pages = await self.get_max_pages(session, url)
        tasks = [self.get_products_on_page(session, "".join([url, "page/", str(i)])) for i in range(1, int(max_pages))]
        responses = [await f for f in tqdm.tqdm(asyncio.as_completed(tasks), total=len(tasks), desc="{BongAlong (Product List)}", position=1)]
        return [i for resp in responses for i in resp]

    async def get_all_product_data(self, session, products):
        tasks = [self.get_product_data(session, prod) for prod in products]
        responses = [await f for f in tqdm.tqdm(asyncio.as_completed(tasks), total=len(tasks), desc="{BongAlong (Product Data)}", position=1)]
        return [resp for resp in responses]

    async def main(self):
        ignore_aiohttp_ssl_eror(asyncio.get_running_loop())
        async with aiohttp.ClientSession() as session:
            try:
                products = await self.get_all_products(session, self.url)
                result = await self.db.product_list.insert_one({
                    "products": products,
                    "date": datetime.datetime.utcnow()
                })
                # tqdm.tqdm.write(f'result {result.inserted_id}')
                for p in tqdm.tqdm(products, total=len(products), desc="{BongAlong (Product Data)}", position=1):
                    await self.get_product_data(session, p)
            except:
                tqdm.tqdm.write("Caught ANY Exception")

if __name__ == "__main__":
    addr = "localhost"
    port = 27017
    client = motor_asyncio.AsyncIOMotorClient(addr, port)
    prog = BongAlong(client)

    loop = asyncio.get_event_loop()
    try:
        loop.create_task(prog.main())
        loop.run_forever()
    except KeyboardInterrupt:
        pass
    finally:
        tqdm.tqdm.write('step: loop.close()')
        loop.close()
