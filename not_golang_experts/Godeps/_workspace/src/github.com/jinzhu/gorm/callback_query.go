package gorm

import (
	"fmt"
	"reflect"
)

func Query(scope *Scope) {
	defer scope.Trace(NowFunc())

	var (
		isSlice        bool
		isPtr          bool
		anyRecordFound bool
		destType       reflect.Type
	)

	var dest = scope.IndirectValue()
	if value, ok := scope.InstanceGet("gorm:query_destination"); ok {
		dest = reflect.Indirect(reflect.ValueOf(value))
	}

	if orderBy, ok := scope.InstanceGet("gorm:order_by_primary_key"); ok {
		if primaryKey := scope.PrimaryKey(); primaryKey != "" {
			scope.Search = scope.Search.clone().order(fmt.Sprintf("%v.%v %v", scope.QuotedTableName(), primaryKey, orderBy))
		}
	}

	if dest.Kind() == reflect.Slice {
		isSlice = true
		destType = dest.Type().Elem()
		if destType.Kind() == reflect.Ptr {
			isPtr = true
			destType = destType.Elem()
		}
	} else {
		scope.Search = scope.Search.clone().limit(1)
	}

	scope.prepareQuerySql()

	if !scope.HasError() {
		rows, err := scope.DB().Query(scope.Sql, scope.SqlVars...)

		if scope.Err(err) != nil {
			return
		}

		defer rows.Close()
		columns, _ := rows.Columns()
		for rows.Next() {
			anyRecordFound = true
			elem := dest
			if isSlice {
				elem = reflect.New(destType).Elem()
			}

			var values = make([]interface{}, len(columns))

			fields := scope.New(elem.Addr().Interface()).Fields()
			for index, column := range columns {
				if field, ok := fields[column]; ok {
					if field.Field.Kind() == reflect.Ptr {
						values[index] = field.Field.Addr().Interface()
					} else {
						values[index] = reflect.New(reflect.PtrTo(field.Field.Type())).Interface()
					}
				} else {
					var value interface{}
					values[index] = &value
				}
			}

			scope.Err(rows.Scan(values...))

			for index, column := range columns {
				value := values[index]
				if field, ok := fields[column]; ok {
					if field.Field.Kind() == reflect.Ptr {
						field.Field.Set(reflect.ValueOf(value).Elem())
					} else if v := reflect.ValueOf(value).Elem().Elem(); v.IsValid() {
						field.Field.Set(v)
					}
				}
			}

			if isSlice {
				if isPtr {
					dest.Set(reflect.Append(dest, elem.Addr()))
				} else {
					dest.Set(reflect.Append(dest, elem))
				}
			}
		}

		if !anyRecordFound && !isSlice {
			scope.Err(RecordNotFound)
		}
	}
}

func AfterQuery(scope *Scope) {
	scope.CallMethod("AfterFind")
}

func init() {
	DefaultCallback.Query().Register("gorm:query", Query)
	DefaultCallback.Query().Register("gorm:after_query", AfterQuery)
}
