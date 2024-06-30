package database

import (
	"database/sql"
	"fmt"
	"math/big"

	"github.com/polarysfoundation/kilocompbot/bot/promotions"
	"github.com/polarysfoundation/kilocompbot/core"
	"github.com/polarysfoundation/kilocompbot/groups"
)

func GetGroups(db *sql.DB) ([]*groups.GroupData, error) {
	row, err := db.Query(`SELECT * FROM groups`)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	var groups_data []*groups.GroupData

	for row.Next() {
		var group groups.GroupData

		err := row.Scan(
			&group.ID,
			&group.CompActive,
			&group.JettonAddress,
			&group.Dedust,
			&group.StonFi,
			&group.Emoji,
		)
		if err != nil {
			return nil, err
		}

		groups_data = append(groups_data, &group)
	}

	if err := row.Err(); err != nil {
		return nil, err
	}

	return groups_data, nil
}

func GetPurchase(db *sql.DB, id string) ([]*core.Purchase, error) {
	rows, err := db.Query(`
	SELECT jetton_address, jetton_name, jetton_symbol, jetton_decimal,
		   buyer_address, ton_amount, token_amount
	FROM order_buy 
	WHERE group_id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var purchases []*core.Purchase

	for rows.Next() {
		var purchase core.Purchase
		var tonAmount, tokenAmount, jettonDecimals string

		err := rows.Scan(
			&purchase.JettonAddress,
			&purchase.JettonName,
			&purchase.JettonSymbol,
			&jettonDecimals,
			&purchase.Buyer,
			&tonAmount,
			&tokenAmount,
		)
		if err != nil {
			return nil, err
		}

		// Convertir strings a *big.Int
		purchase.JettonDecimals, _ = new(big.Int).SetString(jettonDecimals, 10)
		purchase.Ton, _ = new(big.Int).SetString(tonAmount, 10)
		purchase.Token, _ = new(big.Int).SetString(tokenAmount, 10)

		purchases = append(purchases, &purchase)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return purchases, nil
}

func GetSale(db *sql.DB, id string) ([]*core.Sale, error) {
	rows, err := db.Query(`
        SELECT jetton_address, jetton_name, jetton_symbol, jetton_decimal, 
               seller_address, ton_amount, token_amount 
        FROM order_sell 
        WHERE group_id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sales []*core.Sale

	for rows.Next() {
		var sale core.Sale
		var tonAmount, tokenAmount, jettonDecimals string

		err := rows.Scan(
			&sale.JettonAddress,
			&sale.JettonName,
			&sale.JettonSymbol,
			&jettonDecimals,
			&sale.Seller,
			&tonAmount,
			&tokenAmount,
		)
		if err != nil {
			return nil, err
		}

		// Convertir strings a *big.Int
		sale.JettonDecimals, _ = new(big.Int).SetString(jettonDecimals, 10)
		sale.Ton, _ = new(big.Int).SetString(tonAmount, 10)
		sale.Token, _ = new(big.Int).SetString(tokenAmount, 10)

		sales = append(sales, &sale)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sales, nil
}

func GetPromos(db *sql.DB, id string) (*promotions.Params, error) {
	row := db.QueryRow(`SELECT ad_text, button_name, button_link, media FROM promo WHERE id = $1`, id)

	promo := &promotions.Params{} // Inicializa promo aquí

	err := row.Scan(
		&promo.AdName,
		&promo.ButtonName,
		&promo.ButtonLink,
		&promo.Media,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no promo found for id: %s", id)
		}
		return nil, err
	}

	return promo, nil
}

func GetEndTime(db *sql.DB, id string) (int64, error) {
	rows := db.QueryRow(`SELECT timestamp FROM end_time WHERE id = $1`, id)

	var timestamp int64

	err := rows.Scan(
		&timestamp,
	)
	if err != nil {
		return 0, err
	}

	if err := rows.Err(); err != nil {
		return 0, err
	}

	return timestamp, nil
}
