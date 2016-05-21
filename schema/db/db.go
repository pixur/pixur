package db

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

var (
	mysqlAdapter = DbAdapter{
		Quote: func(ident string) string {
			if strings.ContainsAny(ident, "\"\x00`") {
				panic(fmt.Sprintf("Invalid identifier %#v", ident))
			}
			return "`" + ident + "`"
		},
		QuoteCreateBlobIdxCol: func(ident string) string {
			if strings.ContainsAny(ident, "\"\x00`") {
				panic(fmt.Sprintf("Invalid identifier %#v", ident))
			}
			return "`" + ident + "`(255)"
		},
		LockStmt: func(buf *bytes.Buffer, lock Lock) {
			switch lock {
			case LockNone:
			case LockRead:
				buf.WriteString(" LOCK IN SHARE MODE")
			case LockWrite:
				buf.WriteString(" FOR UPDATE")
			default:
				panic(fmt.Errorf("Unknown lock %v", lock))
			}
		},
		BoolType:   "bool",
		IntType:    "int",
		BigIntType: "bigint(20)",
		BlobType:   "blob",
	}

	postgresqlAdapter = DbAdapter{
		Quote: func(ident string) string {
			if strings.ContainsAny(ident, "\"\x00") {
				panic(fmt.Sprintf("Invalid identifier %#v", ident))
			}
			return `"` + ident + `"`
		},
		QuoteCreateBlobIdxCol: func(ident string) string {
			if strings.ContainsAny(ident, "\"\x00") {
				panic(fmt.Sprintf("Invalid identifier %#v", ident))
			}
			return `"` + ident + `"`
		},
		LockStmt: func(buf *bytes.Buffer, lock Lock) {
			switch lock {
			case LockNone:
			case LockRead:
				buf.WriteString(" FOR SHARE")
			case LockWrite:
				buf.WriteString(" FOR UPDATE")
			default:
				panic(fmt.Errorf("Unknown lock %v", lock))
			}
		},
		BoolType:   "bool",
		IntType:    "integer",
		BigIntType: "bigint",
		BlobType:   "bytea",
	}
)

// default
var dbAdapter DbAdapter = mysqlAdapter

type DbAdapter struct {
	Quote, QuoteCreateBlobIdxCol            func(string) string
	LockStmt                                func(*bytes.Buffer, Lock)
	BoolType, IntType, BigIntType, BlobType string
}

type Lock int

var (
	LockNone  Lock = -1
	LockRead  Lock = 0
	LockWrite Lock = 1
)

type Opts struct {
	Prefix      Idx
	Start, Stop Idx
	Lock        Lock
	Reverse     bool
	Limit       int
}

type Idx interface {
	Cols() []string
	Vals() []interface{}
}

// UniqueIdx is a tagging interface that indentifies indexes that uniquely identify a row.
// Columns that are UNIQUE or PRIMARY fit this interface.
type UniqueIdx interface {
	Idx
	Unique()
}

type Querier interface {
	Query(query string, args ...interface{}) (Rows, error)
}

type Executor interface {
	Exec(string, ...interface{}) (Result, error)
}

// Result is a clone of database/sql.Result
type Result interface {
	LastInsertId() (int64, error)
	RowsAffected() (int64, error)
}

// Rows is a clone of database/sql.Rows
type Rows interface {
	Close() error
	Columns() ([]string, error)
	Err() error
	Next() bool
	Scan(dest ...interface{}) error
}

type scanStmt struct {
	opts Opts
	name string
	buf  *bytes.Buffer
	args []interface{}
}

