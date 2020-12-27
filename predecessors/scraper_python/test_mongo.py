import pymongo
from bson.objectid import ObjectId

myclient = pymongo.MongoClient(
    "mongodb+srv://scraperama:scraperama@cluster0.i0xw4.mongodb.net/scrapers?retryWrites=true&w=majority")
mydb = myclient["scrapers"]
mycol = mydb["stores"]

mydict = { "name": "3D Printing Store", "url": "http://www.3dprintingstore.co.za", "test_id": ObjectId("666f6f2d6261722d71757578") }

x = mycol.insert_one(mydict)