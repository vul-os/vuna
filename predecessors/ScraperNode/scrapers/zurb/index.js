const MongoClient = require('mongodb').MongoClient;
const request = require('request');
const cheerio = require('cheerio');

function allProgress(proms, progress_cb) {
    let d = 0;
    progress_cb(0);
    for (const p of proms) {
        p.then(() => {
            d++;
            progress_cb((d * 100) / proms.length);
        });
    }
    return Promise.all(proms);
}


const getMaxPages = (url) => {
    let page = 1;
    let newUrl = `${url}?page=${9999}`
    return new Promise(resp => request(newUrl, (err, res, body) => {
      if (!err && res.statusCode == 200) {
        const $ = cheerio.load(body);
        const tempPage = $('ul[class="pagination float-right"]').find('li[class="current"]').text().trim().replace("You're on page", "");
        const tempPageInt = parseInt(tempPage);
        if (!isNaN(tempPageInt)) {
            page = tempPageInt
        }
      }
      resp(page)
    }));
}

const getItemsOnPage = (url) => {
    return new Promise(resp => request(url, (err, res, body) => {
      const list = [];
      if (!err) {
        const $ = cheerio.load(body);
        $('div[class="large-9 columns"]').find('div[class="columns block product"]').each((index, element) => {
          list.push($(element).find('a').first().attr('href'));
        });
      }
      resp(list)
    }));
}

const getProductData = (url, mongoURL, dbName) => {
    const client = new MongoClient(mongoURL, {useNewUrlParser: true});
    return new Promise(resp => request(url, (err, res, body) => {
      if (!err) {
        const $ = cheerio.load(body);
        const data = $('div[data-component="productDetails"]').attr('data-product-variations');
        const name = $('div[class="small-12 columns title-price no-padding"]').text().trim();
        const price = $('div[class="price"]').find("span").text().trim().replace("R", "");
  
        item = {
          "name": name,
          "price": price,
          "stock": JSON.parse(data),
          "url": url,
          "date": new Date(Date.now()).toISOString()
        }
        client.connect(function(err) { 
          const db = client.db(dbName);
          const collection = db.collection('data');
          collection.insertOne(item, function(err, result) {
            client.close();
          });
        });
      } else {
          console.log("error");
      }
      resp("yes")
    }));
}

const saveAllItems = (mongoURL, dbName, items) => {
    const client = new MongoClient(mongoURL, {useNewUrlParser: true});
    client.connect(function(err) { 
      const db = client.db(dbName);
      const collection = db.collection('productList');
      collection.insertOne({"products": items}, function(err, result) {
        client.close();
      });
    });
  }

const getAllItems = async (url) => {
    const maxPages = await getMaxPages(url);
    const pagenationStr = "?page="
    const pages = [...Array(maxPages).keys()].map(i => url + pagenationStr + String(i + 1));
    return await allProgress(pages.map(function(x) { return getItemsOnPage(x); }),
    (p) => {
       console.log(`Product Links Fashionworld = ${p.toFixed(2)} %`);
    });
  }

function flatten(arr) {
    return arr.reduce(function (flat, toFlatten) {
        return flat.concat(Array.isArray(toFlatten) ? flatten(toFlatten) : toFlatten);
    }, []);
}



const main = async (url, mongoURL, dbName) => {
    const start = new Date();
    const items = await getAllItems(url);
    const murgedItems = flatten(items);
    await saveAllItems(mongoURL, dbName, murgedItems);

    await allProgress(murgedItems.map(function(x) { return getProductData(x, mongoURL, dbName); }),
    (p) => {
       console.log(`Product Data Fashionworld = ${p.toFixed(2)} %`);
    });
    const end = new Date() - start;
    console.log(`end = ${end.toFixed(2)}`);
}

try {
    var URL = 'https://www.fashionworld.co.za/products';
    const mongoURL = 'mongodb://localhost:27017';
    const dbName = 'fashionworld';

    main(URL, mongoURL, dbName);
} catch (e) {
    console.log(e)
}
