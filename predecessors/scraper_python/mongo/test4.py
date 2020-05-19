
from datetime import datetime
from bson.code import Code
import pymongo
import pprint
conn = pymongo.MongoClient(host="localhost:27017")
db = conn.biltongandbudz

# data = db.data.find(
# # {
# #     "date":
# #     {
# #         "$gte": 0#datetime.now().timestamp() - (15 * 24 * 60 * 60 * 1000)
# #     }
# # },
# ).sort("date", pymongo.ASCENDING)

data = db.data.aggregate(
[
    {
        "$group": {
            "_id": {
                  "name": "$name",
            },
            "entries": {
              "$push": {
                "stock": "$stock",
                "price": "$price",
                "name": "$name"
              }
            }
        }
    },
    {"$sort": {"date": -1}}
], allowDiskUse=True)

out = {}
total = 0
for g in data:
    if isinstance(g['entries'][0]['price'], float):
        price = g['entries'][0]['price']
    else:
        # price = float(g['entries'][0]['price'].replace(",", "").replace("R", ""))
        price = g['entries'][0]['price']
    if isinstance(g['entries'][-1]['stock'], int):
        stock = g['entries'][-1]['stock']
    else:
        stock = int(g['entries'][-1]['stock'].replace('in', '').replace('stock', '').strip())
    name = g['entries'][-1]['name']

    if name is not None:
        if int(stock) > 0:
            if isinstance(price, str):

                a = float(price.replace(",", ""))*int(stock)
            else:
                a = float(price)*int(stock)

            out[name] = [a, price, stock]
            total += a

# import collections
# od = collections.OrderedDict(sorted(out.items(), reverse=False))
# out = sorted(out.items(), key=lambda x: x[1], reverse=True)
totol_s = 0
stock_price = 0
for i, o in out.items():
    # print(o)
    rev, price, stock = o
    stock_price += float(stock)*float(price.replace(',', ''))
    print(i, o, stock_price)
# pprint.pprint(out'ads)
print(stock_price, "asd")