// Scan scans a table for matching rows.
func Scan(q Querier, name string, opts Opts, cb func(data []byte) error) (errCap error) {
	s := scanStmt{
		opts: opts,
		name: name,
		buf:  new(bytes.Buffer),
	}

	query, queryArgs := s.buildScan()
	rows, err := q.Query(query, queryArgs...)
	if err != nil {
		return err
	}
	defer func() {
		if newErr := rows.Close(); errCap == nil {
			errCap = newErr
		}
	}()

	for rows.Next() {
		var tmp []byte
		if err := rows.Scan(&tmp); err != nil {
			return err
		}
		if err := cb(tmp); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}

func (s *scanStmt) buildScan() (string, []interface{}) {
	fmt.Fprintf(s.buf, "SELECT %s FROM %s", quoteIdentifier("data"), quoteIdentifier(s.name))

	if s.opts.Prefix != nil && (s.opts.Start != nil || s.opts.Stop != nil) {
		panic("only Prefix or Start|Stop may be specified")
	}
	if s.opts.Prefix != nil {
		s.appendPrefix()
	} else if s.opts.Start != nil || s.opts.Stop != nil {
		s.appendRange()
	}

	if s.opts.Limit > 0 {
		fmt.Fprintf(s.buf, " LIMIT %d", s.opts.Limit)
	}
	s.appendLock()
	s.buf.WriteRune(';')
	return s.buf.String(), s.args
}

func (s *scanStmt) appendPrefix() {
	cols, vals := s.opts.Prefix.Cols(), s.opts.Prefix.Vals()
	if len(vals) != 0 {
		s.buf.WriteString(" WHERE ")
		for i := 0; i < len(vals); i++ {
			if i != 0 {
				s.buf.WriteString(" AND ")
			}
			fmt.Fprintf(s.buf, "%s = ?", quoteIdentifier(cols[i]))
			s.args = append(s.args, vals[i])
		}
	}
	if sortCols := cols[len(vals):]; len(sortCols) != 0 {
		s.appendOrder(sortCols)
	}

}

func (s *scanStmt) appendRange() {
	var (
		startCols, stopCols []string
		startVals, stopVals []interface{}
	)
	if s.opts.Start != nil {
		startCols, startVals = s.opts.Start.Cols(), s.opts.Start.Vals()
	}
	if s.opts.Stop != nil {
		stopCols, stopVals = s.opts.Stop.Cols(), s.opts.Stop.Vals()
	}
	s.buf.WriteString(" WHERE ")
	if len(startVals) != 0 {
		startStmt, startArgs := buildStart(startCols, startVals)
		s.args = append(s.args, startArgs...)
		s.buf.WriteString(startStmt)
	}
	if len(startVals) != 0 && len(stopVals) != 0 {
		s.buf.WriteString(" AND ")
	}
	if len(stopVals) != 0 {
		stopStmt, stopArgs := buildStop(stopCols, stopVals)
		s.args = append(s.args, stopArgs...)
		s.buf.WriteString(stopStmt)
	}
	if len(startVals) != 0 {
		s.appendOrder(startCols)
	} else {
		s.appendOrder(stopCols)
	}
}

func (s *scanStmt) appendOrder(cols []string) {
	s.buf.WriteString(" ORDER BY ")

	var order string
	if !s.opts.Reverse {
		order = " ASC"
	} else {
		order = " DESC"
	}
	for i, col := range cols {
		if i != 0 {
			s.buf.WriteString(", ")
		}
		s.buf.WriteString(quoteIdentifier(col))
		s.buf.WriteString(order)
	}
}

func (s *scanStmt) appendLock() {
	dbAdapter.LockStmt(s.buf, s.opts.Lock)
}

type Columns []string

func (cols Columns) String() string {
	var parts []string
	for _, col := range cols {
		parts = append(parts, quoteIdentifier(col))
	}
	return strings.Join(parts, ", ")
}

func buildStart(cols []string, vals []interface{}) (string, []interface{}) {
	if len(vals) > len(cols) {
		panic("More vals than cols")
	}
	var args []interface{}
	// Disjunctive normal form, you nerd!
	// Start always has the last argument be a ">="
	// 1, 2, 3 arg scans look like:
	// ((A >= ?))
	// ((A > ?) OR (A = ? AND B >= ?))
	// ((A > ?) OR (A = ? AND B > ?) OR (A = ? AND B = ? AND C >= ?))
	var ors []string
	for i := 0; i < len(vals); i++ {
		var ands []string
		for k := 0; k < i; k++ {
			ands = append(ands, quoteIdentifier(cols[k])+" = ?")
			args = append(args, vals[k])
		}
		if i == len(vals)-1 {
			ands = append(ands, quoteIdentifier(cols[i])+" >= ?")
		} else {
			ands = append(ands, quoteIdentifier(cols[i])+" > ?")
		}
		args = append(args, vals[i])
		ors = append(ors, "("+strings.Join(ands, " AND ")+")")
	}
	return "(" + strings.Join(ors, " OR ") + ")", args
}

func buildStop(cols []string, vals []interface{}) (string, []interface{}) {
	if len(vals) > len(cols) {
		panic("More vals than cols")
	}
	var args []interface{}
	// Stop always has the last argument be a "<"
	// 1, 2, 3 arg scans look like:
	// ((A < ?))
	// ((A < ?) OR (A = ? AND B < ?))
	// ((A < ?) OR (A = ? AND B < ?) OR (A = ? AND B = ? AND C < ?))
	var ors []string
	for i := 0; i < len(vals); i++ {
		var ands []string
		for k := 0; k < i; k++ {
			ands = append(ands, quoteIdentifier(cols[k])+" = ?")
			args = append(args, vals[k])
		}
		ands = append(ands, quoteIdentifier(cols[i])+" < ?")
		args = append(args, vals[i])
		ors = append(ors, "("+strings.Join(ands, " AND ")+")")
	}
	return "(" + strings.Join(ors, " OR ") + ")", args
}

var (
	ErrColsValsMismatch = errors.New("db: number of columns and values don't match.")
	ErrNoCols           = errors.New("db: no columns provided")
)

func Insert(exec Executor, name string, cols []string, vals []interface{}) error {
	if len(cols) != len(vals) {
		return ErrColsValsMismatch
	}
	if len(cols) == 0 {
		return ErrNoCols
	}

	valFmt := strings.Repeat("?, ", len(vals)-1) + "?"
	colFmtParts := make([]string, 0, len(cols))
	for _, col := range cols {
		colFmtParts = append(colFmtParts, quoteIdentifier(col))
	}
	colFmt := strings.Join(colFmtParts, ", ")
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", quoteIdentifier(name), colFmt, valFmt)
	_, err := exec.Exec(query, vals...)
	return err
}

func Delete(exec Executor, name string, key UniqueIdx) error {
	cols := key.Cols()
	vals := key.Vals()
	if len(cols) != len(vals) {
		return ErrColsValsMismatch
	}
	if len(cols) == 0 {
		return ErrNoCols
	}

	colFmtParts := make([]string, 0, len(cols))
	for _, col := range cols {
		colFmtParts = append(colFmtParts, quoteIdentifier(col)+" = ?")
	}
	colFmt := strings.Join(colFmtParts, " AND ")
	query := fmt.Sprintf("DELETE FROM %s WHERE %s LIMIT 1;", quoteIdentifier(name), colFmt)
	_, err := exec.Exec(query, vals...)
	return err
}

func Update(exec Executor, name string, cols []string, vals []interface{}, key UniqueIdx) error {
	if len(cols) != len(vals) {
		return ErrColsValsMismatch
	}
	if len(cols) == 0 {
		return ErrNoCols
	}

	idxCols := key.Cols()
	idxVals := key.Vals()
	if len(idxCols) != len(idxVals) {
		return ErrColsValsMismatch
	}
	if len(idxCols) == 0 {
		return ErrNoCols
	}

	colFmtParts := make([]string, 0, len(cols))
	for _, col := range cols {
		colFmtParts = append(colFmtParts, quoteIdentifier(col)+" = ?")
	}
	colFmt := strings.Join(colFmtParts, ", ")

	idxColFmtParts := make([]string, 0, len(idxCols))
	for _, idxCol := range idxCols {
		idxColFmtParts = append(idxColFmtParts, quoteIdentifier(idxCol)+" = ?")
	}
	idxColFmt := strings.Join(idxColFmtParts, " AND ")

	allVals := make([]interface{}, 0, len(vals)+len(idxVals))
	allVals = append(allVals, vals...)
	allVals = append(allVals, idxVals...)

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s LIMIT 1;", quoteIdentifier(name), colFmt, idxColFmt)
	_, err := exec.Exec(query, allVals...)
	return err
}

// quoteIdentifier quotes the ANSI way.  Panics on invalid identifiers.
func quoteIdentifier(ident string) string {
	return dbAdapter.Quote(ident)
}

// Don't use this.  Seriously.
func GetAdapter() DbAdapter {
	return dbAdapter
}

func setPostgreSQLForTest() {
	dbAdapter = postgresqlAdapter
}
