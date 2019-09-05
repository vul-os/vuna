const MongoClient = require('mongodb').MongoClient;
const request = require('request').defaults({ jar: true });
const cheerio = require('cheerio');
const batchPromises = require('batch-promises');

async function runAllQueries(queries) {
  const batches = _.chunk(queries, 2);
  const results = [];
  while (batches.length) {
    const batch = batches.shift();
    const result = await Promises.all(batch.map(runQuery));
    results.push(result)
  }
  return _.flatten(results);
}

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

const getItemsOnPage = (url, cookieJar) => {
    return new Promise(resp => request({ url: url, jar: cookieJar }, (err, res, body) => {
        const list = [];
        if (!err && res.statusCode == 200) {
            const $ = cheerio.load(body);
            $('div[class="shop-container"]').find('div[class="image-fade_in_back"]').each((index, element) => {
                list.push($(element).find('a').first().attr('href'));
            });
        } else {
            console.log("error \n");
        }
        resp(list)
    }));
}

const getMaxPages = (url, cookieJar) => {
    return new Promise(resp => request({ url: url, jar: cookieJar }, (err, res, body) => {
        let page = 1;
        if (!err && res.statusCode == 200) {
            const $ = cheerio.load(body);
            var tempPage = $('nav[class="woocommerce-pagination"]').find('ul[class="page-numbers nav-pagination links text-center"]').find('li')
            page = $(tempPage[tempPage.length - 2]).find('a').text()
        } else {
            console.log("error \n");
        }
        resp(parseInt(page))
    }));
}
function chunkArray(myArray, chunk_size){
    var results = [];
    while (myArray.length) {
        results.push(myArray.splice(0, chunk_size));
    }
    return results;
}

const getAllItems = async (url, cookieJar) => {
    const maxPages = await getMaxPages(url, cookieJar);
    const pagenationStr = "page/"
    const pages = [...Array(maxPages).keys()].map(i => url + pagenationStr + String(i + 1));
    return await allProgress(pages.map(function (x) { return getItemsOnPage(x, cookieJar); }),
        (p) => {
            console.log(`Product Links biltongandbudz = ${p.toFixed(2)} %`);
        });
}

const saveAllItems = (mongoURL, dbName, items) => {
    const client = new MongoClient(mongoURL, { useNewUrlParser: true });
    client.connect(function (err) {
        const db = client.db(dbName);
        const collection = db.collection('productList');
        collection.insertOne({ "products": items }, function (err, result) {
            client.close();
        });
    });
}
const getProductData = (url, mongoURL, dbName) => {
      const client = new MongoClient(mongoURL, {useNewUrlParser: true});
    const headers_ = {
        'User-Agent': 'python-requests/2.22.0',
        'Accept': '*/*',
    }
    return new Promise(resp => request({url: url, headers: headers_, forever: true}, (err, res, body) => {
        if (!err && res.statusCode == 200) {
            const $ = cheerio.load(body);

            const name = $('h1[class="product-title product_title entry-title"]').text().trim();
            const price = $('div[class="product-info summary col-fit col entry-summary product-summary"]').find('span[class="woocommerce-Price-amount amount"]').text().trim();
            const stock = $('p[class="stock in-stock"]').text().trim()

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
        } else {
            console.log("error \n");
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

    const j = request.jar();
    const cookieStr = "age_gate=21&tk_ai=woo:dBlQRi3iIybLcJZIZaok+wyL"
    const cookie = request.cookie(cookieStr);
    j.setCookie(cookie, url);

    const start = new Date();
    const productLinks = await getAllItems(url, j);
    const murgedLinks = flatten(productLinks);
    console.log(murgedLinks.length);
    await saveAllItems(mongoURL, dbName, murgedLinks);

    await allProgress(murgedLinks.map(function(x) { return getProductData(x, mongoURL, dbName); }),
    (p) => {
       console.log(`Products Biltongandbudz = ${p.toFixed(2)} %`);
    });
    const end = new Date() - start;
    console.log(`end = ${end.toFixed(2)}`);

}

try {
    var URL = 'https://www.biltongandbudz.co.za/shop/';
    const mongoURL = 'mongodb://localhost:27017';
    const dbName = 'biltongandbudz';

    main(URL, mongoURL, dbName);
} catch (e) {
    console.log(e)
}
