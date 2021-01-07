package utils

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog/log"
	"os"
	"strconv"
	"strings"
)

var Pool *pgxpool.Pool

type Items struct {
	Name string
	ItemUrl  string
}

type ProdStruct struct {
	ProductName string
	ProductUrl string
	StoreId 	int
	Categories []Items
	Tags []Items
	VarID      int
	Sku		   string
	MaxQty     int
	Price      float32
	Attributes map[string]string
	AvailabilityHtml string
}


func DoAllDb(products []ProdStruct, productName string, productId int, storeId int, URL string,
	tagList []Items, catList []Items) {
	for _, prod := range products {
		log.Info().Msg(
			fmt.Sprintf(
				`Product Var Data -> Name: %s, URL: %s, Price: %f, Qty: %d, VarId: %d, Sku: %s, Attributes: %s, Tags: %s, Categories: %s`,
				productName,
				URL,
				prod.Price,
				prod.MaxQty,
				prod.VarID,
				prod.Sku,
				prod.Attributes,
				tagList,
				catList,
			),
		)
		// todo: fix this kak
		attr := prod.Attributes
		var attrId = -1

		if attr != nil {
			for k, v := range attr {
				attrId, _ = UpsertAttributes(k, v, storeId)
			}
		}
		productVarId, _ := UpsertProductAndProductVariations(
			productId, attrId, strconv.Itoa(prod.VarID), prod.Sku,
		)
		UpsertDatapoints(productVarId, prod.MaxQty, prod.Price)
	}
}

func GetStoreIdByUrl(url string) (int, error) {
	r := strings.NewReplacer(
		"www.", "",
		"https://", "",
		"http://", "",
	)
	url = r.Replace(url)
	query := fmt.Sprintf(`
		SELECT id FROM stores WHERE url = '%s'
	`, url)
	c, err := Pool.Acquire(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Cannot acquire connection (GetStoreIdByUrl)")
		return 0, err
	}
	defer c.Release()
	var resultId int
	err = c.QueryRow(context.Background(), query).Scan(&resultId)
	if err != nil {
		log.Error().Err(err).Msg("QueryRow failed (GetStoreIdByUrl)")
		return 0, err
	}
	log.Info().Msg(fmt.Sprintf("Store ID: %d", resultId))
	return resultId, nil
}


func UpsertItemAndProductItem(tableName string, nameField string, item string, urlItem string,
	productId int, storeId int) (int, error) {
	item = strings.ReplaceAll(item, "'", "''")
	urlItem = strings.ReplaceAll(urlItem, "'", "''")
	c, err := Pool.Acquire(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Cannot acquire connection")
		return 0, err
	}
	defer c.Release()
	queryBaseItem := `
		INSERT INTO {{ .TableName }} ({{ .NameField }}_name, url, store, date_added, date_updated)
		VALUES('{{ .Item }}', '{{ .UrlItem }}', {{ .StoreId }}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (url)
		WHERE store = {{ .StoreId }}
		DO UPDATE SET
			{{ .NameField }}_name = '{{ .Item }}',
			url = '{{ .UrlItem }}',
			store = {{ .StoreId }},
			date_updated = CURRENT_TIMESTAMP
		RETURNING id;`
	r := strings.NewReplacer(
		"{{ .TableName }}", tableName,
		"{{ .NameField }}", nameField,
		"{{ .Item }}", item,
		"{{ .UrlItem }}", urlItem,
		"{{ .StoreId }}", strconv.Itoa(storeId),
	)
	queryItem := r.Replace(queryBaseItem)
	var resultId int
	err = c.QueryRow(context.Background(), queryItem).Scan(&resultId)
	if err != nil {
		log.Error().Err(err).Msg("QueryRow failed (UpsertItemAndProductItem)")
		return 0, err
	}
	log.Info().Msg(fmt.Sprintf("DB Insert (%s) ResultID: %d", tableName, resultId))

	queryBaseItemProduct := `
		INSERT INTO product_{{ .TableName }} (product_id, {{ .NameField }}_id, date_added, date_updated)
		VALUES({{ .ProductId }}, {{ .Item }}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (product_id, {{ .NameField }}_id)
		DO UPDATE SET
			product_id = {{ .ProductId }},
			{{ .NameField }}_id = {{ .Item }},
			date_updated = CURRENT_TIMESTAMP
		RETURNING id;`
	r = strings.NewReplacer(
		"{{ .TableName }}", tableName,
		"{{ .NameField }}", nameField,
		"{{ .Item }}", strconv.Itoa(resultId),
		"{{ .ProductId }}", strconv.Itoa(productId),
	)
	queryItem = r.Replace(queryBaseItemProduct)
	err = c.QueryRow(context.Background(), queryItem).Scan(&resultId)
	if err != nil {
		log.Error().Err(err).Msg("QueryRow failed (UpsertItemAndProductItem)")
		return 0, err
	}
	log.Info().Msg(fmt.Sprintf("DB Insert product_(%s) ResultID: %d", tableName, resultId))
	return resultId, nil
}


