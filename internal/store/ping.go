package store

func (p *Postgres) Ping() error {
	return p.DB.Ping()
}
