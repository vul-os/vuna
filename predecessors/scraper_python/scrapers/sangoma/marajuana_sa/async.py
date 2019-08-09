import asyncio
import aiohttp
import html5lib
import tqdm
from motor import motor_asyncio
import datetime
from bs4 import BeautifulSoup
from pyppeteer import launch

URL = 'https://marijuanasa.co.za/product/grandaddy-bruce-feminized/'


async def scrape_page(url, page, db):
    await page.goto(url)
    content = await page.content()
    soup = BeautifulSoup(content, 'html5lib')

    soup = soup.find("div", {"class": "summary entry-summary"})
    name = soup.find("h1", {"class": "product_title entry-title"}).getText()
    price = soup.find("span", {"class": "woocommerce-Price-amount amount"}).getText()
    stock = soup.find("p", {"class": "stock in-stock"}).getText()
    meta = soup.find("div", {"class": "product_meta"})
    meta_sku = meta.find("span", {"class": "sku"}).getText()
    meta_cat = meta.find("span", {"class": "posted_in"}).findAll("a")
    meta_cats = {cat['href']: cat.getText() for cat in meta_cat}
    print(name, price, stock, meta_sku, meta_cats)


async def main(url, addr="localhost", port=27017):
    while True:
        client = motor_asyncio.AsyncIOMotorClient(addr, port)
        browser = await launch()
        page = await browser.newPage()
        db = client.trophyseeds
        await scrape_page(url, page, db)

        await browser.close()

        # products = await get_all_products(url)
        # result = await db.product_list.insert_one({
        #     "products": products,
        #     "date": datetime.datetime.utcnow()
        # })
        # # print(f'result {result.inserted_id}')
        # await get_all_product_data(products, db)

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
