
from datetime import datetime
from bson.code import Code
import pymongo
conn = pymongo.MongoClient(host="localhost:27017")
db = conn.communica

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
#
# for d in data:
#     print(d)

sales = 0

def get_sales(d):
    sales = 0
    last_val = d['entries'][0]['stock']
    for data in d['entries']:
        if data['stock'] != last_val:
            sales += 1
        last_val = data['stock']

    return sales

out = {}

for g in data:
    _id = g['_id']
    sales = get_sales(g)
    rev = sales * float(g['entries'][0]['price'].replace(",", ""))
    out[rev] = [_id, sales]


import collections
od = collections.OrderedDict(sorted(out.items(), reverse=True))

print(len(od.keys()))


total = float(0)
for k, p in od.items():
    print(k, p)
    total += k

print(total)
    # print(f"group: {_id} sales: {sales} revenue: {rev}")


