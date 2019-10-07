
from datetime import datetime
from bson.code import Code
import pymongo
import pprint
conn = pymongo.MongoClient(host="localhost:27017")
db = conn.pishop

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
])
out = {}
total = 0
for g in data:
    if isinstance(g['entries'][0]['price'], float):
        price = g['entries'][0]['price']
    else:
        price = float(g['entries'][0]['price'].replace(",", "").replace("R", ""))

    if isinstance(g['entries'][-1]['stock'], int):
        stock = g['entries'][-1]['stock']
    else:
        stock = int(g['entries'][-1]['stock'].replace('in', '').replace('stock', '').strip())
    name = g['entries'][-1]['name']

    if name is not None:
        if int(stock) > 0:
            a = float(price)*int(stock)
            out[name] = [a, price, stock]
            total += a

# import collections
# od = collections.OrderedDict(sorted(out.items(), reverse=False))
out = sorted(out.items(), key=lambda x: x[1], reverse=True)
for i, o in enumerate(out):
    print(i, o)
# pprint.pprint(out)
print(total)
