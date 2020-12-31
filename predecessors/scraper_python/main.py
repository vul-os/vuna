# main.py
import os
import subprocess

from flask import Flask
from flask import request

app = Flask(__name__)

@app.route('/', methods = ['GET', 'POST', 'DELETE'])
def run_spider():
    """
    /home/imran/.local/share/DBeaverData/workspace6/General/Scripts
    https://stackoverflow.com/questions/44820119/how-to-use-multiple-service-accounts-with-gcloud
    :return:
    :rtype:
    """
    spider_name = request.data.decode('utf-8')
    spider_name = spider_name if spider_name else None
    print(f"Starting Spyder: {spider_name}")
    if spider_name is not None:
        subprocess.check_output(['scrapy', 'crawl', spider_name, "-o", "output.json"], cwd=f'./scrapers/{spider_name}')
        # with open("output.json") as items_file:
        #     print(items_file.read())
    return spider_name


if __name__ == '__main__':
    app.run(debug=True, host="0.0.0.0", port=int(os.environ.get("PORT", 8080)))

