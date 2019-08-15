import asyncio
from pyppeteer import launch
from bs4 import BeautifulSoup
import html5lib


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


async def main(loop):
    browser = await launch()
    page = await browser.newPage()
    await page.goto('https://marijuanasa.co.za/shop/')
    content = await page.content()
    await browser.close()

    a = await list_helper(content)

    links = []
    for k, v in a.items():
        if isinstance(v, dict):
            if 'children' in v.keys():
                for i, j in v.items():

                    _, __link = zip(*j.items())
                    links.append(__link)
            else:
                links.append(v)
    print(links)
    return content




loop = asyncio.get_event_loop()
a = loop.run_until_complete(main(loop))
