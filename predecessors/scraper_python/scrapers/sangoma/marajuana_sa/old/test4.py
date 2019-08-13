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

    soup = soup.find("ul", {"class": "page-numbers"})
    max_pages = soup.findAll("li")[-2].getText()


    await browser.close()

    return content



asyncio.get_event_loop().run_until_complete(main())

