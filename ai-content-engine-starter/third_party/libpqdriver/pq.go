package pq

/*
#cgo pkg-config: libpq
#include <libpq-fe.h>
#include <stdlib.h>
*/
import "C"

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

func init() {
	sql.Register("postgres", &postgresDriver{})
}

type postgresDriver struct{}

func (d *postgresDriver) Open(name string) (driver.Conn, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	conn := C.PQconnectdb(cname)
	if conn == nil {
		return nil, fmt.Errorf("connect postgres: nil connection")
	}
	if C.PQstatus(conn) != C.CONNECTION_OK {
		err := pqError(conn, "connect postgres")
		C.PQfinish(conn)
		return nil, err
	}
	return &postgresConn{conn: conn}, nil
}

type postgresConn struct {
	conn *C.PGconn
}

func (c *postgresConn) Prepare(string) (driver.Stmt, error) {
	return nil, fmt.Errorf("prepare is not supported")
}

func (c *postgresConn) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	C.PQfinish(c.conn)
	c.conn = nil
	return nil
}

func (c *postgresConn) Begin() (driver.Tx, error) {
	return nil, fmt.Errorf("transactions are not supported")
}

func (c *postgresConn) Ping(context.Context) error {
	if c == nil || c.conn == nil {
		return fmt.Errorf("postgres connection is nil")
	}
	if C.PQstatus(c.conn) != C.CONNECTION_OK {
		return pqError(c.conn, "ping postgres")
	}
	return nil
}

func (c *postgresConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	result, err := c.exec(query, args)
	if err != nil {
		return nil, err
	}
	defer C.PQclear(result)

	status := C.PQresultStatus(result)
	if status != C.PGRES_COMMAND_OK && status != C.PGRES_TUPLES_OK {
		return nil, pqResultError(c.conn, result, "exec postgres query")
	}
	rowsAffected, err := parseRowsAffected(result)
	if err != nil {
		return nil, err
	}
	return driver.RowsAffected(rowsAffected), nil
}

func (c *postgresConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	result, err := c.exec(query, args)
	if err != nil {
		return nil, err
	}
	status := C.PQresultStatus(result)
	if status != C.PGRES_TUPLES_OK {
		err := pqResultError(c.conn, result, "query postgres rows")
		C.PQclear(result)
		return nil, err
	}
	return &postgresRows{
		result:    result,
		rowCount:  int(C.PQntuples(result)),
		columnCnt: int(C.PQnfields(result)),
	}, nil
}

func (c *postgresConn) exec(query string, args []driver.NamedValue) (*C.PGresult, error) {
	if c == nil || c.conn == nil {
		return nil, fmt.Errorf("postgres connection is nil")
	}
	cquery := C.CString(query)
	defer C.free(unsafe.Pointer(cquery))

	values := make([]*C.char, len(args))
	defer func() {
		for _, value := range values {
			if value != nil {
				C.free(unsafe.Pointer(value))
			}
		}
	}()

	for i, arg := range args {
		if arg.Value == nil {
			continue
		}
		text, err := formatValue(arg.Value)
		if err != nil {
			return nil, err
		}
		values[i] = C.CString(text)
	}

	var valuePtr **C.char
	if len(values) > 0 {
		valuePtr = (**C.char)(unsafe.Pointer(&values[0]))
	}
	result := C.PQexecParams(c.conn, cquery, C.int(len(args)), nil, valuePtr, nil, nil, 0)
	if result == nil {
		return nil, pqError(c.conn, "exec postgres query")
	}
	return result, nil
}

type postgresRows struct {
	result    *C.PGresult
	rowIndex  int
	rowCount  int
	columnCnt int
}

func (r *postgresRows) Columns() []string {
	columns := make([]string, r.columnCnt)
	for i := 0; i < r.columnCnt; i++ {
		columns[i] = C.GoString(C.PQfname(r.result, C.int(i)))
	}
	return columns
}

func (r *postgresRows) Close() error {
	if r == nil || r.result == nil {
		return nil
	}
	C.PQclear(r.result)
	r.result = nil
	return nil
}

func (r *postgresRows) Next(dest []driver.Value) error {
	if r.rowIndex >= r.rowCount {
		return io.EOF
	}
	for i := 0; i < r.columnCnt; i++ {
		if C.PQgetisnull(r.result, C.int(r.rowIndex), C.int(i)) == 1 {
			dest[i] = nil
			continue
		}
		raw := C.GoStringN(C.PQgetvalue(r.result, C.int(r.rowIndex), C.int(i)), C.int(C.PQgetlength(r.result, C.int(r.rowIndex), C.int(i))))
		dest[i] = parseValue(uint32(C.PQftype(r.result, C.int(i))), raw)
	}
	r.rowIndex++
	return nil
}

func pqError(conn *C.PGconn, prefix string) error {
	if conn == nil {
		return fmt.Errorf("%s: unknown postgres error", prefix)
	}
	message := strings.TrimSpace(C.GoString(C.PQerrorMessage(conn)))
	if message == "" {
		message = "unknown postgres error"
	}
	return fmt.Errorf("%s: %s", prefix, message)
}

func pqResultError(conn *C.PGconn, result *C.PGresult, prefix string) error {
	if result != nil {
		message := strings.TrimSpace(C.GoString(C.PQresultErrorMessage(result)))
		if message != "" {
			return fmt.Errorf("%s: %s", prefix, message)
		}
	}
	return pqError(conn, prefix)
}

func parseRowsAffected(result *C.PGresult) (int64, error) {
	count := strings.TrimSpace(C.GoString(C.PQcmdTuples(result)))
	if count == "" {
		return 0, nil
	}
	value, err := strconv.ParseInt(count, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse rows affected: %w", err)
	}
	return value, nil
}

func formatValue(value any) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case bool:
		if v {
			return "true", nil
		}
		return "false", nil
	case int:
		return strconv.Itoa(v), nil
	case int8, int16, int32, int64:
		return fmt.Sprintf("%d", v), nil
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v), nil
	case float32, float64:
		return fmt.Sprintf("%v", v), nil
	case time.Time:
		return v.UTC().Format(time.RFC3339Nano), nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

func parseValue(oid uint32, raw string) driver.Value {
	switch oid {
	case 16:
		return raw == "t" || strings.EqualFold(raw, "true")
	case 20, 21, 23:
		if value, err := strconv.ParseInt(raw, 10, 64); err == nil {
			return value
		}
	case 700, 701, 1700:
		if value, err := strconv.ParseFloat(raw, 64); err == nil {
			return value
		}
	case 1114:
		if ts, err := parseTimestamp(raw, false); err == nil {
			return ts
		}
	case 1184:
		if ts, err := parseTimestamp(raw, true); err == nil {
			return ts
		}
	}
	return raw
}

func parseTimestamp(raw string, withZone bool) (time.Time, error) {
	layouts := []string{
		time.RFC3339Nano,
		"2006-01-02 15:04:05.999999999-07",
		"2006-01-02 15:04:05.999999-07",
		"2006-01-02 15:04:05-07",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05.999999",
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		var (
			ts  time.Time
			err error
		)
		if withZone {
			ts, err = time.Parse(layout, raw)
		} else {
			ts, err = time.ParseInLocation(layout, raw, time.UTC)
		}
		if err == nil {
			return ts, nil
		}
	}
	return time.Time{}, fmt.Errorf("parse timestamp %q", raw)
}
