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
        if sub.ul is None:
            continue
        data = await get_info(sub)
        if sub.ul is not None:
            # recurse down
            data['children'] = await parse_sub_list(sub.ul)
        result[sub.a.get_text(strip=True)] = data
    return result


async def list_helper(content):
    soup = BeautifulSoup(content, 'html5lib')
    soup = soup.find("ul", {"class": "product-categories"})
    return await parse_list(soup)

# @asyncio.coroutine
# def parse_ul(elem):
#     result = {}
#     for sub in elem.find_all('li', recursive=False):
#         if sub.ul is None:
#             continue
#         data = {k: v for k, v in get_info(sub)}
#         if sub.ul is not None:
#             # recurse down
#             data['children'] = parse_ul(sub.ul)
#         result[sub.a.get_text(strip=True)] = data
#     return result

async def main(loop):
    browser = await launch()
    page = await browser.newPage()
    await page.goto('https://marijuanasa.co.za/shop/')
    content = await page.content()
    await browser.close()

    a = await list_helper(content)
    print(a)
    return content




loop = asyncio.get_event_loop()
a = loop.run_until_complete(main(loop))
