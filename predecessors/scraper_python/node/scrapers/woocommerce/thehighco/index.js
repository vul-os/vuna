// var thc = require('./node/scrapers/woocommerce/thehighco/run.js');
var URL = 'https://thehighco.co.za/shop/';

var request = require('request');
const cheerio = require('cheerio');

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

function getMaxPages(url) {
  return 1;
}

function getAllItems(url) {
  const maxPages = getMaxPages(url);
  const pagenationStr = "page/"
  const pages = [...Array(maxPages).keys()].map(i => url + pagenationStr + String(i + 1));
  return Promise.all(pages.map(pageUrl => getItemsOnPage(pageUrl)));
}

const getProductData = (url) => {
  return new Promise(resp => request(url, (err, res, body) => {
    const list = [];
    if (!err && res.statusCode == 200) {
      const $ = cheerio.load(body);
      const _main = $('div[class="summary entry-summary"]');
      const name = _main.find('h1[class="product_title entry-title"]').text();
      const sku = _main.find('span[class="sku"]');
      const stock = _main.find('input[class="input-text qty text"]').attr('max');
      console.log(stock);
    }
    resp(list)
  }));
}

const main = async (url) => {

  const categories = await getItemsOnPage(url);
  const productLinks = await Promise.all(categories.map(getAllItems))
  const merged = [].concat.apply([], [].concat.apply([], productLinks));
  await Promise.all(merged.map(getProductData))

  // const data = await Promise.all(productLinks.map(inner => inner.map(func)))
  console.log(merged);
}

// getItemsOnPage(URL);
try {
  getProductData('https://thehighco.co.za/product/ocb-premium-kingsize-paper');

  // main(URL);
} catch (e) {
  console.log(e)
}
