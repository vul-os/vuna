import asyncio
import aiohttp
import html5lib
import tqdm
from motor import motor_asyncio
import datetime
from bs4 import BeautifulSoup
import json
import math
import time

from python.scrapers.aiohttp_exception import ignore_aiohttp_ssl_eror

props = {
    "sort": "best-selling",
    "display": "grid",
    "product_available": "false",
    "variant_available": "false",
    "build_filter_tree": "false",
    "check_cache": "false",
    "sort_first": "available",
    "callback": "BCSfFilterCallback&event_type=page",
}

class Communica(object):
    def __init__(self, client):
        self.url = 'https://www.trophyseeds.com/shop/'
        self.db = client.communica
        self.pages = 70


    async def get_product_data(self, page):
        for product in page['products']:
            var = product["variants"][0]
            name = var['sku']
            price = var['price']
            avail = var['available']
            stock = var['inventory_quantity']

            # print(name, price, avail, stock)


            result = await self.db.data.insert_one({
                "raw": product,
                "name": name,
                "price": price,
                "avail": avail,
                "stock": stock,
                "date": datetime.datetime.utcnow()
            })
            # tqdm.tqdm.write(f'result {result.inserted_id}')

    async def get_max_pages(self, session, url, cur_time):
        _props = {**{
            "t": str(int(cur_time)),
            "shop": "communica-south-africa.myshopify.com",
            "page": "1",
            "limit": "20"
        }, **props}
        _props_str = '&'.join([f'{key}={value}' for key, value in _props.items()])
        _url = f"{url}?{_props_str}"
        page = await self.load_page(session, _url)
        num_products = int(page['total_product'])
        num_pages = math.ceil(num_products / self.pages)
        return num_pages

    async def load_page(self, session, url):
        async with session.get(url) as resp:
            text = await resp.read()

        text = text.decode("utf-8")

        str_list = [
            "/**/",
            "typeof",
            "BCSfFilterCallback === 'function' && BCSfFilterCallback("
        ]

        for st in str_list:
            text = text.replace(st, "")
        ret = text[:-2]

        if ret is not None and len(ret) > 0:
            return json.loads(ret)
        else:
            return None

    async def load_page_num(self, session, url, page, cur_time):
        _props = {**{
            "t": str(int(cur_time)),
            "shop": "communica-south-africa.myshopify.com",
            "page": str(page),
            "limit": str(self.pages)

        }, **props}
        _props_str = '&'.join([f'{key}={value}' for key, value in _props.items()])
        _url = f"{url}?{_props_str}"
        page = await self.load_page(session, _url)
        await self.get_product_data(page)

    async def get_products(self, session, url, cur_time):
        max_pages = await self.get_max_pages(session, url, cur_time)
        tasks = [self.load_page_num(session, url, page, cur_time) for page in range(0, max_pages + 1)]
        responses = [await f for f in tqdm.tqdm(asyncio.as_completed(tasks), total=len(tasks), desc="{Communica (Product Data)}", position=1)]
        # return responses

    async def main(self):
        ignore_aiohttp_ssl_eror(asyncio.get_running_loop())
        async with aiohttp.ClientSession() as session:
            # try:
            cur_time = time.time()
            url = "https://services.mybcapps.com/bc-sf-filter/filter"
            await self.get_products(session, url, cur_time)

            # except:
            #     tqdm.tqdm.write("Caught ANY Exception")

if __name__ == "__main__":
    addr = "localhost"
    port = 27017
    client = motor_asyncio.AsyncIOMotorClient(addr, port)
    prog = Communica(client)

    loop = asyncio.get_event_loop()
    try:
        loop.create_task(prog.main())
        loop.run_forever()
    except KeyboardInterrupt:
        pass
    finally:
        tqdm.tqdm.write('step: loop.close()')
        loop.close()
