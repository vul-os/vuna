const threeDPrintingStore = require('./scrapers/bigcommerce/3dprintingstore/index.js');
const trophySeeds = require('./scrapers/woocommerce/trophyseeds/index.js');



(async () => {
    
    await threeDPrintingStore.main();
    await trophySeeds.main();

})();


