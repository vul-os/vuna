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


async def scrape_categories(url, page, db):
    await page.goto(url)
    content = await page.content()
    return await list_helper(content)


async def get_max_pages(page, url):
    await page.goto(url)
    content = await page.content()
    soup = BeautifulSoup(content, 'html5lib')
    return content.find("ul", {"class": "page-numbers"}).findAll("li")[-2].getText()


async def get_info(node):
    cat = node.find("a")
    link = cat['href']
    name = cat.getText()
    return {name: link}


async def parse_sub_list(content):
    result = {}
    for sub in content.find_all('li', recursive=False):
        result[sub.a.get_text(strip=True)] = await get_info(sub)
    return result


async def parse_list(content):
    result = {}
    for sub in content.find_all('li', recursive=False):
        data = await get_info(sub)
        if sub.ul is not None:
            data['children'] = await parse_sub_list(sub.ul)
        result[sub.a.get_text(strip=True)] = data
    return result


async def list_helper(content):
    soup = BeautifulSoup(content, 'html5lib')
    soup = soup.find("ul", {"class": "product-categories"})
    return await parse_list(soup)


async def main(url, addr="localhost", port=27017):
    while True:
        client = motor_asyncio.AsyncIOMotorClient(addr, port)
        browser = await launch()
        page = await browser.newPage()
        db = client.trophyseeds


        await scrape_categories(url, page, db)


        max_pages = await get_max_pages(page, url)
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
