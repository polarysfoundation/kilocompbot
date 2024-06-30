package database

import (
	"database/sql"
	"math/big"
)

func WriteGroups(db *sql.DB, id string, compActive bool, jettonAddress string, dedust string, stonfi string, emoji string) error {
	sqlStatement := "INSERT INTO groups (id, comp_active, jetton_address, dedust_address, stonfi_address, emoji) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (id) DO UPDATE SET comp_active = EXCLUDED.comp_active, jetton_address = EXCLUDED.jetton_address, dedust_address = EXCLUDED.dedust_address, stonfi_address = EXCLUDED.stonfi_address, emoji = EXCLUDED.emoji"
	_, err := db.Exec(sqlStatement, id, compActive, jettonAddress, dedust, stonfi, emoji)
	if err != nil {
		return err
	}

	return nil
}

func WritePurchases(db *sql.DB, id string, jettonAddress string, jettonName string, jettonSymbol string, jettonDecimals *big.Int, buyer string, ton *big.Int, token *big.Int) error {
	sqlStatement := "INSERT INTO order_buy (group_id, jetton_address, jetton_name, jetton_symbol, jetton_decimal, buyer_address, ton_amount, token_amount) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)"
	_, err := db.Exec(sqlStatement, id, jettonAddress, jettonName, jettonSymbol, jettonDecimals.Int64(), buyer, ton.Int64(), token.Int64())
	if err != nil {
		return err
	}

	return nil
}

func WriteSales(db *sql.DB, id string, jettonAddress string, jettonName string, jettonSymbol string, jettonDecimals *big.Int, seller string, ton *big.Int, token *big.Int) error {
	sqlStatement := "INSERT INTO order_sell (group_id, jetton_address, jetton_name, jetton_symbol, jetton_decimal, seller_address, ton_amount, token_amount) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)"
	_, err := db.Exec(sqlStatement, id, jettonAddress, jettonName, jettonSymbol, jettonDecimals.Int64(), seller, ton.Int64(), token.Int64())
	if err != nil {
		return err
	}
	return nil
}

func WritePromo(db *sql.DB, id string, adName string, buttonName string, buttonLink string, media string) error {
	sqlStatement := "INSERT INTO promo (id, ad_text, button_name, button_link, media) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (id) DO UPDATE SET ad_text = EXCLUDED.ad_text, button_name = EXCLUDED.button_name, button_link = EXCLUDED.button_link, media = EXCLUDED.media"
	_, err := db.Exec(sqlStatement, id, adName, buttonName, buttonLink, media)
	if err != nil {
		return err
	}
	return nil
}
