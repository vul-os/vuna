addr = "192.168.8.120"
port = 27017

from datetime import datetime
from bson.code import Code
import pymongo
conn = pymongo.MongoClient(host="localhost:27017")
db = conn.trophyseeds

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
            "stock": {
                  "$addToSet": "$stock"
            }
        }
    },
    {"$sort": {"date": -1}}
])

sales = 0

for d in data:
    print(d)
