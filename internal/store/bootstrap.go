package store

func (p *Postgres) Bootstrap() error {
	// Создаем таблицу users
	createUserTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		uuid      TEXT PRIMARY KEY,
		username  TEXT UNIQUE,
		password  TEXT,
		created_at TIMESTAMP
	);`

	if _, err := p.DB.Exec(createUserTableSQL); err != nil {
		p.logger.Log.Error(err)
		return err
	}

	// SQL запрос для создания индекса на поле Username, если он не существует
	createIndexUserSQL := `CREATE UNIQUE INDEX IF NOT EXISTS index_username ON users (username);`

	if _, err := p.DB.Exec(createIndexUserSQL); err != nil {
		p.logger.Log.Error(err)
		return err
	}

	// Создаем таблицу orders
	createOrdersTableSQL := `
	CREATE TABLE IF NOT EXISTS orders (
	  order_number          TEXT PRIMARY KEY,
	  user_uuid             TEXT,
	  bonus_accrual_status  TEXT,
	  created_at            TIMESTAMP,
	  FOREIGN KEY (user_uuid) REFERENCES users(uuid)
	);
`
	if _, err := p.DB.Exec(createOrdersTableSQL); err != nil {
		p.logger.Log.Error(err)
		return err
	}

	// Создаем таблицу с начислениями по orders
	createAccrualsTableSQL := `
CREATE TABLE IF NOT EXISTS accruals_calculated (
	id SERIAL PRIMARY KEY,
	user_uuid TEXT,
	order_number TEXT,
	accrued REAL,
	FOREIGN KEY (order_number) REFERENCES orders(order_number),
	FOREIGN KEY (user_uuid) REFERENCES users(uuid));
`
	if _, err := p.DB.Exec(createAccrualsTableSQL); err != nil {
		p.logger.Log.Error(err)
		return err
	}

	// Создаем таблицу с балансами пользователей
	createBalanceTableSQL := `
CREATE TABLE IF NOT EXISTS balances (
	id SERIAL PRIMARY KEY,
	user_uuid TEXT,
	total_bonus REAL NOT NULL,
	redeemed_bonus REAL NOT NULL,
	FOREIGN KEY (user_uuid) REFERENCES users(uuid)
);
`
	if _, err := p.DB.Exec(createBalanceTableSQL); err != nil {
		p.logger.Log.Error(err)
		return err
	}

	// Создаем таблицу с withdrawals пользователей
	createWithdrawalsTableSQL := `
CREATE TABLE IF NOT EXISTS withdrawals (
	id SERIAL PRIMARY KEY,
	user_uuid TEXT,
	order_number TEXT,
	withdrew REAL,
	created_at TIMESTAMP,
	FOREIGN KEY (user_uuid) REFERENCES users(uuid)
);
`
	if _, err := p.DB.Exec(createWithdrawalsTableSQL); err != nil {
		p.logger.Log.Error(err)
		return err
	}

	// Триггеры для обновления баланса бонусов
	triggerSQL := `
CREATE OR REPLACE FUNCTION update_bonus_balance()
RETURNS TRIGGER AS $$
BEGIN
    -- Обновляем баланс бонусов для пользователя
    UPDATE balances
    SET total_bonus = total_bonus + NEW.accrued
    WHERE user_uuid = NEW.user_uuid;

    -- Если пользователя нет в таблице user_bonus, создаем новую запись
    IF NOT FOUND THEN
        INSERT INTO balances(user_uuid, total_bonus, redeemed_bonus)
        VALUES (NEW.user_uuid, NEW.accrued, 0);
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_bonus') THEN
		CREATE TRIGGER update_bonus
		AFTER INSERT ON accruals_calculated
		FOR EACH ROW EXECUTE FUNCTION update_bonus_balance();
	END IF;
END $$;
`
	if _, err := p.DB.Exec(triggerSQL); err != nil {
		p.logger.Log.Error(err)
		return err
	}

	// Триггер для списания
	triggerWithdrawSQL := `
CREATE OR REPLACE FUNCTION update_bonus_balance_withdraw()
RETURNS TRIGGER AS $$
BEGIN
    -- Обновляем баланс бонусов для пользователя
    UPDATE balances
    SET redeemed_bonus = balances.redeemed_bonus + NEW.withdrew
    WHERE user_uuid = NEW.user_uuid;

    -- Если пользователя нет в таблице user_bonus, создаем новую запись
    IF NOT FOUND THEN
        INSERT INTO balances(user_uuid, total_bonus, redeemed_bonus)
        VALUES (NEW.user_uuid, 0, NEW.withdrew);
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_bonus_withdraw') THEN
		CREATE TRIGGER update_bonus_withdraw
		AFTER INSERT ON withdrawals
		FOR EACH ROW EXECUTE FUNCTION update_bonus_balance_withdraw();
	END IF;
END $$;
`
	if _, err := p.DB.Exec(triggerWithdrawSQL); err != nil {
		p.logger.Log.Error(err)
		return err
	}

	// SQL запрос для создания индекса на поле orders_number, если он не существует
	createIndexOrdersSQL := `CREATE UNIQUE INDEX IF NOT EXISTS index_order_number ON orders (order_number);`

	if _, err := p.DB.Exec(createIndexOrdersSQL); err != nil {
		p.logger.Log.Error(err)
		return err
	}

	p.logger.Log.Info("Bootstrap completed successfully")
	return nil
}
