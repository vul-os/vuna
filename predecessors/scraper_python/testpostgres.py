import psycopg2
import datetime

connection = psycopg2.connect(
    database="scrapers", user="scrapers", password="scrapers", host="38.17.53.117", port=17435)

#
# product_name = "asdasd"
# price = 10.56
# store_id = 21
# url = "https://yourpoese.com"
#
# product_id = 123
# stock = 2
#
#
cursor = connection.cursor()
# query1 = f"""
#   INSERT INTO products (name, store, url, price, date_added)
#   VALUES('{product_name}', {store_id}, '{url}', {price}, '{datetime.datetime.now()}')
#   ON CONFLICT (url) DO UPDATE SET
#         name = '{product_name}',
#         store = {store_id},
#         url = '{url}',
#         price = {price},
#         date_added = '{datetime.datetime.now()}'
#     RETURNING id;
#   """
# cursor.execute(query1)
# connection.commit()
#
# lastid = cursor.fetchone()
# print(lastid)
#
query = f"""
  INSERT INTO datapoint (product, stock, date_scraped, date_added)
  VALUES({product_id}, {stock}, '{datetime.datetime.now()}', '{datetime.datetime.now()}')
  RETURNING id;
  """

# query = f"""
#    INSERT INTO stores (name, date_added)
#    VALUES('Three D Printing Store', '{datetime.datetime.now()}')
# """

cursor.execute(query)
connection.commit()

lastid = cursor.fetchone()
print(lastid)
