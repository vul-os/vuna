# main.py
import os
import subprocess

from flask import Flask
from flask import request

app = Flask(__name__)

running_spiders = []

@app.route('/', methods = ['GET', 'POST', 'DELETE'])
def run_spider():
    spider_name = request.data.decode('utf-8')
    print(f"Starting Spyder: {spider_name}")
    if spider_name is not None and spider_name not in running_spiders:
        running_spiders.append(spider_name)
        print(f"Running Spiders: {running_spiders}")
        subprocess.check_output(['scrapy', 'crawl', spider_name, "-o", "output.json"], cwd=f'./{spider_name}')
        running_spiders.remove(spider_name)
        # with open("output.json") as items_file:
        #     print(items_file.read())

    return spider_name


if __name__ == '__main__':
    app.run(debug=True, host="0.0.0.0", port=int(os.environ.get("PORT", 8080)))

# https://stackoverflow.com/questions/44820119/how-to-use-multiple-service-accounts-with-gcloud