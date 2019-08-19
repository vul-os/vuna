
const puppeteer = require('puppeteer');
var URL = 'https://thehighco.co.za/shop/';

const crawlUrl = async (url) => {
  const browser = await puppeteer.launch();
  // Open new tab.
  const page = await browser.newPage();
  await page.goto(url);
  // Evaluate code in a context of page and get your data.
  const result = await page.evaluate(document.querySelector("ul[class='products clearfix products-3']").querySelectorAll("li"), (element) => {
    return element.querySelector('a').href 
  });
  console.log(result);
  results.push(result);
  // Close it.
  await page.close();
}

const main = async (url) => {
  const browser = await puppeteer.launch();
  // Open new tab.
  const page = await browser.newPage();
  await page.goto(url);
  await page.waitForSelector('ul[class="products clearfix products-3"]');
  // Evaluate code in a context of page and get your data.
  const result = await page.evaluate(document.querySelector('ul[class="products clearfix products-3"]').querySelectorAll("li"), (element) => {
    return element.querySelector('a').href 
  });
  console.log(result);
  results.push(result);
  // Close it.
  await page.close();

  // res = await crawlUrl(url);
  // console.log(res);
  // const categories = await getItemsOnPage(url);
  // const productLinks = await Promise.all(categories.map(getAllItems))
  // const merged = [].concat.apply([], [].concat.apply([], productLinks));
  // await Promise.all(merged.map(getProductData))

  // // const data = await Promise.all(productLinks.map(inner => inner.map(func)))
  // console.log(merged);
}

main(URL);
// try {
//   main(URL);
// } catch (e) {
//   console.log(e)
// }


