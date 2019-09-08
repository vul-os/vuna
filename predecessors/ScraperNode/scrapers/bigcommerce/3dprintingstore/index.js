const MongoClient = require('mongodb').MongoClient;
const cheerio = require('cheerio');
const Promise = require("bluebird");
const request = require("request");
const _ = require('lodash');
var tr = require('tor-request');
var RateLimiter = require('limiter').RateLimiter;
var random_useragent = require('random-useragent');

// var limiter = new RateLimiter(1, 50); // at most 1 request every 100 ms
// var throttledRequest = function() {
//     var requestArgs = arguments;
//     limiter.removeTokens(1, function() {
//         tr.request.apply(this, requestArgs);
//     });
// };


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
  return new Promise(resp => tr.request(url, (err, res, body) => {
    const list = [];
    if (!err && res.statusCode == 200) {
      const $ = cheerio.load(body);
      $('ul[class="ProductList"]').find('li').each((index, element) => {
        list.push($(element).find('a').first().attr('href'));
      });
    }
    resp(list)
  }));
}

const getMaxPages = (url) => {
    return new Promise(resp => tr.request(url, (err, res, body) => {
      let page = 1;
      if (!err && res.statusCode == 200) {
        const $ = cheerio.load(body);
        const tempPage = $('ul[class="PagingList"]').find('li');
        const tempPageInt = parseInt($(tempPage[tempPage.length - 1]).find('a').text());
        if (!isNaN(tempPageInt)) {
            page = tempPageInt
        }
      }
      resp(page)
    }));
}

const getAllItems = async (url) => {
  const maxPages = await getMaxPages(url);
  const pagenationStr = "?page="
  const pages = [...Array(maxPages).keys()].map(i => url + pagenationStr + String(i + 1));
  return Promise.all(pages.map(pageUrl => getItemsOnPage(pageUrl)));
}

const getAll = async (pages) => {
    return Promise.all(pages.map(pageUrl => getAllItems(pageUrl)));
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
  return new Promise(resp => tr.request({url: `${url}?setCurrencyId=1`, 'User-Agent': random_useragent.getRandom()}, (err, res, body) => {
    if (!err && res.statusCode == 200) {
      const $ = cheerio.load(body);
      const _main = $('div[class="ContentArea"]');

      const name = _main.find('h1[class="title"]').text().trim();
      const price = _main.find('span[class="ProductDetailsPriceIncTax"]').text().trim();
      const stock = _main.find('div[class="DetailRow InventoryLevel"]').find('div[class="Value"]').text().trim();
      item = {
        "name": name,
        "price": price,
        "stock": stock,
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
      console.log("error getting product data");
      resp({url: url, success: false});
    }
    resp({url: url, success: true});
  }));
}

const getCategories = (url) => {
  return new Promise(resp => tr.request(url, (err, res, body) => {
    const list = []
    if (!err) {
      const $ = cheerio.load(body);
      $('div[class="SitemapCategories"]').find('ul[class=""]').children("li").each((index, element) => {
        list.push($(element).find('a').first().attr('href'));
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

function shuffle(a) {
  for (let i = a.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1));
      [a[i], a[j]] = [a[j], a[i]];
  }
  return a;
}


const main = async () => {
    const url = 'http://www.3dprintingstore.co.za/sitemap/categories/';
    const mongoURL = 'mongodb://localhost:27017';  
    const dbName = '3dprintingstore';
    while (true) {
      const start = new Date();
      const categories = await getCategories(url);
      const productLinks = await getAll(categories);
      let murgedLinks = shuffle([...new Set(flatten(productLinks))]);
      
      saveAllItems(mongoURL, dbName, murgedLinks);

      const resp = await allProgress(murgedLinks.map(function(x) { return getProductData(x, mongoURL, dbName); }),
      (p) => {
        console.log(`Products 3DPrintingStore = ${p.toFixed(2)} %`);
      });

      const end = (new Date() - start) / 1000;
      console.log(`end = ${end.toFixed(2)}`);
    }

   
}

module.exports.main = main;
  

// try {
//   var URL = 'http://www.3dprintingstore.co.za/sitemap/categories/';
//   const mongoURL = 'mongodb://localhost:27017';  
//   const dbName = '3dprintingstore';
  
//   main(URL, mongoURL, dbName);
// } catch (e) {
//   console.log(e)
// }
