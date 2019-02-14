package tableengines

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"

	"github.com/jackc/pgx"

	"github.com/ikitiki/pg2ch/pkg/config"
	"github.com/ikitiki/pg2ch/pkg/message"
	"github.com/ikitiki/pg2ch/pkg/utils"
)

type CollapsingMergeTreeTable struct {
	genericTable

	signColumn string
}

func NewCollapsingMergeTree(conn *sql.DB, name string, tblCfg config.Table) *CollapsingMergeTreeTable {
	t := CollapsingMergeTreeTable{
		genericTable: newGenericTable(conn, name, tblCfg),
		signColumn:   tblCfg.SignColumn,
	}
	t.chColumns = append(t.chColumns, tblCfg.SignColumn)

	return &t
}

func (t *CollapsingMergeTreeTable) Sync(pgTx *pgx.Tx) error {
	return t.genSync(pgTx, t)
}

func (t *CollapsingMergeTreeTable) Write(p []byte) (n int, err error) {
	rec, err := csv.NewReader(bytes.NewReader(p)).Read()
	if err != nil {
		return 0, err
	}

	if err := t.stmntExec(append(t.convertStrings(rec), 1)); err != nil {
		return 0, fmt.Errorf("could not insert: %v", err)
	}

	return len(p), nil
}

func (t *CollapsingMergeTreeTable) Insert(lsn utils.LSN, new message.Row) error {
	return t.bufferCmdSet(commandSet{
		append(t.convertTuples(new), 1),
	})
}

func (t *CollapsingMergeTreeTable) Update(lsn utils.LSN, old, new message.Row) error {
	return t.bufferCmdSet(commandSet{
		append(t.convertTuples(old), -1),
		append(t.convertTuples(new), 1),
	})
}

func (t *CollapsingMergeTreeTable) Delete(lsn utils.LSN, old message.Row) error {
	return t.bufferCmdSet(commandSet{
		append(t.convertTuples(old), -1),
	})
}
