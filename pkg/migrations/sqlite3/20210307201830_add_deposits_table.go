package sqlite3

import (
	"context"

	"github.com/c9s/rockhopper"
)

func init() {
	AddMigration(upAddDepositsTable, downAddDepositsTable)

}

func upAddDepositsTable(ctx context.Context, tx rockhopper.SQLExecutor) (err error) {
	// This code is executed when the migration is applied.

	_, err = tx.ExecContext(ctx, "CREATE TABLE `deposits`\n(\n    `gid`      INTEGER PRIMARY KEY AUTOINCREMENT,\n    `exchange` VARCHAR(24)    NOT NULL,\n    -- asset is the asset name (currency)\n    `asset`    VARCHAR(10)    NOT NULL,\n    `address`  VARCHAR(64)    NOT NULL DEFAULT '',\n    `amount`   DECIMAL(16, 8) NOT NULL,\n    `txn_id`   VARCHAR(64)    NOT NULL,\n    `time`     DATETIME(3)    NOT NULL\n);")
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "CREATE UNIQUE INDEX `deposits_txn_id` ON `deposits` (`exchange`, `txn_id`);")
	if err != nil {
		return err
	}

	return err
}

func downAddDepositsTable(ctx context.Context, tx rockhopper.SQLExecutor) (err error) {
	// This code is executed when the migration is rolled back.

	_, err = tx.ExecContext(ctx, "DROP INDEX IF EXISTS `deposits_txn_id`;")
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "DROP TABLE IF EXISTS `deposits`;")
	if err != nil {
		return err
	}

	return err
}
