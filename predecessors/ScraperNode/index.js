const threeDPrintingStore = require('./scrapers/bigcommerce/3dprintingstore/index.js');
var URL = 'http://www.3dprintingstore.co.za/sitemap/categories/';
const mongoURL = 'mongodb://localhost:27017';  
const dbName = '3dprintingstore';
(async function () {
    var threeDPrintingStore = await threeDPrintingStore(URL, mongoURL, dbName);
    console.log(threeDPrintingStore);
})();



