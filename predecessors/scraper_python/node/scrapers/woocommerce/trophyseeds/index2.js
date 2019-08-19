const puppeteer = require('puppeteer');

const scrapeAllPages = async (url, browser) => {
    return scrapePage(url, browser);
    
}

const scrapePage = async (url, browser) => {
    let page = await browser.newPage();
    await page.setViewport({ width: 1920, height: 1080 });
    await page.setRequestInterception(true);

    page.on('request', (req) => {
        if(req.resourceType() == 'font' || req.resourceType() == 'image'){
            req.abort();
        }
        else {
            req.continue();
        }
    });
    await page.goto(url);
    const SELECTOR = 'ul[class="products columns-3"]'
    // await page.waitForSelector();
    const result = await page.$$eval(
        SELECTOR,
          nodes =>
            nodes.map(element => {
              console.log(element);
              return element;
            })    
    );
    // const result = await page.evaluate(() => {
    //     let results = []
    //     document.querySelector('ul[class="products columns-3"]').children, (element) => {
    //         console.log(element);
    //         result.push(element);
    //     }
    //     return results;
    // });

    return result;
}


const main = async (url, mongoURL, dbName) => {
    let browser = await puppeteer.launch({ headless: false });

    console.log(await scrapeAllPages(url, browser));
    return 0;

  }
  
  try {
    var URL = 'https://www.trophyseeds.com/shop/';
    const mongoURL = 'mongodb://localhost:27017';  
    const dbName = 'trophyseeds';
    
    main(URL, mongoURL, dbName);
  } catch (e) {
    console.log(e)
  }
  