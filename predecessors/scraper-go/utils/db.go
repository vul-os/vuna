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


func GetStoreIdByUrl(url string) (int, error) {
	r := strings.NewReplacer(
		"www.", "",
		"https://", "",
		"http://", "",
	)
	url = r.Replace(url)
	fmt.Println(url)
	query := fmt.Sprintf(`
		SELECT id FROM stores WHERE url = '%s'
	`, url)
	c, err := Pool.Acquire(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Cannot acquire connection")
		return 0, err
	}
	defer c.Release()
	var resultId int
	err = c.QueryRow(context.Background(), query).Scan(&resultId)
	if err != nil {
		log.Error().Err(err).Msg("QueryRow failed")
		return 0, err
	}
	log.Info().Msg(fmt.Sprintf("Store ID: %d", resultId))
	return resultId, nil
}

func UpsertItem(tableName string, nameField string, name string, urlItem string, storeId int) (int, error) {
	queryBase := `
		INSERT INTO {{ .TableName }} ({{ .NameField }}_name, url, store, date_added, date_updated)
		VALUES('{{ .Name }}', '{{ .UrlItem }}', {{ .StoreId }},
			 $1::timestamptz, $1::timestamptz)
		ON CONFLICT (url)
		WHERE store = {{ .StoreId }}
		DO UPDATE SET
			{{ .NameField }}_name = '{{ .Name }}',
			url = '{{ .UrlItem }}',
			store = {{ .StoreId }},
			date_updated = $1::timestamptz
		RETURNING id;`
	r := strings.NewReplacer(
		"{{ .TableName }}", tableName,
		"{{ .NameField }}", nameField,
		"{{ .Name }}", name,
		"{{ .UrlItem }}", urlItem,
		"{{ .StoreId }}", strconv.Itoa(storeId),
	)
	query := r.Replace(queryBase)
	c, err := Pool.Acquire(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Cannot acquire connection")
		return 0, err
	}
	defer c.Release()
	var resultId int
	nowTime := time.Now().Local()
	err = c.QueryRow(context.Background(), query, nowTime).Scan(&resultId)
	if err != nil {
		log.Error().Err(err).Msg("QueryRow failed")
		return 0, err
	}
	log.Info().Msg(fmt.Sprintf("DB Insert (%s) ResultID: %d", tableName, resultId))
	return resultId, nil
}

func UpsertStore(name string, urlItem string) (int, error) {
	queryBase := `
		INSERT INTO stores  (store_name, url, date_added, date_updated)
		VALUES('{{ .Name }}', '{{ .UrlItem }}',
			 $1::timestamptz, $1::timestamptz)
		ON CONFLICT (url)
		DO UPDATE SET
			store_name = '{{ .Name }}',
			url = '{{ .UrlItem }}',
			date_updated = $1::timestamptz
		RETURNING id;`
	r := strings.NewReplacer(
		"{{ .Name }}", name,
		"{{ .UrlItem }}", urlItem,
	)
	query := r.Replace(queryBase)
	c, err := Pool.Acquire(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Cannot acquire connection")
		return 0, err
	}
	defer c.Release()
	var resultId int
	err = c.QueryRow(context.Background(), query, nowTime).Scan(&resultId)
	if err != nil {
		log.Error().Err(err).Msg("QueryRow failed")
		return 0, err
	}
	log.Info().Msg(fmt.Sprintf("DB Insert Store ResultID: %d", resultId))
	return resultId, nil
}

func UpsertAttributes(name string, storeId int) (int, error) {
	queryBase := `
		INSERT INTO attributes (attribute_name, store_id, date_added, date_updated)
		VALUES('{{ .Name }}', {{ .StoreId }},
			 $1::timestamptz, $1::timestamptz)
		ON CONFLICT (attribute_name, store_id)
		DO UPDATE SET
			attribute_name = '{{ .Name }}',
			store_id = {{ .StoreId }},
			date_updated = $1::timestamptz
		RETURNING id;`
	r := strings.NewReplacer(
		"{{ .Name }}", name,
		"{{ .StoreId }}", strconv.Itoa(storeId),
	)
	query := r.Replace(queryBase)
	c, err := Pool.Acquire(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Cannot acquire connection")
		return 0, err
	}
	defer c.Release()
	var resultId int
	nowTime := time.Now().Local()
	err = c.QueryRow(context.Background(), query, nowTime).Scan(&resultId)
	if err != nil {
		log.Error().Err(err).Msg("QueryRow failed")
		return 0, err
	}
	log.Info().Msg(fmt.Sprintf("DB Insert Attributes ResultID: %d", resultId))
	return resultId, nil
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

