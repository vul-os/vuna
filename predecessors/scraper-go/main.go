package main

import (
	"scraper-go/scrapers"
	"scraper-go/utils"
)

var baseUrl = "https://www.botshop.co.za"

func main() {
	utils.GenerateConnPool()

	//dbpool, err := pgxpool.Connect(context.Background(), dbUrl)
	//if err != nil {
	//	fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
	//	os.Exit(1)
	//}
	//defer dbpool.Close()

	//c, err := utils.Pool.Acquire(context.Background())
	//if err != nil {
	//	log.Error().Err(err).Msg("Cannot acquire connection")
	//}
	//
	//var greeting string
	//err = c.QueryRow(context.Background(), "select 'Hello, world!'").Scan(&greeting)
	//if err != nil {
	//	fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
	//	os.Exit(1)
	//}
	//
	//fmt.Println(greeting)


	scrapers.Scrape(baseUrl)
	//utils.UpsertItem("tags", "tag", "wp-attr-testy", "https://store.com/testy", 1)
}