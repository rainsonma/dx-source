package migrations

type M20260407000001CreateGameJunctionTables struct{}

func (r *M20260407000001CreateGameJunctionTables) Signature() string {
	return "20260407000001_create_game_junction_tables"
}

func (r *M20260407000001CreateGameJunctionTables) Up() error {
	return nil
}

func (r *M20260407000001CreateGameJunctionTables) Down() error {
	return nil
}
