import asyncio
from pyppeteer import launch
from bs4 import BeautifulSoup
import html5lib


async def main():
    browser = await launch()
    page = await browser.newPage()
    await page.goto('https://marijuanasa.co.za/product/grandaddy-bruce-feminized/')
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
    await browser.close()

    return content



asyncio.get_event_loop().run_until_complete(main())

