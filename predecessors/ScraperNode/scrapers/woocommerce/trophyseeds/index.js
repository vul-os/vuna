const MongoClient = require('mongodb').MongoClient;
const request = require('request');
const cheerio = require('cheerio');

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
  return new Promise(resp => request(url, (err, res, body) => {
    const list = [];
    if (!err && res.statusCode == 200) {
      const $ = cheerio.load(body);
      $('ul[class="products columns-3"]').find('li').each((index, element) => {
        list.push($(element).find('a').first().attr('href'));
      });
    }
    resp(list)
  }));
}

const getMaxPages = (url) => {
  return new Promise(resp => request(url, (err, res, body) => {
    let page = 1;
    if (!err && res.statusCode == 200) {
      const $ = cheerio.load(body);
      var tempPage = $('nav[class="woocommerce-pagination"]').find('ul[class="page-numbers"]').find('li')
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
  return new Promise(resp => request(url, (err, res, body) => {
    const client = new MongoClient(mongoURL, {native_parser:true});
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

const main = async () => {
  const url = 'https://www.trophyseeds.com/shop/';
  const mongoURL = 'mongodb://localhost:27017';  
  const dbName = 'trophyseeds';
  while (true) {
    const start = new Date();
    const productLinks = await getAllItems(url);
    const murgedProducts = [].concat.apply([], [].concat.apply([], productLinks));
    await saveAllItems(mongoURL, dbName, murgedProducts);
    // await Promise.all(murgedProducts.map(getProductData));
    await allProgress(murgedProducts.map(function(x) { return getProductData(x, mongoURL, dbName); }),
    (p) => {
      console.log(`Products Trophy Seeds = ${p.toFixed(2)} %`);
    });
    const end = new Date() - start;
    console.log(`end trophyseeds = ${end.toFixed(2)}`);
  }
}

module.exports.main = main;

// try {
//   var URL = 'https://www.trophyseeds.com/shop/';
//   const mongoURL = 'mongodb://localhost:27017';  
//   const dbName = 'trophyseeds';
  
//   main(URL, mongoURL, dbName);
// } catch (e) {
//   console.log(e)
// }
