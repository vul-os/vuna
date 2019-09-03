const MongoClient = require('mongodb').MongoClient;
const request = require('request');
const cheerio = require('cheerio');

const getMaxPages = async (url, pages) => {
    url = constructUrl(url, 1, 1, 70);
    const page = await loadPage(url);
    const num_products = parseInt(page['total_product']);
    const num_pages = Math.ceil(num_products / pages);
    return num_pages
}

const constructUrl = (baseURL, t, page, limit) => {
  let items = {
    "t": (t) ? t.toString() : '',
    "q": "",
    "shop": "communica-south-africa.myshopify.com",
    "page": (page) ? page.toString() : '',
    "limit": (limit) ? limit.toString() : '',
    "sort": "best-selling",
    "display": "grid",
    "collection_scope": "",
    "product_available": "false",
    "variant_available": "false",
    "build_filter_tree": "false",
    "check_cache": "false",
    "callback": "BCSfFilterCallback",
    "event_type": "init"
  }

  let urlStr = Object.keys(items).map(function(key, index) {
    return `${key}=${items[key]}`;
  }).join("&");

  return `${baseURL}?${urlStr}`;
}

const getPageJson = (body) => {
  const strList = [
    "/**/",
    "typeof",
    "BCSfFilterCallback === 'function' && BCSfFilterCallback("
  ]

  let rawHtml = body;
  strList.forEach(function (item, index) {
    rawHtml = rawHtml.replace(item, "");
  });
  rawHtml = rawHtml.slice(0, -2); 
  return JSON.parse(rawHtml)
}

const loadPage = (url) => {
  return new Promise(resp => request(url, (err, res, body) => {
    resp(getPageJson(body));
  }));
}

const handlePage = (page, mongoClient, dbName) => {
  return new Promise(resp => { 
     let items = [];  
     page['products'].forEach(function (item, index) {
      const varient = item["variants"][0];
      itemData = {
        "name": varient['sku'],
        "price": varient['price'],
        "avail": varient['available'],
        "stock": varient['inventory_quantity'],
        "date": new Date(Date.now()).toISOString()
      };
      items.push(itemData);
      const db = mongoClient.db(dbName);
      const collection = db.collection('data');
      collection.insertOne(itemData);
     });
     resp(items);
  });
}

const getAllItems = async (url, t, limit, mongoURL, dbName) => {
  const client = new MongoClient(mongoURL, {useNewUrlParser: true});
  await client.connect(function(err) { 
  });

  const maxPages = await getMaxPages(url, 70);
  const pages = [...Array(maxPages).keys()].map(i => constructUrl(url, t, i+1, limit));
  const pagesData = await Promise.all(pages.map(pageUrl => loadPage(pageUrl)));
  const data = await Promise.all(pagesData.map(data => handlePage(data, client, dbName)));
  console.log(flatten(data).length);
  client.close();
}

function flatten(arr) {
  return arr.reduce(function (flat, toFlatten) {
    return flat.concat(Array.isArray(toFlatten) ? flatten(toFlatten) : toFlatten);
  }, []);
}



const main = async (url, mongoURL, dbName) => {
    await getAllItems(url, 1567545942552, 70, mongoURL, dbName);
    // const out = constructUrl(url, 1, 1, 70);
    // const val = await loadPage(out);
    // val['products'].forEach(function (item, index) {
    //   console.log(item["varients"]);
    // });
}

try {
  var URL = "https://services.mybcapps.com/bc-sf-filter/filter";
  const mongoURL = 'mongodb://localhost:27017';  
  const dbName = 'communica';
  
  main(URL, mongoURL, dbName);
} catch (e) {
  console.log(e)
}