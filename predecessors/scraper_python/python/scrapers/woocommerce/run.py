from python.scrapers.woocommerce.marajuana_sa.async_ import MarajuanaSA
from python.scrapers.woocommerce.trophy_seeds.async_ import TrophySeeds
# from python.scrapers.woocommerce.sacredseeds.async_ import SacredSeeds
# from python.scrapers.woocommerce.thehighco.async_ import TheHighCo
from motor import motor_asyncio
import asyncio

async def main(progs):
    while True:
        for prog in progs:
            await prog.main()

if __name__ == "__main__":
    addr = "localhost"
    port = 27017
    client = motor_asyncio.AsyncIOMotorClient(addr, port)
    progs = [MarajuanaSA(client),
             TrophySeeds(client),]
             # SacredSeeds(client),
             # TheHighCo(client)]

    loop = asyncio.get_event_loop()
    try:
        loop.create_task(main(progs))
        loop.run_forever()
    except KeyboardInterrupt:
        pass
    finally:
        print('step: loop.close()')
        loop.close()


