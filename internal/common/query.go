package common

import (
	"errors"
	"fmt"
	"github.com/FreeLeh/GoFreeDB/internal/models"
	"regexp"
	"strconv"
	"strings"
)

var selectStmtStringKeyword = regexp.MustCompile("^(date|datetime|timeofday)")

type whereInterceptorFunc func(where string) string

type QueryBuilder struct {
	replacer         *strings.Replacer
	columns          []string
	where            string
	whereArgs        []interface{}
	whereInterceptor whereInterceptorFunc
	orderBy          []string
	limit            uint64
	offset           uint64
}

func (q *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
	q.where = condition
	q.whereArgs = args
	return q
}

func (q *QueryBuilder) OrderBy(ordering []models.ColumnOrderBy) *QueryBuilder {
	orderBy := make([]string, 0, len(ordering))
	for _, o := range ordering {
		orderBy = append(orderBy, o.Column+" "+string(o.OrderBy))
	}

	q.orderBy = orderBy
	return q
}

func (q *QueryBuilder) Limit(limit uint64) *QueryBuilder {
	q.limit = limit
	return q
}

func (q *QueryBuilder) Offset(offset uint64) *QueryBuilder {
	q.offset = offset
	return q
}

func (q *QueryBuilder) Generate() (string, error) {
	stmt := &strings.Builder{}
	stmt.WriteString("select")

	if err := q.writeCols(stmt); err != nil {
		return "", err
	}
	if err := q.writeWhere(stmt); err != nil {
		return "", err
	}
	if err := q.writeOrderBy(stmt); err != nil {
		return "", err
	}
	if err := q.writeOffset(stmt); err != nil {
		return "", err
	}
	if err := q.writeLimit(stmt); err != nil {
		return "", err
	}

	return stmt.String(), nil
}

func (q *QueryBuilder) writeCols(stmt *strings.Builder) error {
	stmt.WriteString(" ")

	translated := make([]string, 0, len(q.columns))
	for _, col := range q.columns {
		translated = append(translated, q.replacer.Replace(col))
	}

	stmt.WriteString(strings.Join(translated, ", "))
	return nil
}

func (q *QueryBuilder) writeWhere(stmt *strings.Builder) error {
	where := q.where
	if q.whereInterceptor != nil {
		where = q.whereInterceptor(q.where)
	}

	nArgs := strings.Count(where, "?")
	if nArgs != len(q.whereArgs) {
		return fmt.Errorf("number of arguments required in the 'where' clause (%d) is not the same as the number of provided arguments (%d)", nArgs, len(q.whereArgs))
	}

	where = q.replacer.Replace(where)
	tokens := strings.Split(where, "?")

	result := make([]string, 0)
	result = append(result, strings.TrimSpace(tokens[0]))

	for i, token := range tokens[1:] {
		arg, err := q.convertArg(q.whereArgs[i])
		if err != nil {
			return fmt.Errorf("failed converting 'where' arguments: %v, %w", arg, err)
		}
		result = append(result, arg, strings.TrimSpace(token))
	}

	stmt.WriteString(" where ")
	stmt.WriteString(strings.Join(result, " "))
	return nil
}

func (q *QueryBuilder) convertArg(arg interface{}) (string, error) {
	switch converted := arg.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return q.convertInt(arg)
	case float32, float64:
		return q.convertFloat(arg)
	case string, []byte:
		return q.convertString(arg)
	case bool:
		return strconv.FormatBool(converted), nil
	default:
		return "", errors.New("unsupported argument type")
	}
}

func (q *QueryBuilder) convertInt(arg interface{}) (string, error) {
	switch converted := arg.(type) {
	case int:
		return strconv.FormatInt(int64(converted), 10), nil
	case int8:
		return strconv.FormatInt(int64(converted), 10), nil
	case int16:
		return strconv.FormatInt(int64(converted), 10), nil
	case int32:
		return strconv.FormatInt(int64(converted), 10), nil
	case int64:
		return strconv.FormatInt(converted, 10), nil
	case uint:
		return strconv.FormatUint(uint64(converted), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(converted), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(converted), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(converted), 10), nil
	case uint64:
		return strconv.FormatUint(converted, 10), nil
	default:
		return "", errors.New("unsupported argument type")
	}
}

func (q *QueryBuilder) convertFloat(arg interface{}) (string, error) {
	switch converted := arg.(type) {
	case float32:
		return strconv.FormatFloat(float64(converted), 'f', -1, 64), nil
	case float64:
		return strconv.FormatFloat(converted, 'f', -1, 64), nil
	default:
		return "", errors.New("unsupported argument type")
	}
}

func (q *QueryBuilder) convertString(arg interface{}) (string, error) {
	switch converted := arg.(type) {
	case string:
		cleaned := strings.ToLower(strings.TrimSpace(converted))
		if selectStmtStringKeyword.MatchString(cleaned) {
			return converted, nil
		}
		return strconv.Quote(converted), nil
	case []byte:
		return strconv.Quote(string(converted)), nil
	default:
		return "", errors.New("unsupported argument type")
	}
}

func (q *QueryBuilder) writeOrderBy(stmt *strings.Builder) error {
	if len(q.orderBy) == 0 {
		return nil
	}

	stmt.WriteString(" order by ")
	result := make([]string, 0, len(q.orderBy))

	for _, o := range q.orderBy {
		result = append(result, q.replacer.Replace(o))
	}

	stmt.WriteString(strings.Join(result, ", "))
	return nil
}

func (q *QueryBuilder) writeOffset(stmt *strings.Builder) error {
	if q.offset == 0 {
		return nil
	}

	stmt.WriteString(" offset ")
	stmt.WriteString(strconv.FormatInt(int64(q.offset), 10))
	return nil
}

func (q *QueryBuilder) writeLimit(stmt *strings.Builder) error {
	if q.limit == 0 {
		return nil
	}

	stmt.WriteString(" limit ")
	stmt.WriteString(strconv.FormatInt(int64(q.limit), 10))
	return nil
}

func NewQueryBuilder(
	colReplacements map[string]string,
	whereInterceptor whereInterceptorFunc,
	colSelected []string,
) *QueryBuilder {
	replacements := make([]string, 0, 2*len(colReplacements))
	for col, repl := range colReplacements {
		replacements = append(replacements, col, repl)
	}

	return &QueryBuilder{
		replacer:         strings.NewReplacer(replacements...),
		columns:          colSelected,
		whereInterceptor: whereInterceptor,
	}
}
