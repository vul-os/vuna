const MongoClient = require('mongodb').MongoClient;
const cheerio = require('cheerio');
var RateLimiter = require('limiter').RateLimiter;
var tr = require('tor-request');
var random_useragent = require('random-useragent');

var limiter = new RateLimiter(1, 1000);
var throttledRequest = function() {
    var requestArgs = arguments;
    limiter.removeTokens(1, function() {
        tr.request.apply(this, requestArgs);
    });
};


function allProgress(proms, progress_cb) {
  let d = 0;
  progress_cb(0);
  for (const p of proms) {
    p.then(()=> {    
      d ++;
      progress_cb( (d * 100) / proms.length );
    });
  }
  return Promise.all(proms);
}

const getItemsOnPage = (url) => {
  return new Promise(resp => tr.request({ url: url, 'User-Agent': random_useragent.getRandom()}, (err, res, body) => {
    const list = [];
    if (!err && res.statusCode == 200) {
      const $ = cheerio.load(body);
      $('main[id="main"]').find('ul[class="products columns-3"]').find('li').each((index, element) => {
        list.push($(element).find('a').first().attr('href'));
      });
    }
    resp(list)
  }));
}

const getMaxPages = (url) => {
    return new Promise(resp => tr.request({ url: url, 'User-Agent': random_useragent.getRandom()}, (err, res, body) => {
      let page = 1;
      if (!err && res.statusCode == 200) {
        const $ = cheerio.load(body);
        const tempPage = $('nav[class="woocommerce-pagination"]').find('li');
        const tempPageInt = parseInt($(tempPage[tempPage.length - 2]).find('a').text());
        if (!isNaN(tempPageInt)) {
            page = tempPageInt
        }
      }
      resp(page)
    }));
}

const getAllItems = async (url) => {
  const maxPages = await getMaxPages(url);
  const pagenationStr = "page/"
  const pages = [...Array(maxPages).keys()].map(i => url + pagenationStr + String(i + 1));
  return Promise.all(pages.map(pageUrl => getItemsOnPage(pageUrl)));
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

const getProductData = (url, mongoURL, dbName) => {
  const client = new MongoClient(mongoURL, {useNewUrlParser: true});
  return new Promise(resp => throttledRequest({ url: url, 'User-Agent': random_useragent.getRandom()}, (err, res, body) => {
    if (!err && res.statusCode == 200) {
      const $ = cheerio.load(body);
      const _main = $('div[class="summary entry-summary"]');

      const name = _main.find('h1[class="product_title entry-title"]').text();
      const price = _main.find('span[class="woocommerce-Price-amount amount"]').text();
      const stock = _main.find('p[class="stock in-stock"]').text()

      item = {
        "name": name,
        "price": price,
        "stock": stock,
        "date": new Date(Date.now()).toISOString()
      }
      client.connect(function(err) { 
        const db = client.db(dbName);
        const collection = db.collection('data');
        collection.insertOne(item, function(err, result) {
          client.close();
        });
      });
    }
    resp("yes")
  }));
}

const getCategories = (url) => {
  return new Promise(resp => request(url, (err, res, body) => {
    const list = []
    if (!err) {
      const $ = cheerio.load(body);
      $('div[class="widget woocommerce widget_product_categories"]').find('li').each((index, element) => {
        const ele = $(element);
        if (!(ele.find('li').length > 0)) {
          list.push(ele.find('a').attr('href'));
        }
      });
    }
    resp(list)
  }));
}

function flatten(arr) {
    return arr.reduce(function (flat, toFlatten) {
      return flat.concat(Array.isArray(toFlatten) ? flatten(toFlatten) : toFlatten);
    }, []);
}
  

const main = async () => {
    const url = 'https://sacredseeds.co.za/shop/';
    const mongoURL = 'mongodb://localhost:27017';  
    const dbName = 'sacredseeds';
    while (true) {
      const start = new Date();
      const categories = await getItemsOnPage(url);
      // const productLinks = await Promise.all(categories.map(getAllItems));
      const productLinks = await allProgress(categories.map(function(x) { return getAllItems(x); }),
      (p) => {
          console.log(`Products Links Sacred Seeds = ${p.toFixed(2)} %`);
      });
      const murgedProducts = flatten(productLinks);
      await saveAllItems(mongoURL, dbName, murgedProducts);

      await allProgress(murgedProducts.map(function(x) { return getProductData(x, mongoURL, dbName); }),
      (p) => {
          console.log(`Products Sacred Seeds = ${p.toFixed(2)} %`);
      });
      const end = new Date() - start;
      console.log(`end sacred seeds = ${end.toFixed(2)}`);
    }
}

module.exports.main = main;

// try {
//   var URL = 'https://sacredseeds.co.za/shop/';
//   const mongoURL = 'mongodb://localhost:27017';  
//   const dbName = 'sacredseeds';
  
//   main(URL, mongoURL, dbName);
// } catch (e) {
//   console.log(e)
// }
