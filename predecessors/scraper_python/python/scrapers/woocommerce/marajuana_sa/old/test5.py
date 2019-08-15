import asyncio
from pyppeteer import launch
from bs4 import BeautifulSoup
import html5lib


async def main():
    browser = await launch()
    page = await browser.newPage()
    await page.goto('https://marijuanasa.co.za/product-category/marijuana-seeds/')
    content = await page.content()
    soup = BeautifulSoup(content, 'html5lib')

    soup = soup.find("main", {"id": "main"})
    soup = soup.find("ul", {"class": "products columns-4"})
    products = soup.findAll("li")
    link_list = [p.find("a")["href"] for p in products]
    print(link_list)


    await browser.close()

    return content



asyncio.get_event_loop().run_until_complete(main())