func UpsertItem(tableName string, nameField string, item string, urlItem string, storeId int) (int, error) {
	item = strings.ReplaceAll(item, "'", "''")
	urlItem = strings.ReplaceAll(urlItem, "'", "''")
	c, err := Pool.Acquire(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Cannot acquire connection")
		return 0, err
	}
	defer c.Release()
	queryBaseItem := `
		INSERT INTO {{ .TableName }} ({{ .NameField }}_name, url, store, date_added, date_updated)
		VALUES('{{ .Name }}', '{{ .UrlItem }}', {{ .StoreId }}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (url)
		WHERE store = {{ .StoreId }}
		DO UPDATE SET
			{{ .NameField }}_name = '{{ .Name }}',
			url = '{{ .UrlItem }}',
			store = {{ .StoreId }},
			date_updated = CURRENT_TIMESTAMP
		RETURNING id;`
	r := strings.NewReplacer(
		"{{ .TableName }}", tableName,
		"{{ .NameField }}", nameField,
		"{{ .Name }}", item,
		"{{ .UrlItem }}", urlItem,
		"{{ .StoreId }}", strconv.Itoa(storeId),
	)
	queryItem := r.Replace(queryBaseItem)

	var resultId int
	err = c.QueryRow(context.Background(), queryItem).Scan(&resultId)
	if err != nil {
		log.Error().Err(err).Msg("QueryRow failed (UpsertItem)")
		return 0, err
	}
	log.Info().Msg(fmt.Sprintf("DB Insert (%s) ResultID: %d", tableName, resultId))
	return resultId, nil
}

func UpsertStore(name string, urlItem string) (int, error) {
	c, err := Pool.Acquire(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Cannot acquire connection")
		return 0, err
	}
	defer c.Release()
	queryFind := fmt.Sprintf(`
		SELECT id FROM stores WHERE url = '%s'
	`, urlItem)
	queryBase := `
		INSERT INTO stores  (store_name, url, date_added, date_updated)
		VALUES('{{ .Name }}', '{{ .UrlItem }}', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (url)
		DO UPDATE SET
			store_name = '{{ .Name }}',
			url = '{{ .UrlItem }}',
			date_updated = CURRENT_TIMESTAMP
		RETURNING id;`
	r := strings.NewReplacer(
		"{{ .Name }}", name,
		"{{ .UrlItem }}", urlItem,
	)
	query := r.Replace(queryBase)
	var resultId int
	err = c.QueryRow(context.Background(), queryFind).Scan(&resultId)
	if err != nil {
		log.Error().Err(err).Msg(fmt.Sprintf("QueryRow failed -> Inserting Store: %s", name))
	}
	if resultId > 1 {
		return resultId, err
	}
	err = c.QueryRow(context.Background(), query).Scan(&resultId)
	if err != nil {
		log.Error().Err(err).Msg("QueryRow failed")
		return 0, err
	}
	log.Info().Msg(fmt.Sprintf("DB Insert Store ResultID: %d", resultId))
	return resultId, nil
}

