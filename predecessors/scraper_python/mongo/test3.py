
from datetime import datetime
from bson.code import Code
import pymongo
conn = pymongo.MongoClient(host="localhost:27017")
db = conn.netram

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
                "url": "$url"
              }
            }
        }
    },
    {"$sort": {"date": -1}}
], allowDiskUse=True)
#
# for d in data:
#     print(d)

sales = 0

def get_sales(d):
    sales = 0
    last_val = d['entries'][0]['stock']
    if not isinstance(d['entries'][0]['stock'], int):
        last_val = int(d['entries'][0]['stock'].replace("in stock", "").replace(",", "").replace("R", ""))

    for data in d['entries']:
        _d = data['stock']
        if not isinstance(data['stock'], int):
            _d = int(data['stock'].replace("in stock", ""))
        if _d < last_val:
            sales += 1
        last_val = _d

    return sales

out = []

for g in data:
    _id = g['_id']
    sales = get_sales(g)
    price = g['entries'][0]['price']
    url = g['entries'][0]['url']
    if not isinstance(g['entries'][0]['price'], float):
        if not isinstance(g['entries'][0]['price'], int):
            price = price.replace(",", "").replace("R", "")

    rev = sales * float(price)
    out.append([float(sales), float(price), rev,_id, url])

out = sorted(out, key=lambda x: x[2])
total = float(0)
num = 1
num_big_ = 0
num_big_1 = 0
num_big_1_rev = 0
for d in reversed(out):
    print(d)
    num += 1
    total += d[2]
    if d[0] > 0:
        num_big_ += 1
    if d[1] > 1000:
        num_big_1 += 1
        num_big_1_rev += d[2]

# print(float(num_big_1_rev)/ float(total))
# print(float(num_big_1)/ float(num))
# print(float(num_big_)/ float(num))
# print(num_big_1)
# print(num_big_)
# print(num_big_1_rev)
print(total)
    # print(f"group: {_id} sales: {sales} revenue: {rev}")


