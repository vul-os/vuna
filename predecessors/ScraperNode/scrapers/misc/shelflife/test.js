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

const getCategories = (url) => {
    return new Promise(resp => request(url, (err, res, body) => {
        let links = [];
        if (!err && res.statusCode == 200) {
          const $ = cheerio.load(body);
          $('body').find('div[class="container-fluid header_bottom"]').children().eq(1).find('div[role="navigation"]').find("li").each((index, element) => {
            links.push(`${url}${$(element).find('a').attr('href')}`);
          });
        }
        resp(links);
    }));
}

const getItemsOnPage = (baseUrl, url) => {
    return new Promise(resp => request(url, (err, res, body) => {
      const list = [];
      if (!err && res.statusCode == 200) {
        const $ = cheerio.load(body);
        $('div[class="row push_both push_top push_bottom light_row"]').find('div[class="col-xs-6 col-sm-3"]').each((index, element) => {
          const ele = $(element).find('a').first().attr('href');
          if (ele.includes("products/")) {
              list.push(`${baseUrl}${ele}`)
          }
        });
      }
      resp(list)
    }));
}

const getAllItemsCat = async (baseUrl, url) => {
    let pageStr = "?&page="
    const maxPage = 50;
    var links = [];
    for (var page = 1; page < maxPage; page++) {
        linkz = await getItemsOnPage(baseUrl, `${url}/${pageStr}${page.toString()}`);
        if (linkz.length > 0) {
            links.push(linkz);
        } else {
            break;
        }
    }
    return links
}

const getAllItems = async (url) => {
    const categories = await getCategories(url);
    const links = await allProgress(categories.map(function(x) { return getAllItemsCat(url, x); }),
    (p) => {
        console.log(`Links Shelflife = ${p.toFixed(2)} %`);
    });
    return links;
}

const getAllData = async (urls, mongoURL, dbName) => {
    await allProgress(urls.map(function(x) { return getProductData(x, mongoURL, dbName); }),
    (p) => {
        console.log(`Data Shelflife = ${p.toFixed(2)} %`);
    });
    return links;   
}

const getProductData = (url, mongoURL, dbName) => {
    // const client = new MongoClient(mongoURL, {useNewUrlParser: true});
    return new Promise(resp => request(url, (err, res, body) => {
      if (!err && res.statusCode == 200) {
        const $ = cheerio.load(body);
        const mainBlock = $('div[class="row light_row push_both push_top push_bottom"]')
        const gender = $('div[class="gender_block"]').text().trim();
        const name = $('h1[class="title"]').text().trim();
        const price = mainBlock.find('div[class="price"]').text().trim();
        const skuIdText = $('h3[class="text_pale"]').text().trim();
        const skuId = $('input[name="prod"]').attr('value');
        const sizes = [];

        $('select[id="size"]').find('option').each((index, element) => {
            const ele = $(element).text();
            if (!ele.includes('Select')) {
                const sze = ele.replace(' ', '+');
                const stock = getVariationStock(skuId, sze);
                sizes.push([sze, stock]);
            }
        });

        item = {
          "name": name,
          "gender": gender,
          "price": price,
          "skuId": skuId,
          "skuIdText": skuIdText,
          "url": url,
          "sizes": sizes,
          "date": new Date(Date.now()).toISOString()
        }
        
        return item;
        // client.connect(function(err) { 
        //   const db = client.db(dbName);
        //   const collection = db.collection('data');
        //   collection.insertOne(item, function(err, result) {
        //     client.close();
        //   });
        // });
      }
    })).then(items => {
        console.log(items);
    });
  }

const getVariationStock = (productID, size) => {
    const dataUrl = 'https://www.shelflife.co.za/ajax/prod_qty_dropdown.php'

    var cookieJar = request.jar();
    payload = {
        'prod': productID,
        'size': size,
        'qty': '0'
    }
    return new Promise(resp => request.post({url: dataUrl, jar: cookieJar, json: payload}, (err, res, body) => {
      var maxVal = 0;
      if (!err && res.statusCode == 200) {
        const $ = cheerio.load(body);
        if ($('option').length > 0) {
            $('option').each((index, element) => {
                const ele = parseInt($(element).attr('value'));
                if (ele > maxVal) {
                    maxVal = ele;
                }
            });
        }
      }
      resp(maxVal)
    }));
}


function flatten(arr) {
    return arr.reduce(function (flat, toFlatten) {
      return flat.concat(Array.isArray(toFlatten) ? flatten(toFlatten) : toFlatten);
    }, []);
}

const main = async (url, mongoURL, dbName) => {

    const links = await getAllItems(url);
    const murgedLinks = flatten(links);

    await getAllData(murgedLinks, mongoURL, dbName);
    console.log(murgedLinks.length);
    // await getAllItems(url);
}

try {
  var URL = 'https://www.shelflife.co.za/';
  const mongoURL = 'mongodb://localhost:27017';  
  const dbName = 'shelflife';
  
  main(URL, mongoURL, dbName);
} catch (e) {
  console.log(e)
}
