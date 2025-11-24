package sql

import (
	"database/sql"
	"fmt"
	"reflect"

	"github.com/jackc/pgx/v5"
)

// scanRowsToResult scans rows and converts to the provided result pointer
func scanRowsToResult(rows pgx.Rows, result interface{}) error {
	if result == nil {
		return fmt.Errorf("result pointer cannot be nil")
	}

	// Get the value that the pointer points to
	val := reflect.ValueOf(result)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("result must be a pointer")
	}

	// Get the element type
	elemType := val.Elem().Type()

	// Handle different result types
	switch elemType.Kind() {
	case reflect.Slice:
		return scanRowsToSlice(rows, result)
	case reflect.Struct:
		return scanRowsToStruct(rows, result)
	case reflect.Map:
		return scanRowsToMap(rows, result)
	default:
		return fmt.Errorf("unsupported result type: %s", elemType.Kind())
	}
}

// scanRowsToSlice scans rows into a slice
func scanRowsToSlice(rows pgx.Rows, result interface{}) error {
	val := reflect.ValueOf(result).Elem()
	elemType := val.Type().Elem()

	// Get column names
	fieldDescriptions := rows.FieldDescriptions()
	columns := make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		columns[i] = string(fd.Name)
	}

	// Scan each row
	for rows.Next() {
		// Create a new element
		elem := reflect.New(elemType).Elem()

		// Handle different element types
		switch elemType.Kind() {
		case reflect.Struct:
			if err := scanRowToStruct(rows, elem, columns); err != nil {
				return err
			}
		case reflect.Map:
			if err := scanRowToMapElement(rows, elem, columns); err != nil {
				return err
			}
		default:
			// For primitive types, scan directly
			if err := rows.Scan(elem.Addr().Interface()); err != nil {
				return fmt.Errorf("failed to scan row: %w", err)
			}
		}

		// Append to slice
		val.Set(reflect.Append(val, elem))
	}

	return rows.Err()
}

// scanRowsToStruct scans rows into a struct (expecting single row)
func scanRowsToStruct(rows pgx.Rows, result interface{}) error {
	val := reflect.ValueOf(result)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("result must be a pointer to struct")
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("result must be a pointer to struct")
	}

	// Get column names
	fieldDescriptions := rows.FieldDescriptions()
	columns := make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		columns[i] = string(fd.Name)
	}

	// Scan the first row
	if rows.Next() {
		return scanRowToStruct(rows, val, columns)
	}

	return sql.ErrNoRows
}

// scanRowToStruct scans a single row into a struct
func scanRowToStruct(rows pgx.Rows, val reflect.Value, columns []string) error {
	// Create values slice for scanning
	values := make([]interface{}, len(columns))
	scanArgs := make([]interface{}, len(values))

	for i := range values {
		scanArgs[i] = &values[i]
	}

	// Scan the row
	if err := rows.Scan(scanArgs...); err != nil {
		return fmt.Errorf("failed to scan row: %w", err)
	}

	// Map values to struct fields
	return mapValuesToStruct(val, columns, values)
}

// scanRowToMapElement scans a single row into a map element
func scanRowToMapElement(rows pgx.Rows, elem reflect.Value, columns []string) error {
	// Create values slice for scanning
	values := make([]interface{}, len(columns))
	scanArgs := make([]interface{}, len(values))

	for i := range values {
		scanArgs[i] = &values[i]
	}

	// Scan the row
	if err := rows.Scan(scanArgs...); err != nil {
		return fmt.Errorf("failed to scan row: %w", err)
	}

	// Create map
	elem.Set(reflect.MakeMap(elem.Type()))

	// Map values to map
	for i, col := range columns {
		elem.SetMapIndex(reflect.ValueOf(col), reflect.ValueOf(values[i]))
	}

	return nil
}

// scanRowsToMap scans rows into a map (expecting single row)
func scanRowsToMap(rows pgx.Rows, result interface{}) error {
	val := reflect.ValueOf(result)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("result must be a pointer to map")
	}

	val = val.Elem()
	if val.Kind() != reflect.Map {
		return fmt.Errorf("result must be a pointer to map")
	}

	// Get column names
	fieldDescriptions := rows.FieldDescriptions()
	columns := make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		columns[i] = string(fd.Name)
	}

	// Scan the first row
	if rows.Next() {
		return scanRowToMapElement(rows, val, columns)
	}

	return sql.ErrNoRows
}

// mapValuesToStruct maps column values to struct fields
func mapValuesToStruct(val reflect.Value, columns []string, values []interface{}) error {
	typ := val.Type()

	for i, col := range columns {
		// Find struct field by tag or name
		field := findStructField(typ, col)
		if field.Name == "" {
			// Skip if field not found
			continue
		}

		// Set the field value
		if err := setStructField(val.FieldByIndex(field.Index), values[i]); err != nil {
			return fmt.Errorf("failed to set field %s: %w", col, err)
		}
	}

	return nil
}

// findStructField finds a struct field by name or tag
func findStructField(typ reflect.Type, name string) reflect.StructField {
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		// Check field name
		if field.Name == name {
			return field
		}

		// Check json tag
		if tag := field.Tag.Get("json"); tag == name {
			return field
		}

		// Check db tag
		if tag := field.Tag.Get("db"); tag == name {
			return field
		}
	}

	return reflect.StructField{}
}

// setStructField sets a struct field value
func setStructField(field reflect.Value, value interface{}) error {
	if !field.CanSet() {
		return fmt.Errorf("field cannot be set")
	}

	// Handle nil values
	if value == nil {
		field.Set(reflect.Zero(field.Type()))
		return nil
	}

	// Convert value to field type
	val := reflect.ValueOf(value)

	// If types match, set directly
	if val.Type().AssignableTo(field.Type()) {
		field.Set(val)
		return nil
	}

	// Handle type conversions
	switch field.Kind() {
	case reflect.String:
		field.SetString(fmt.Sprintf("%v", value))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch v := value.(type) {
		case int:
			field.SetInt(int64(v))
		case int64:
			field.SetInt(v)
		case float64:
			field.SetInt(int64(v))
		default:
			field.SetInt(0)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch v := value.(type) {
		case int:
			field.SetUint(uint64(v))
		case int64:
			field.SetUint(uint64(v))
		case float64:
			field.SetUint(uint64(v))
		default:
			field.SetUint(0)
		}
	case reflect.Float32, reflect.Float64:
		switch v := value.(type) {
		case float64:
			field.SetFloat(v)
		case int:
			field.SetFloat(float64(v))
		case int64:
			field.SetFloat(float64(v))
		default:
			field.SetFloat(0)
		}
	case reflect.Bool:
		switch v := value.(type) {
		case bool:
			field.SetBool(v)
		default:
			field.SetBool(fmt.Sprintf("%v", v) == "true")
		}
	default:
		return fmt.Errorf("unsupported field type: %s", field.Type())
	}

	return nil
}

// scanSingleValue scans a single value from a row
func scanSingleValue(row pgx.Row, result interface{}) error {
	return row.Scan(result)
}

// scanMultipleValues scans multiple values from a row
func scanMultipleValues(row pgx.Row, results ...interface{}) error {
	return row.Scan(results...)
}
