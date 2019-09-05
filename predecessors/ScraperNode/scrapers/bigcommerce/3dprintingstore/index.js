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
      $('ul[class="ProductList"]').find('li').each((index, element) => {
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
  return new Promise(resp => request(url, (err, res, body) => {
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

const getCategories = (url) => {
  return new Promise(resp => request(url, (err, res, body) => {
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
  

const main = async () => {
    var url = 'http://www.3dprintingstore.co.za/sitemap/categories/';
    const mongoURL = 'mongodb://localhost:27017';  
    const dbName = '3dprintingstore';

    const categories = await getCategories(url);
    const produclLinks = await getAll(categories);
    const murgedLinks = flatten(produclLinks);
    await saveAllItems(mongoURL, dbName, murgedLinks);

    await allProgress(murgedLinks.map(function(x) { return getProductData(x, mongoURL, dbName); }),
    (p) => {
        console.log(`Products 3D Printing Store = ${p.toFixed(2)} %`);
    });
    
}

module.exports.main = main;


  

// try {
//   var URL = 'http://www.3dprintingstore.co.za/sitemap/categories/';
//   const mongoURL = 'mongodb://localhost:27017';  
//   const dbName = '3dprintingstore';
  
//   threeDPrintingStore(URL, mongoURL, dbName);
// } catch (e) {
//   console.log(e)
// }
