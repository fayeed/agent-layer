package postgres

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type stringArrayValue []string

func (a stringArrayValue) Value() (driver.Value, error) {
	return json.Marshal([]string(a))
}

type stringArrayScan struct {
	target *[]string
}

func pqTextArray(values []string) driver.Valuer {
	return stringArrayValue(values)
}

func pqArrayScan(target *[]string) interface{ Scan(src any) error } {
	return stringArrayScan{target: target}
}

func (s stringArrayScan) Scan(src any) error {
	switch value := src.(type) {
	case nil:
		*s.target = nil
		return nil
	case []byte:
		return json.Unmarshal(value, s.target)
	case string:
		return json.Unmarshal([]byte(value), s.target)
	default:
		return fmt.Errorf("unsupported array scan type %T", src)
	}
}
