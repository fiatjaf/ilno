package database

import (
	"context"
	"time"

	"github.com/fiatjaf/ilno/ilno"
)

func (d *Database) IsBannedUser(ctx context.Context, key string) bool {
	ctx, cancel := d.withTimeout(ctx)
	defer cancel()

	var flag int64
	err := d.DB.QueryRowContext(ctx, d.statement["is_banned_user"], key).Scan(&flag)
	return (err == nil) && (flag == 1)
}

func (d *Database) ListBannedUsers(ctx context.Context) ([]ilno.BannedUser, error) {
	ctx, cancel := d.withTimeout(ctx)
	defer cancel()

	var banList []ilno.BannedUser

	rows, err := d.DB.QueryContext(ctx, d.statement["ban_list"])
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var key string
		var bannedAt int64
		err := rows.Scan(&key, &bannedAt)
		if err != nil {
			return nil, wraperror(err)
		}
		banList = append(banList, ilno.BannedUser{key, bannedAt})
	}
	if rows.Err() != nil {
		return nil, wraperror(err)
	}
	return banList, nil
}

func (d *Database) BanUser(ctx context.Context, key string) error {
	ctx, cancel := d.withTimeout(ctx)
	defer cancel()
	_, err := d.DB.ExecContext(ctx, d.statement["ban_user"], key, time.Now().Unix())
	return err
}

func (d *Database) UnbanUser(ctx context.Context, key string) error {
	ctx, cancel := d.withTimeout(ctx)
	defer cancel()
	_, err := d.DB.ExecContext(ctx, d.statement["unban_user"], key)
	return err
}
