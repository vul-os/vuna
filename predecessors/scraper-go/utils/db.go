package utils

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog/log"
	"strconv"
	"strings"
	"time"
)

var Pool *pgxpool.Pool


func UpsertItem(tableName string, nameField string, name string, urlItem string, storeId int) {
	queryBase := `
		INSERT INTO {{ .TableName }} ({{ .NameField }}_name, url, store, date_added, date_updated)
		VALUES('{{ .Name }}', '{{ .UrlItem }}', {{ .StoreId }},
			 $1::date, $1::date)
		ON CONFLICT (url) 		
		WHERE store = {{ .StoreId }}
		DO UPDATE SET
			{{ .NameField }}_name = '{{ .Name }}',
			url = '{{ .UrlItem }}',
			store = {{ .StoreId }},
			date_updated = $1::date
		RETURNING id;`
	r := strings.NewReplacer(
		"{{ .TableName }}", tableName,
				"{{ .NameField }}", nameField,
				"{{ .Name }}", name,
				"{{ .UrlItem }}", urlItem,
				"{{ .StoreId }}", strconv.Itoa(storeId),
		)
    query := r.Replace(queryBase)
	//fmt.Println(query)
	c, err := Pool.Acquire(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Cannot acquire connection")
	}
	defer c.Release()
	var resultId int
	err = c.QueryRow(context.Background(), query, time.Now().UTC()).Scan(&resultId)
	if err != nil {
		log.Error().Err(err).Msg("QueryRow failed")
	}
	log.Info().Msg(fmt.Sprintf("DB Insert ResultID: %d", resultId))
}

func GenerateConnPool() {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", "scrapers",
		"scrapers", "38.17.53.117", 17435, "scrapers_ecom")
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Error().Err(err).Msg("Error configuring the database")
	}

	Pool, err = pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		log.Error().Err(err).Msg("Error connecting to the database")
	}
}

