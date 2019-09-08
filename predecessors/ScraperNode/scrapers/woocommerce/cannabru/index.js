const MongoClient = require('mongodb').MongoClient;
const cheerio = require('cheerio');
var RateLimiter = require('limiter').RateLimiter;
var tr = require('tor-request');
var random_useragent = require('random-useragent');

var limiter = new RateLimiter(1, 200); // at most 1 request every 100 ms
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
      $('div[class="shop-container"]').find('div[class="products row row-small large-columns-3 medium-columns-3 small-columns-2"]').find("div").each((index, element) => {
        list.push($(element).find('div[class="image-fade_in_back"]').first().find('a').first().attr('href'));
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
      var tempPage = $('nav[class="woocommerce-pagination"]').find('li')
      page = $(tempPage[tempPage.length - 2]).find('a').text()
    }
    resp(parseInt(page))
  }));
}

const getAllItems = async (url) => {
  const maxPages = await getMaxPages(url);
  const pagenationStr = "page/"
  const pages = [...Array(maxPages).keys()].map(i => url + pagenationStr + String(i + 1));
  return Promise.all(pages.map(pageUrl => getItemsOnPage(pageUrl)));
}

const saveAllItems = (mongoURL, dbName, items) => {
  const client = new MongoClient(mongoURL, {native_parser: true});
  client.connect(function(err) { 
    const db = client.db(dbName);
    const collection = db.collection('productList');
    collection.insertOne({"products": items}, function(err, result) {
      client.close();
    });
  });
}

const getProductData = (url, mongoURL, dbName) => {
  const client = new MongoClient(mongoURL, {native_parser: true});
  return new Promise(resp => throttledRequest({ url: url, 'User-Agent': random_useragent.getRandom()}, (err, res, body) => {
    if (!err && res.statusCode == 200) {
      const $ = cheerio.load(body);
      const _main = $('div[class="product-info summary col-fit col entry-summary product-summary"]');

      const name = _main.find('h1[class="product-title product_title entry-title"]').text();
      const price = _main.find('span[class="woocommerce-Price-amount amount"]').text();
      const stock = _main.find('p[class="price product-page-price"]').find('p[class="stock in-stock"]').text()

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


function flatten(arr) {
    return arr.reduce(function (flat, toFlatten) {
      return flat.concat(Array.isArray(toFlatten) ? flatten(toFlatten) : toFlatten);
    }, []);
}
  

const main = async () => {
  const url = 'https://cannabru.online/cannabank/shop/';
  const mongoURL = 'mongodb://localhost:27017';  
  const dbName = 'cannabru';
  while (true) {
    const start = new Date();
    const productLinks = await getAllItems(url);
    const murgedProducts = [...new Set(flatten(productLinks))].filter(n => n);

    await saveAllItems(mongoURL, dbName, murgedProducts);
    // await Promise.all(murgedProducts.map(getProductData));
    await allProgress(murgedProducts.map(function(x) { return getProductData(x, mongoURL, dbName); }),
    (p) => {
      console.log(`Products Cannabru = ${p.toFixed(2)} %`);
    });
    const end = new Date() - start;
    console.log(`end cannabru= ${end.toFixed(2)}`);
  }
}

module.exports.main = main;

// try {
//   var URL = 'https://cannabru.online/cannabank/shop/';
//   const mongoURL = 'mongodb://localhost:27017';  
//   const dbName = 'cannabru';
  
//   main(URL, mongoURL, dbName);
// } catch (e) {
//   console.log(e)
// }
