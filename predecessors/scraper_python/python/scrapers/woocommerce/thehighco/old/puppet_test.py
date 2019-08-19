import asyncio
from pyppeteer import launch
from bs4 import BeautifulSoup
import html5lib


async def main():
    browser = await launch()
    page = await browser.newPage()
    await page.goto('https://thehighco.co.za/product/mothership-faberge-egg/')
    content = await page.content()
    soup = BeautifulSoup(content, 'html5lib')

    product_name = soup.find("h1", {"class": "product_title entry-title"}).getText().strip()
    product_price = soup.find("span", {"class": "woocommerce-Price-amount amount"}).getText().replace("R", "")
    product_stock__ = soup.find("div", {"class": "quantity buttons_added"})
    print(product_stock__.find("input", {"class": "input-text qty text"})['max'])
    await browser.close()

    return content



asyncio.get_event_loop().run_until_complete(main())

