import asyncio
import aiohttp
import html5lib
import tqdm
from motor import motor_asyncio
import datetime
from bs4 import BeautifulSoup
from random import choice

URL = 'https://marijuanasa.co.za/shop/'

async def jon(session, url):
    async with session.get(url) as resp:
        await resp.read()

def batch(iterable, n=1):
    l = len(iterable)
    for ndx in range(0, l, n):
        yield iterable[ndx:min(ndx + n, l)]

async def main(url, addr="localhost", port=27017):
    while True:
        async with aiohttp.ClientSession() as session:
            tasks = [jon(session, url) for i in range(0, 1000)]

            for a in batch(tasks, 5):
                responses = [await f for f in tqdm.tqdm(asyncio.as_completed(a), total=len(a))]


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

