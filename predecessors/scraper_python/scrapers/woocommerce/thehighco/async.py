import asyncio
import aiohttp
import html5lib
import tqdm
from motor import motor_asyncio
import datetime
from bs4 import BeautifulSoup
from random import choice
import types
import certifi
import ssl

URL = 'https://thehighco.co.za/shop/'

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

def random_headers():
    return {'User-Agent': choice(desktop_agents),'Accept':'text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8'}

async def get_product_data(session, url, db):
    try:

        async with session.get(url, headers=random_headers()) as resp:
            text = await resp.read()
            soup = BeautifulSoup(text.decode('utf-8'), 'html5lib')

            soup = soup.find("div", {"class": "summary entry-summary"})

            if soup is None:
                return None

            product_name = soup.find("h1", {"class": "product_title entry-title"}).getText().strip()
            product_price = soup.find("span", {"class": "woocommerce-Price-amount amount"}).getText().replace("R", "")
            product_stock__ = soup.find("div", {"class": "quantity buttons_added"})
            product_stock_ = int(product_stock__.find("input", {"class": "input-text qty text"})['max']) \
                if product_stock__ is not None else 0
            product_stock = int(product_stock_) if isinstance(product_stock_, int) else 0
            product_short_disc_ = soup.find("div", {"class": "woocommerce-product-details__short-description"})
            product_short_disc = product_short_disc_.getText().strip() if product_short_disc_ is not None else None
            # product_category_ = soup.find("span", {"class": "posted_in"}).find("a")
            # product_meta = soup.find("div", {"class": "product_meta"})

            # product_category_link = product_category_['href']
            # product_category = product_category_.getText()

            result = await db.data.insert_one({
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
        print("Oops, the connection was dropped before we finished")
    except aiohttp.ClientError:
        # something went wrong in general. Not a connection error, that was handled
        # above.
        print("Oops, something else went wrong with the request")

    except ssl.SSLError as e:
        print('ssl Error handled')

async def get_max_pages(session, url):
    return 1
    # try:
    #
    #     async with session.get(url, headers=random_headers()) as resp:
    #         text = await resp.read()
    #
    #         soup = BeautifulSoup(text.decode('utf-8'), 'html5lib')
    #         soup = soup.find("nav", {"class": "woocommerce-pagination"})
    #         if soup is None:
    #             return None
    #         return soup.findAll("li")[-2].find('a').getText()
    #
    # except aiohttp.ClientConnectionError:
    #     # something went wrong with the exception, decide on what to do next
    #     print("Oops, the connection was dropped before we finished")
    # except aiohttp.ClientError:
    #     # something went wrong in general. Not a connection error, that was handled
    #     # above.
    #     print("Oops, something else went wrong with the request")


async def get_items_on_page(session, url):
    try:
        async with session.get(url, headers=random_headers()) as resp:
            text = await resp.read()

            soup = BeautifulSoup(text.decode('utf-8'), 'html5lib')
            soup = soup.find('main', {'id': 'main'}) \
                .find('ul', {'class': 'products clearfix products-3'}) \
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
        print("Oops, the connection was dropped before we finished")
        return False
    except aiohttp.ClientError:
        # something went wrong in general. Not a connection error, that was handled
        # above.
        print("Oops, something else went wrong with the request")
        return False
    except ssl.SSLError as e:
        print('ssl Error handled')

async def get_all_products_per_category(session, url):
    max_pages_ = await get_max_pages(session, url)
    max_pages = 1 if max_pages_ is None else max_pages_
    linkz = ["".join([url, "page/", str(i)]) for i in range(1, int(max_pages)+1)]
    tasks = [get_items_on_page(session, i) for i in linkz]
    responses = [await f for f in asyncio.as_completed(tasks)]
    if not len(responses) > 0:
        return None
    _ret = []
    for i in responses:
        if isinstance(i, types.GeneratorType) or isinstance(i, list):
            _ret.extend(i)
        else:
            _ret.append(i)
    # print(f"cat: {url}, pages: {max_pages}, num_items: {len(_ret)}")
    return _ret

async def get_all_product_data(session, products, db):
    tasks = [get_product_data(session, prod, db) for prod in products]
    responses = [await f for f in tqdm.tqdm(asyncio.as_completed(tasks), total=len(tasks))]
    return [resp for resp in responses]

async def get_all_products(session, categories):
    tasks = [get_all_products_per_category(session, cat) for cat in categories]
    responses = [await f for f in tqdm.tqdm(asyncio.as_completed(tasks), total=len(tasks))]
    return responses

async def main(url, addr="localhost", port=27017):
    client = motor_asyncio.AsyncIOMotorClient(addr, port)
    db = client.the_high_co
    conn = aiohttp.TCPConnector(limit=3)
    async with aiohttp.ClientSession(connector=conn) as session:
        while True:
            categories = await get_items_on_page(session, url)
            products = await get_all_products(session, categories)
            result = await db.product_list.insert_one({
                "products": products,
                "date": datetime.datetime.utcnow()
            })
            flat_list = []
            for sublist in products:
                for item in sublist:
                    flat_list.append(item)

            # # print(f'result {result.inserted_id}')
            await get_all_product_data(session, flat_list, db)

if __name__ == "__main__":
    loop = asyncio.get_event_loop()

    try:
        loop.create_task(main(URL))
        loop.run_forever()
    except KeyboardInterrupt:
        pass
    finally:
        print('step: loop.close()')
        loop.close()




