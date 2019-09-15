import asyncio
import aiohttp
import tqdm
from motor import motor_asyncio
import datetime
from bs4 import BeautifulSoup
from random import choice
from scrapers.aiohttp_exception import ignore_aiohttp_ssl_eror

desktop_agents = ['Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.99 Safari/537.36',
                 'Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.99 Safari/537.36',
                 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.99 Safari/537.36',
                 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_1) AppleWebKit/602.2.14 (KHTML, like Gecko) Version/10.0.1 Safari/602.2.14',
                 'Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.71 Safari/537.36',
                 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.98 Safari/537.36',
                 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.98 Safari/537.36',
                 'Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.71 Safari/537.36',
                 'Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.99 Safari/537.36',
                 'Mozilla/5.0 (Windows NT 10.0; WOW64; rv:50.0) Gecko/20100101 Firefox/50.0']

class MarajuanaSA(object):
    def __init__(self, client):
        self.url = 'https://marijuanasa.co.za/shop/'
        self.db = client.marajuana_sa

    @staticmethod
    def random_headers():
        return {'User-Agent': choice(desktop_agents),'Accept':'text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8'}

    async def get_product_data(self, session, url):
        try:
            async with session.get(url, headers=self.random_headers(), timeout=aiohttp.client.ClientTimeout(total=10)) as resp:
                text = await resp.read()
                soup = BeautifulSoup(text.decode('utf-8'), 'html5lib')

                soup = soup.find("div", {"class": "summary entry-summary"})

                product_name = soup.find("h1", {"class": "product_title entry-title"}).getText().strip()
                product_price = soup.find("span", {"class": "woocommerce-Price-amount amount"}).getText().replace("R", "")
                product_stock_str = soup.find("p", {"class": "stock in-stock"})
                product_stock_ = int(product_stock_str.getText()\
                                                  .replace("in", "")\
                                                  .replace("stock", "")\
                                                  .replace("(can be backordered)", "")\
                                                  .strip()) if product_stock_str is not None else None

                product_stock = int(product_stock_) if isinstance(product_stock_, int) else 0
                product_short_disc_ = soup.find("div", {"class": "woocommerce-product-details__short-description"})
                product_short_disc = product_short_disc_.getText().strip() if product_short_disc_ is not None else None
                # product_category_ = soup.find("span", {"class": "posted_in"}).find("a")
                # product_meta = soup.find("div", {"class": "product_meta"})

                # product_category_link = product_category_['href']
                # product_category = product_category_.getText()

                result = await self.db.data.insert_one({
                    "name": product_name,
                    "price": product_price,
                    "stock": product_stock,
                    "shortDisc": product_short_disc,
                    # "category": product_category,
                    # "categoryLink": product_category_link,
                    "url": url,
                    "date": datetime.datetime.utcnow()
                })
                # print(f'result {result.inserted_id}')
        except aiohttp.ClientConnectionError:
            # something went wrong with the exception, decide on what to do next
            tqdm.tqdm.write("Oops, the connection was dropped before we finished")
        except aiohttp.ClientError:
            # something went wrong in general. Not a connection error, that was handled
            # above.
            tqdm.tqdm.write("Oops, something else went wrong with the request")

    async def get_max_pages(self, session, url):
        try:
            async with session.get(url, headers=self.random_headers(), timeout=aiohttp.client.ClientTimeout(total=10)) as resp:
                text = await resp.read()

                soup = BeautifulSoup(text.decode('utf-8'), 'html5lib')
                soup = soup.find("nav", {"class": "woocommerce-pagination"})
                max_pages = soup.findAll("li")[-2].find('a').getText()
                return max_pages

        except aiohttp.ClientConnectionError:
            # something went wrong with the exception, decide on what to do next
            tqdm.tqdm.write("Oops, the connection was dropped before we finished")
        except aiohttp.ClientError:
            # something went wrong in general. Not a connection error, that was handled
            # above.
            tqdm.tqdm.write("Oops, something else went wrong with the request")


    async def get_products_on_page(self, session, url):
        try:
            async with session.get(url, headers=self.random_headers(), timeout=aiohttp.client.ClientTimeout(total=10)) as resp:
                text = await resp.read()

                soup = BeautifulSoup(text.decode('utf-8'), 'html5lib')
                soup = soup.find('main', {'id': 'main'}) \
                    .find('ul', {'class': 'products columns-4'}) \
                    .findAll('li')
                if len(soup) is 0:
                    return None
                product_links = (str(links.find('a')['href']) for links in soup)
                return product_links
        # do something with the response if needed

        # here, the async with context for the response ends, and the response is
        # released.
        except aiohttp.ClientConnectionError:
            # something went wrong with the exception, decide on what to do next
            tqdm.tqdm.write("Oops, the connection was dropped before we finished")
        except aiohttp.ClientError:
            # something went wrong in general. Not a connection error, that was handled
            # above.
            tqdm.tqdm.write("Oops, something else went wrong with the request")


    async def get_all_products(self, session, url):
        max_pages = await self.get_max_pages(session, url)
        tasks = [self.get_products_on_page(session, "".join([url, "page/", str(i)])) for i in range(1, int(max_pages) + 1)]
        responses = [await f for f in tqdm.tqdm(asyncio.as_completed(tasks), total=len(tasks), desc="{Marajuana SA (Product List)}", position=1)]
        return [i for resp in responses for i in resp]

    async def get_all_product_data(self, session, products):
        tasks = [self.get_product_data(session, prod) for prod in products]

        responses = [await f for f in tqdm.tqdm(asyncio.as_completed(tasks), total=len(tasks), desc="{Marajuana SA (Product Data)}", position=1)]
        resp = []
        for r in responses:
            resp.append(r)
        return resp


    async def main(self):
        # ignore_aiohttp_ssl_eror(asyncio.get_running_loop())
        conn = aiohttp.TCPConnector(limit=1)
        async with aiohttp.ClientSession(connector=conn) as session:
            try:
                products = await self.get_all_products(session, self.url)
                result = await self.db.product_list.insert_one({
                    "products": products,
                    "date": datetime.datetime.utcnow()
                })
                # print(f'result {result.inserted_id}')
                await self.get_all_product_data(session, products)
            except:
                pass


if __name__ == "__main__":
    addr = "localhost"
    port = 27017
    client = motor_asyncio.AsyncIOMotorClient(addr, port)

    prog = MarajuanaSA(client)

    loop = asyncio.get_event_loop()
    try:
        loop.create_task(prog.main())
        loop.run_forever()
    except KeyboardInterrupt:
        pass
    finally:
        tqdm.tqdm.write('step: loop.close()')
        loop.close()
