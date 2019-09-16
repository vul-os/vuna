
from datetime import datetime
from bson.code import Code
import pymongo
import pprint
conn = pymongo.MongoClient(host="localhost:27017")
db = conn.threedprintingstore

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
              }
            }
        }
    },
    {"$sort": {"date": -1}}
])
out = {}
total = 0
for g in data:
    _id = g['_id']['name']
    price = g['entries'][0]['price']#.replace(",", "").replace("R", "")
    stock = g['entries'][-1]['stock']
    if _id is not None:
        a = float(price)*int(stock)
        out[_id] = [a, price, stock]
        total += a

import collections
# od = collections.OrderedDict(sorted(out.items(), reverse=False))
out = sorted(out.items(), key=lambda x: x[1], reverse=True)
pprint.pprint(out)
print(total)
