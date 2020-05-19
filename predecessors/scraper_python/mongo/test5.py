"""
-> Last Day
-> Last 7 Days
-> Last 30 Days
-> Last 90 Days
-> Month To Date
"""

"""
-> Total Revenue
-> Total Stock Value
-> Number of Items sold
"""

"""
graphs
-> Total Revenue
-> Number of Items Sold 
"""

from bson.code import Code
import pymongo



def get_sales(d):
    sales = 0
    last_val = d['entries'][0]['stock']
    first_val = last_val
    if not isinstance(d['entries'][0]['stock'], int):
        last_val = int(d['entries'][0]['stock'].replace("in stock", "").replace(",", "").replace("R", ""))

    for data in d['entries']:
        _d = data['stock']
        if not isinstance(data['stock'], int):
            _d = int(data['stock'].replace("in stock", ""))
        if _d < last_val:
            sales += last_val-_d
        last_val = _d

    return sales, first_val


def process(db, datetime_range):

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
                    "url": "$url",
                    "name": "$name"
                  }
                }
            }
        },
        {"$sort": {"date": -1}}
    ], allowDiskUse=True)

    total_revenue = 0
    total_stock = 0
    num_items_sold = 0

    for d in data:
        _id = d['_id']
        sales, stock_level = get_sales(d)
        price = float(d['entries'][0]['price'].replace("R", "").replace(",", "."))
        url = d['entries'][0]['url']

        rev = float(sales)*float(price)
        stock = float(stock_level)*float(price)
        graph_data.append([url, rev, stock, sales])

        total_revenue += rev
        total_stock += stock
        num_items_sold += sales

    return total_revenue, total_stock, num_items_sold




if __name__ == "__main__":
    addr = "192.168.8.120"
    port = 27017

    conn = pymongo.MongoClient(host="192.168.8.120:27017")
    db = conn.biltongandbuds
    # collection = db.data

    total_revenue, total_stock, num_items_sold = process(db, None)

    print(total_revenue, total_stock)

