import asyncio
from pyppeteer import launch
from bs4 import BeautifulSoup
import html5lib


async def main():
    browser = await launch()
    page = await browser.newPage()
    await page.goto('https://dankmountain.co.za/?product_cat=angelo-chillums')
    content = await page.content()
    soup = BeautifulSoup(content, 'html5lib')

    soup = soup.find("main", {"id": "main"}).find("ul", {"class": "products columns-3"}).findAll("li")

    ret_list = []
    links = []
    for s in soup:
        z = s.find("a", {"class": "woocommerce-LoopProduct-link"})
        if not isinstance(z, int):
            ret_list.append({
                "name": z.find("h2", {"class": "woocommerce-loop-product__title"}).getText(),
                "link": z['href']
            })
            links.append(z['href'])
    print(ret_list)



    await browser.close()

    return content

if __name__ == "__main__":
    asyncio.get_event_loop().run_until_complete(main())





