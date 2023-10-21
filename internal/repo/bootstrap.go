package repo

func (p *Postgres) Bootstrap() error {
	// Создаем таблицу users
	createUserTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		uuid      UUID PRIMARY KEY,
		username  TEXT UNIQUE,
		password  TEXT,
		created_at TIMESTAMP
	);`

	if _, err := p.db.Exec(createUserTableSQL); err != nil {
		p.logger.Log.Error(err)
		return err
	}

	// SQL запрос для создания индекса на поле Username, если он не существует
	createIndexUserSQL := `CREATE UNIQUE INDEX IF NOT EXISTS index_username ON users (username);`

	if _, err := p.db.Exec(createIndexUserSQL); err != nil {
		p.logger.Log.Error(err)
		return err
	}

	// Создаем таблицу orders
	createOrdersTableSQL := `
	CREATE TABLE IF NOT EXISTS orders (
	  order_number          VARCHAR(16) PRIMARY KEY,
	  user_uuid             UUID,
	  bonus_accrual_status  TEXT,
	  created_at            TIMESTAMP,
	  FOREIGN KEY (user_uuid) REFERENCES users(uuid)
	);
`
	if _, err := p.db.Exec(createOrdersTableSQL); err != nil {
		p.logger.Log.Error(err)
		return err
	}

	// Создаем таблицу с начислениями по orders
	createAccrualsTableSQL := `
CREATE TABLE IF NOT EXISTS accruals_calculated (
	id SERIAL PRIMARY KEY,
	user_uuid UUID,
	order_number VARCHAR(16),
	accrued REAL,
	FOREIGN KEY (order_number) REFERENCES orders(order_number),
	FOREIGN KEY (user_uuid) REFERENCES users(uuid));
`
	if _, err := p.db.Exec(createAccrualsTableSQL); err != nil {
		p.logger.Log.Error(err)
		return err
	}

	// Создаем таблицу с балансами пользователей
	createBalanceTableSQL := `
CREATE TABLE IF NOT EXISTS balances (
	id SERIAL PRIMARY KEY,
	user_uuid UUID,
	total_bonus REAL NOT NULL,
	redeemed_bonus REAL NOT NULL,
	FOREIGN KEY (user_uuid) REFERENCES users(uuid)
);
`
	if _, err := p.db.Exec(createBalanceTableSQL); err != nil {
		p.logger.Log.Error(err)
		return err
	}

	// Создаем таблицу с withdrawals пользователей
	createWithdrawalsTableSQL := `
CREATE TABLE IF NOT EXISTS withdrawals (
	id SERIAL PRIMARY KEY,
	user_uuid UUID,
	order_number VARCHAR(16),
	withdrew REAL,
	created_at TIMESTAMP,
	FOREIGN KEY (user_uuid) REFERENCES users(uuid)
);
`
	if _, err := p.db.Exec(createWithdrawalsTableSQL); err != nil {
		p.logger.Log.Error(err)
		return err
	}

	// SQL запрос для создания индекса на поле orders_number, если он не существует
	createIndexOrdersSQL := `CREATE UNIQUE INDEX IF NOT EXISTS index_order_number ON orders (order_number);`

	if _, err := p.db.Exec(createIndexOrdersSQL); err != nil {
		p.logger.Log.Error(err)
		return err
	}

	p.logger.Log.Info("Bootstrap completed successfully")
	return nil
}
