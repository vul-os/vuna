import asyncio
import aiohttp
import html5lib
import tqdm
from motor import motor_asyncio
import datetime
from bs4 import BeautifulSoup

SELECTED_URL = 'https://www.trophyseeds.com/product/ethos-genetics-zsweet-inzanity/'
URL = 'https://www.trophyseeds.com/shop/'
SELECTED_URL3 = 'https://www.trophyseeds.com/shop/page/1'


# class TrophySeeds(object):
async def get_product_data(url, db):
    async with aiohttp.ClientSession() as session:
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

        result = await db.data.insert_one({
            "name": product_name,
            "price": product_price,
            "stock": product_stock,
            "shortDisc": product_short_disc,
            "category": product_category,
            "categoryLink": product_category_link,
            "url": url,
            "date": datetime.datetime.utcnow()
        })
        # print(f'result {result.inserted_id}')


async def get_max_pages(url):
    async with aiohttp.ClientSession() as session:
        async with session.get(url) as resp:
            text = await resp.read()

        soup = BeautifulSoup(text.decode('utf-8'), 'html5lib')
        soup = soup.find("nav", {"class": "woocommerce-pagination"})
        max_pages = soup.findAll("li")[-2].find('a').getText()
        return max_pages


async def get_products_on_page(url):
    async with aiohttp.ClientSession() as session:
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


async def get_all_products(url):
    max_pages = await get_max_pages(url)
    tasks = [get_products_on_page("".join([url, "page/", str(i)])) for i in range(1, int(max_pages))]
    responses = [await f for f in tqdm.tqdm(asyncio.as_completed(tasks), total=len(tasks))]
    return [i for resp in responses for i in resp]


async def get_all_product_data(products, db):
    tasks = [get_product_data(prod, db) for prod in products]
    responses = [await f for f in tqdm.tqdm(asyncio.as_completed(tasks), total=len(tasks))]
    return [resp for resp in responses]


async def main(url, addr="localhost", port=27017):
    while True:
        client = motor_asyncio.AsyncIOMotorClient(addr, port)
        db = client.trophyseeds

        products = await get_all_products(url)
        result = await db.product_list.insert_one({
            "products": products,
            "date": datetime.datetime.utcnow()
        })
        # print(f'result {result.inserted_id}')
        await get_all_product_data(products, db)

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
