package database

import "database/sql"

func RemoveCompData(client *sql.DB, id string) (bool, error) {
	sqlStatement := `DELETE FROM order_buy WHERE group_id = $1`
	_, err := client.Exec(sqlStatement, id)
	if err != nil {
		return false, err
	}
	return true, nil
}

func RemoveSaleData(client *sql.DB, id string) (bool, error) {
	sqlStatement := `DELETE FROM order_sell WHERE group_id = $1`
	_, err := client.Exec(sqlStatement, id)
	if err != nil {
		return false, err
	}
	return true, nil
}

func RemoveEndTimeData(client *sql.DB, id string) (bool, error) {
	sqlStatement := `DELETE FROM end_time WHERE id = $1`
	_, err := client.Exec(sqlStatement, id)
	if err != nil {
		return false, err
	}
	return true, nil
}
