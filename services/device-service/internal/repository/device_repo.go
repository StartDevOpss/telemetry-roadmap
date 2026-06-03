package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DeviceRepository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *DeviceRepository {
	return &DeviceRepository{db: db}
}

func (r *DeviceRepository) Migrate(ctx context.Context) error {
	_, err := r.db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS devices (
			id             TEXT PRIMARY KEY,
			last_lat       FLOAT8,
			last_lon       FLOAT8,
			last_battery   FLOAT8,
			last_temp_c    FLOAT8,
			last_speed_kmh FLOAT8,
			updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE TABLE IF NOT EXISTS processed_events (
			event_id     UUID PRIMARY KEY,
			processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	return err
}

// MarkProcessed insere o event_id. Retorna true se já existia (duplicata).
func (r *DeviceRepository) MarkProcessed(ctx context.Context, eventID string) (duplicate bool, err error) {
	tag, err := r.db.Exec(ctx,
		`INSERT INTO processed_events (event_id) VALUES ($1) ON CONFLICT DO NOTHING`,
		eventID,
	)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 0, nil
}

func (r *DeviceRepository) UpsertDevice(ctx context.Context, id string, lat, lon, battery, tempC, speedKmh float64) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO devices (id, last_lat, last_lon, last_battery, last_temp_c, last_speed_kmh, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (id) DO UPDATE SET
			last_lat       = EXCLUDED.last_lat,
			last_lon       = EXCLUDED.last_lon,
			last_battery   = EXCLUDED.last_battery,
			last_temp_c    = EXCLUDED.last_temp_c,
			last_speed_kmh = EXCLUDED.last_speed_kmh,
			updated_at     = EXCLUDED.updated_at
	`, id, lat, lon, battery, tempC, speedKmh)
	return err
}
