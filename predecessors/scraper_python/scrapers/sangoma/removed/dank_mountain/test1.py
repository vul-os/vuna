import asyncio
from pyppeteer import launch
from bs4 import BeautifulSoup
import html5lib


async def main():
    browser = await launch()
    page = await browser.newPage()
    await page.goto('https://dankmountain.co.za/?page_id=240')
    content = await page.content()
    soup = BeautifulSoup(content, 'html5lib')

    soup = soup.find("div", {"id": "secondary"}).find("ul", {"class": "product-categories"})
    ret_list = []
    links = []
    for s in soup:
        z = s.find("a")
        if not isinstance(z, int):
            ret_list.append({
                "name": z.getText(),
                "link": z['href']
            })
            links.append(z['href'])
    print(ret_list)
    await browser.close()

    return content

if __name__ == "__main__":
    asyncio.get_event_loop().run_until_complete(main())





