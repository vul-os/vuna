// //woocommerce
const biltongAndBudz = require('./scrapers/woocommerce/biltongandbudz/index.js');
// const bongAlong = require('./scrapers/woocommerce/bongalong/index.js');
const cannaBru = require('./scrapers/woocommerce/cannabru/index.js');
// const theBeanBank = require('./scrapers/woocommerce/cannabru/index.js');
const trophySeeds = require('./scrapers/woocommerce/trophyseeds/index.js');
const marajuanaSa = require('./scrapers/woocommerce/marajuanasa/index.js');
const sacredSeeds = require('./scrapers/woocommerce/sacredseeds/index.js');


// // shopify
const communica = require('./scrapers/shopify/communica/index.js');

// // zurb 
// const fashionWorld = require('./scrapers/zurb/fashionworld/index.js');

// // misc
// const shelfLife = require('./scrapers/misc/shelflife/index.js');

// // bigcommerce
const threeDPrintingStore = require('./scrapers/bigcommerce/3dprintingstore/index.js');


//53799.00
//57610.00


var cluster = require('cluster');
var numCPUs = require('os').cpus().length;

if (cluster.isMaster) {
  // Fork workers.
  for (var i = 0; i < numCPUs; i++) {
    cluster.fork();
  }

   Object.keys(cluster.workers).forEach(function(id) {
    console.log("I am running with ID : "+cluster.workers[id].process.pid);
  });

  cluster.on('exit', function(worker, code, signal) {
    console.log('worker ' + worker.process.pid + ' died');
  });
} else {

    (async () => {
        // const start = new Date();
        let proms = [communica.main(), threeDPrintingStore.main(), 
            trophySeeds.main(), biltongAndBudz.main(), marajuanaSa.main(), sacredSeeds.main(), cannaBru.main(),
            fashionWorld.main(), shelfLife.main()
        ]
        Promise.all(proms);
        // // await threeDPrintingStore.main();
        // // await trophySeeds.main();
        // const end = new Date() - start;
        // console.log(`end = ${end.toFixed(2)}`);
    })();
}