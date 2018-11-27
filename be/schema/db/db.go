// Package db includes database functionality for Pixur.
package db // import "pixur.org/pixur/be/schema/db"

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"pixur.org/pixur/be/status"
)

// Commiter is an interface to commit transactions.  It is not threadsafe
type Commiter interface {
	// Commit returns nil iff the transaction was successful.  This method may only be called once.
	// If this method returns true, future calls to Rollback() will be no-ops and always return nil.
	// If an error is returned, the transaction may have been successful, or it may have failed.
	Commit() error
	// Rollback reverts the transaction.  Invocations after first are no-ops and always return nil.
	// If an error is returned, the caller does not need to call Rollback() again.
	Rollback() error
}

// QuerierExecutorCommitter is an interface that matches the sql.Tx type.
type QuerierExecutorCommitter interface {
	Querier
	Commiter
	Executor
}

// Beginner begins transactions. Upon a successful call to Being(), callers should take care
// to either Commit() or Rollback() the QuerierExecutorCommitter.
type Beginner interface {
	// Begin begins a transaction
	Begin(context.Context) (QuerierExecutorCommitter, error)
}

// DB represents a database.  It is the entry point to creating transactions.
type DB interface {
	DBAdaptable
	IDAllocatable
	Beginner
	// Close closes the database, and should be called when the db is no longer needed.
	Close() error
	// InitSchema initializes the database tables.  This is typically used in unit tests.
	InitSchema(context.Context, []string) error
}

type DBAdaptable interface {
	// Adapter returns an adapter for use with a database.
	Adapter() DBAdapter
}

type IDAllocatable interface {
	// IDAllocator returns an IDAlloc for use with a database.
	IDAllocator() *IDAlloc
}

// Retryable is implemented by errors that can describe their retryability.
type Retryable interface {
	CanRetry() bool
}

// Open is the main entry point to the db package.  adapterName is one of the registered types.
// dataSourceName is the connection information to initiate the db.
func Open(ctx context.Context, adapterName, dataSourceName string) (DB, error) {
	adapter, present := adapters[adapterName]
	if !present {
		return nil, status.InvalidArgument(nil, "no adapter", adapterName)
	}
	return adapter.Open(ctx, dataSourceName)
}

func OpenForTest(ctx context.Context, adapterName string) (DB, error) {
	adapter, present := adapters[adapterName]
	if !present {
		return nil, status.InvalidArgument(nil, "no adapter", adapterName)
	}
	return adapter.OpenForTest(ctx)
}

var adapters = make(map[string]DBAdapter)

func RegisterAdapter(a DBAdapter) {
	name := a.Name()
	if _, present := adapters[name]; present {
		panic(name + " already present")
	}
	adapters[name] = a
}

func GetAllAdapters() []DBAdapter {
	var all []DBAdapter
	var names []string
	for name, _ := range adapters {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		all = append(all, adapters[name])
	}
	return all
}

type DBAdapter interface {
	Name() string
	Quote(string) string
	BlobIdxQuote(string) string
	LockStmt(*strings.Builder, Lock)
	BoolType() string
	IntType() string
	BigIntType() string
	BlobType() string
	// Is this database inherently serial?
	SingleTx() bool

	Open(ctx context.Context, dataSourceName string) (DB, error)
	OpenForTest(context.Context) (DB, error)

	// Can the given error be retried?
	RetryableErr(error) bool
}

type Lock int

const (
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
	buf  *strings.Builder
	args []interface{}
	adap DBAdapter
}

// Scan scans a table for matching rows.
func Scan(q Querier, name string, opts Opts, cb func(data []byte) error, adap DBAdapter) error {
	return localScan(q, name, opts, cb, adap)
}

func localScan(
	q Querier, name string, opts Opts, cb func(data []byte) error, adap DBAdapter) (stscap status.S) {
	s := scanStmt{
		opts: opts,
		name: name,
		buf:  new(strings.Builder),
		adap: adap,
	}

	query, queryArgs := s.buildScan()
	rows, err := q.Query(query, queryArgs...)
	if err != nil {
		return status.From(err)
	}
	defer func() {
		if newErr := rows.Close(); newErr != nil {
			status.ReplaceOrSuppress(&stscap, status.From(newErr))
		}
	}()

	for rows.Next() {
		var tmp []byte
		if err := rows.Scan(&tmp); err != nil {
			return status.From(err)
		}
		if err := cb(tmp); err != nil {
			return status.From(err)
		}
	}
	if err := rows.Err(); err != nil {
		return status.From(err)
	}

	return nil
}

func (s *scanStmt) buildScan() (string, []interface{}) {
	fmt.Fprintf(s.buf, "SELECT %s FROM %s", s.adap.Quote("data"), s.adap.Quote(s.name))

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
			fmt.Fprintf(s.buf, "%s = ?", s.adap.Quote(cols[i]))
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
	if len(startVals) != 0 || len(stopVals) != 0 {
		s.buf.WriteString(" WHERE ")
	}
	if len(startVals) != 0 {
		startStmt, startArgs := buildStart(startCols, startVals, s.adap)
		s.args = append(s.args, startArgs...)
		s.buf.WriteString(startStmt)
	}
	if len(startVals) != 0 && len(stopVals) != 0 {
		s.buf.WriteString(" AND ")
	}
	if len(stopVals) != 0 {
		stopStmt, stopArgs := buildStop(stopCols, stopVals, s.adap)
		s.args = append(s.args, stopArgs...)
		s.buf.WriteString(stopStmt)
	}
	if len(startCols) != 0 {
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
		s.buf.WriteString(s.adap.Quote(col))
		s.buf.WriteString(order)
	}
}

