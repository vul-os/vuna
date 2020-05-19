from flask import Flask, request, jsonify
from bs4 import BeautifulSoup
import requests
from pprint import pprint
from bson.code import Code
import pymongo
import datetime

app = Flask(__name__)

addr = "192.168.8.120"
port = 27017

conn = pymongo.MongoClient(host="192.168.8.120:27017")
db = conn.fashionworld


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


def get_graph(d):
    data_ = []
    url = d['entries'][0]['url']
    price = float(d['entries'][0]['price'].replace("R", "").replace(",", "."))

    last_stock = 0
    for data in d['entries']:
        stock = data['stock']
        diff = last_stock - stock
        if diff < 0:
            diff = 0
        date = data['date']
        rev = diff * price
        data_.append([date, diff, rev])
    return url, data_


def get_to_from_date(date_range):
    to_date = datetime.datetime.now()

    from_date = datetime.datetime.now() - datetime.timedelta(days=date_range)
    from_date = from_date

    return to_date, from_date


def process_table(date_range):
    to_, from_ = get_to_from_date(date_range)

    data = db.data.aggregate(
    [
        {
            "$match": {
                "date": {
                    '$lte': to_, '$gte': from_
                }
            }
        },
        {
            "$group": {
                "_id": {
                      "name": "$name",
                      "sizeId": "$sizeId"
                },
                "entries": {
                  "$push": {
                    "stock": "$stock",
                    "price": "$price",
                    "url": "$url",
                    "name": "$name",
                    "date": "$date"
                  }
                }
            }
        },
        {"$sort": {"date": -1}}
    ], allowDiskUse=True)

    total_revenue = 0
    total_stock = 0
    num_items_sold = 0

    table_data = []

    for d in data:
        sales, stock_level = get_sales(d)
        price = float(d['entries'][0]['price'].replace("R", "").replace(",", "."))
        url = d['entries'][0]['url']

        rev = float(sales)*float(price)
        stock = float(stock_level)*float(price)

        table_data.append([url, rev, stock, sales])

        total_revenue += rev
        total_stock += stock
        num_items_sold += sales

    out_table_data = sorted(table_data, key=lambda x: x[1])[::-1]

    return total_revenue, total_stock, num_items_sold, out_table_data[0:25]


def scrape_more_data(url):
    r = requests.get(url)
    raw_html = r.content
    soup = BeautifulSoup(raw_html, 'html.parser')
    pic = soup.findAll('div', {'data-component': 'productGallery'})
    if len(pic) is 0:
        pic = ""
    else:
        pic = pic[0]
        pic = pic.findAll('img')[0]['src']
    name = soup.find('div', {'class': 'small-12 columns title-price no-padding'}).find('h1').getText()
    if name is None:
        name = ""
    return pic, name


def make_graphs(url, db, date_range):
    to_, from_ = get_to_from_date(date_range)
    data = db.data.find({
        "url": url,
        "date": {
            '$lte': to_, '$gte': from_
        }
    })

    d_ = []
    for d in data:
        print(d)
        d_.append(d)

    return data


@app.route('/api/get_data/', methods=['GET', 'POST'])
def add_message():
    content = request.json
    date_range = content['date_range']

    total_revenue, total_stock, num_items_sold, table_data = process_table(date_range)
    top_five_items = table_data[0:5]

    d_ = []
    for item in table_data:
        url = item[0]
        # pic, name = scrape_more_data(url)
        d_.append([url, item[1], item[2]])

    g_ = []
    for i in top_five_items:
        url = i[0]
        graph = make_graphs(url, db, date_range)
        # pic, name = scrape_more_data(url)

        g_.append([url, graph])

    print(g_)

    return jsonify({
        "totalRevenue":total_revenue,
        "totalStock": total_stock,
        "itemsSold": num_items_sold,
        "tableData": d_,
        "graph": g_
    })

if __name__ == '__main__':
    app.run(host= '0.0.0.0',debug=True)