func UpsertAttributes(name string, value string, storeId int) (int, error) {
	name = strings.ReplaceAll(name, "'", "''")
	value = strings.ReplaceAll(value, "'", "''")
	queryBase := `
		INSERT INTO attributes (attribute_name, attribute_value, store_id, date_added, date_updated)
		VALUES('{{ .Name }}', '{{ .Value }}', {{ .StoreId }}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (attribute_name, attribute_value, store_id)
		DO UPDATE SET
			attribute_name = '{{ .Name }}',
			attribute_value = '{{ .Value }}',
			store_id = {{ .StoreId }},
			date_updated = CURRENT_TIMESTAMP
		RETURNING id;`
	r := strings.NewReplacer(
		"{{ .Name }}", name,
		"{{ .Value }}", value,
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
	err = c.QueryRow(context.Background(), query).Scan(&resultId)
	if err != nil {
		log.Error().Err(err).Msg("QueryRow failed (UpsertAttributes)")
		return 0, err
	}
	log.Info().Msg(fmt.Sprintf("DB Insert Attributes ResultID: %d", resultId))
	return resultId, nil
}

func UpsertProductAndProductVariations(productId int, attrId int, varIdRaw string, sku string) (int, error){
	sku = strings.ReplaceAll(sku, "'", "''")
	c, err := Pool.Acquire(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Cannot acquire connection")
		return 0, err
	}
	defer c.Release()

	queryBaseItem := `
		INSERT INTO variations (attribute_id, variation_id_raw, sku, date_added, date_updated)
		VALUES($1, '{{ .VarIdRaw }}', '{{ .Sku }}', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (attribute_id, variation_id_raw, sku)
		DO UPDATE SET
			attribute_id = $1,
			variation_id_raw = '{{ .VarIdRaw }}',
			sku = '{{ .Sku }}',
			date_updated = CURRENT_TIMESTAMP
		RETURNING id;`
	r := strings.NewReplacer(
		"{{ .VarIdRaw }}", varIdRaw,
		"{{ .Sku }}", sku,
	)
	queryItem := r.Replace(queryBaseItem)
	var resultId int
	err = c.QueryRow(context.Background(), queryItem, attrId).Scan(&resultId)
	if err != nil {
		log.Error().Err(err).Msg("QueryRow failed (UpsertProductAndProductVariations)")
		return 0, err
	}
	log.Info().Msg(fmt.Sprintf("DB Insert variations ResultID: %d", resultId))

	queryBaseItemProduct := `
		INSERT INTO product_variations (product_id, variation_id, date_added, date_updated)
		VALUES({{ .ProductId }}, {{ .VarId }}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (product_id, variation_id)
		DO UPDATE SET
			product_id = {{ .ProductId }},
			variation_id = {{ .VarId }},
			date_updated = CURRENT_TIMESTAMP
		RETURNING id;`
	r = strings.NewReplacer(
		"{{ .ProductId }}", strconv.Itoa(productId),
		"{{ .VarId }}", strconv.Itoa(resultId),
	)
	queryItem = r.Replace(queryBaseItemProduct)
	//fmt.Println(queryItem)
	err = c.QueryRow(context.Background(), queryItem).Scan(&resultId)
	if err != nil {
		log.Error().Err(err).Msg("QueryRow Failed UpsertProductAndProductVariations (2)")
		os.Exit(0)
		return 0, err
	}
	log.Info().Msg(fmt.Sprintf("DB Insert product_variations ResultID: %d", resultId))
	return resultId, nil
}

func UpsertDatapoints(variationId int, stock int, price float32, ) (int, error) {
	queryFind := fmt.Sprintf(`
		SELECT id FROM datapoints 
		WHERE variation_id = %d AND stock = %d AND price = %.2f AND date_added = CURRENT_TIMESTAMP`,
		variationId, stock, price,
	)
	query := fmt.Sprintf(`
		INSERT INTO datapoints (variation_id, stock, price, date_added)
		VALUES(%d, %d, %.2f, CURRENT_TIMESTAMP) RETURNING id`,
		variationId, stock, price,
	)
	c, err := Pool.Acquire(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Cannot acquire connection")
		return 0, err
	}
	defer c.Release()
	var resultId int
	err = c.QueryRow(context.Background(), queryFind).Scan(&resultId)
	if err == nil {
		fmt.Println(resultId)
		log.Error().Err(err).Msg("QueryRow failed (UpsertDatapoints Find)")
		return 0, err
	}
	err = c.QueryRow(context.Background(), query).Scan(&resultId)
	if err != nil {
		log.Error().Err(err).Msg("QueryRow failed (UpsertDatapoints)")
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