func (s *scanStmt) appendLock() {
	s.adap.LockStmt(s.buf, s.opts.Lock)
}

type Columns []string

func (cols Columns) String() string {
	panic("fix quote function")
	var parts []string
	for _, col := range cols {
		parts = append(parts /*quoteIdentifier(*/, col /*)*/)
	}
	return strings.Join(parts, ", ")
}

func buildStart(cols []string, vals []interface{}, adap DBAdapter) (string, []interface{}) {
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
			ands = append(ands, adap.Quote(cols[k])+" = ?")
			args = append(args, vals[k])
		}
		if i == len(vals)-1 {
			ands = append(ands, adap.Quote(cols[i])+" >= ?")
		} else {
			ands = append(ands, adap.Quote(cols[i])+" > ?")
		}
		args = append(args, vals[i])
		ors = append(ors, "("+strings.Join(ands, " AND ")+")")
	}
	return "(" + strings.Join(ors, " OR ") + ")", args
}

func buildStop(cols []string, vals []interface{}, adap DBAdapter) (string, []interface{}) {
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
			ands = append(ands, adap.Quote(cols[k])+" = ?")
			args = append(args, vals[k])
		}
		ands = append(ands, adap.Quote(cols[i])+" < ?")
		args = append(args, vals[i])
		ors = append(ors, "("+strings.Join(ands, " AND ")+")")
	}
	return "(" + strings.Join(ors, " OR ") + ")", args
}

const (
	errColsValsMismatch = "db: number of columns and values don't match."
	errNoCols           = "db: no columns provided"
)

func Insert(exec Executor, name string, cols []string, vals []interface{}, adap DBAdapter) error {
	return localInsert(exec, name, cols, vals, adap)
}

func localInsert(
	exec Executor, name string, cols []string, vals []interface{}, adap DBAdapter) status.S {
	if len(cols) != len(vals) {
		return status.InvalidArgument(nil, errColsValsMismatch)
	}
	if len(cols) == 0 {
		return status.InvalidArgument(nil, errNoCols)
	}

	valFmt := strings.Repeat("?, ", len(vals)-1) + "?"
	colFmtParts := make([]string, 0, len(cols))
	for _, col := range cols {
		colFmtParts = append(colFmtParts, adap.Quote(col))
	}
	colFmt := strings.Join(colFmtParts, ", ")
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", adap.Quote(name), colFmt, valFmt)
	if _, err := exec.Exec(query, vals...); err != nil {
		return status.From(err)
	}
	return nil
}

func Delete(exec Executor, name string, key UniqueIdx, adap DBAdapter) error {
	return localDelete(exec, name, key, adap)
}

func localDelete(exec Executor, name string, key UniqueIdx, adap DBAdapter) status.S {
	cols := key.Cols()
	vals := key.Vals()
	if len(cols) != len(vals) {
		return status.InvalidArgument(nil, errColsValsMismatch)
	}
	if len(cols) == 0 {
		return status.InvalidArgument(nil, errNoCols)
	}

	colFmtParts := make([]string, 0, len(cols))
	for _, col := range cols {
		colFmtParts = append(colFmtParts, adap.Quote(col)+" = ?")
	}
	colFmt := strings.Join(colFmtParts, " AND ")
	query := fmt.Sprintf("DELETE FROM %s WHERE %s;", adap.Quote(name), colFmt)
	if _, err := exec.Exec(query, vals...); err != nil {
		return status.From(err)
	}
	return nil
}

func Update(exec Executor, name string, cols []string, vals []interface{}, key UniqueIdx,
	adap DBAdapter) error {
	return localUpdate(exec, name, cols, vals, key, adap)
}

func localUpdate(exec Executor, name string, cols []string, vals []interface{}, key UniqueIdx,
	adap DBAdapter) status.S {
	if len(cols) != len(vals) {
		return status.InvalidArgument(nil, errColsValsMismatch)
	}
	if len(cols) == 0 {
		return status.InvalidArgument(nil, errNoCols)
	}

	idxCols := key.Cols()
	idxVals := key.Vals()
	if len(idxCols) != len(idxVals) {
		return status.InvalidArgument(nil, errColsValsMismatch)
	}
	if len(idxCols) == 0 {
		return status.InvalidArgument(nil, errNoCols)
	}

	colFmtParts := make([]string, 0, len(cols))
	for _, col := range cols {
		colFmtParts = append(colFmtParts, adap.Quote(col)+" = ?")
	}
	colFmt := strings.Join(colFmtParts, ", ")

	idxColFmtParts := make([]string, 0, len(idxCols))
	for _, idxCol := range idxCols {
		idxColFmtParts = append(idxColFmtParts, adap.Quote(idxCol)+" = ?")
	}
	idxColFmt := strings.Join(idxColFmtParts, " AND ")

	allVals := make([]interface{}, 0, len(vals)+len(idxVals))
	allVals = append(allVals, vals...)
	allVals = append(allVals, idxVals...)

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s;", adap.Quote(name), colFmt, idxColFmt)
	if _, err := exec.Exec(query, allVals...); err != nil {
		return status.From(err)
	}
	return nil
}
