package main

import (
	"fmt"
	"strings"
)

func scheduleJob(jobName string, jobUrl string) {

	cmd := fmt.Sprintf("gcloud scheduler jobs create http %s --uri \"%s\" --schedule \"%s\" " +
		"--message-body \"%s\"",
		jobName, "https://scraper-go-ddvea5y2ha-uc.a.run.app", "0 */6 * * *", jobUrl)
	fmt.Println(cmd)
	//out, err := exec.Command(cmd).Output()
	//
	//// if there is an error with our execution
	//// handle it here
	//if err != nil {
	//	fmt.Printf("%s", err)
	//	os.Exit(0)
	//}
	//output := string(out[:])
	//fmt.Println(output)
}

func utility() {
	var toScrapeMap = [...]string{
		"biltongandbudz.co.za",
		"marijuanasa.co.za"    ,
		"sacredseeds.co.za"     ,
		"thehighco.co.za"        ,
		"budbuddies.co.za"        ,
		"feedaseed.co.za"          ,
		"smokinggunseeds.co.za"     ,
		"botshop.co.za"              ,
		"solomonstackle.co.za"        ,
		"bikemarket.co.za"             ,
		"matrixwarehouse.co.za"         ,
		"usn.co.za"                      ,
		"mymonsters.co.za"                ,
		"soundimports.co.za"               ,
		"livestainable.co.za"               ,
		"mavericklangaming.co.za"            ,
		"bottic.co.za"                        ,
		"thebeltshop.co.za"                    ,
		"farahwatches.co.za"                    ,
	}



	for _, element := range toScrapeMap {
		baseUrl := strings.TrimSpace(string(element))
		storeNameRep := strings.NewReplacer(
			".co", "",
			".za", "",
			".com", "",
			"https://", "",
			"http://", "",
			"/", "",
		)
		urlRep := strings.NewReplacer(
			"https://", "",
			"http://", "",
			"/", "",
		)
		urlReplaced := urlRep.Replace(baseUrl)
		storeNameReplaced := storeNameRep.Replace(urlReplaced)

		scheduleJob(storeNameReplaced, urlReplaced)
	}
}
//gcloud scheduler jobs create http JOB --schedule=SCHEDULE --uri=URI


