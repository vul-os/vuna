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
      $('ul[class="products clearfix products-3"]').find('li').each((index, element) => {
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

const getAllItem = async (url) => {
    const maxPages = await getMaxPages(url);
    if (!isNaN(maxPages)) {
        const pagenationStr = "page/"
        const pages = [...Array(maxPages).keys()].map(i => url + pagenationStr + String(i + 1));
        return Promise.all(pages.map(pageUrl => getItemsOnPage(pageUrl)));
    } else {
        return Promise.all([getItemsOnPage(url)]); 
    }
}

const getAllItems = async (urls) => {
    return Promise.all(urls.map(pageUrl => getAllItem(pageUrl)));
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
  return new Promise(resp => request(url, (err, res, body) => {
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
        "url": url,
        "date": new Date(Date.now()).toISOString()
      }
      console.log(item);
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

const main = async (url, mongoURL, dbName) => {
  const start = new Date();
  const categories = await getItemsOnPage(url);
  const murgedCategories = flatten(categories);

  const productLinks = await getAllItems(murgedCategories);
  const murgedProducts = flatten(productLinks);
  await saveAllItems(mongoURL, dbName, murgedProducts);


  await allProgress(murgedProducts.map(function(x) { return getProductData(x, mongoURL, dbName); }),
  (p) => {
     console.log(`Products The High Co = ${p.toFixed(2)} %`);
  });
  const end = new Date() - start;
  console.log(`end = ${end.toFixed(2)}`);
}

try {
  var URL = 'https://thehighco.co.za/shop/';
  const mongoURL = 'mongodb://localhost:27017';  
  const dbName = 'thehighco';
  
  main(URL, mongoURL, dbName);
} catch (e) {
  console.log(e)
}
