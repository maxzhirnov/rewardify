package store

func (r *Postgres) Bootstrap() error {
	// Создаем таблицу users
	createUserTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		uuid      TEXT PRIMARY KEY,
		username  TEXT UNIQUE,
		password  TEXT,
		created_at TIMESTAMP
	);`

	if _, err := r.DB.Exec(createUserTableSQL); err != nil {
		r.logger.Log.Error(err)
		return err
	}

	// SQL запрос для создания индекса на поле Username, если он не существует
	createIndexUserSQL := `CREATE UNIQUE INDEX IF NOT EXISTS index_username ON users (username);`

	if _, err := r.DB.Exec(createIndexUserSQL); err != nil {
		r.logger.Log.Error(err)
		return err
	}

	// Создаем таблицу orders
	createOrdersTableSQL := `
	CREATE TABLE IF NOT EXISTS orders (
	  order_number          TEXT PRIMARY KEY,
	  user_uuid             TEXT,
	  bonus_accrual_status  TEXT,
	  bonuses_accrued       REAL,
	  bonuses_spent         REAL,
	  created_at            TIMESTAMP,
	  FOREIGN KEY (user_uuid) REFERENCES users(uuid)
	);
`
	if _, err := r.DB.Exec(createOrdersTableSQL); err != nil {
		r.logger.Log.Error(err)
		return err
	}

	// SQL запрос для создания индекса на поле orders_number, если он не существует
	createIndexOrdersSQL := `CREATE UNIQUE INDEX IF NOT EXISTS index_order_number ON orders (order_number);`

	if _, err := r.DB.Exec(createIndexOrdersSQL); err != nil {
		r.logger.Log.Error(err)
		return err
	}

	r.logger.Log.Info("Bootstrap completed successfully")
	return nil
}
