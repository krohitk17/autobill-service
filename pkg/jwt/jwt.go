package JWTUtil

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	Id  string `json:"id"`
	Exp int64  `json:"exp"`
}

func (j JWTClaims) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":  j.Id,
		"exp": j.Exp,
	}
}

func (j *JWTClaims) FromMap(m map[string]interface{}) error {
	if m == nil {
		return fmt.Errorf("input map is nil")
	}

	if v, ok := m["id"]; ok && v != nil {
		switch val := v.(type) {
		case string:
			j.Id = val
		default:
			j.Id = fmt.Sprintf("%v", val)
		}
	}

	if v, ok := m["exp"]; ok && v != nil {
		switch val := v.(type) {
		case jwt.NumericDate:
			j.Exp = val.Local().Unix()
		case time.Time:
			j.Exp = val.Local().Unix()
		case json.Number:
			if parsed, err := val.Int64(); err == nil {
				j.Exp = parsed
			} else {
				if f, err2 := val.Float64(); err2 == nil {
					j.Exp = int64(f)
				} else {
					return fmt.Errorf("invalid exp value: %v", err)
				}
			}
		case string:
			if parsed, err := strconv.ParseInt(val, 10, 64); err == nil {
				j.Exp = parsed
			} else {
				if f, err2 := strconv.ParseFloat(val, 64); err2 == nil {
					j.Exp = int64(f)
				} else {
					return fmt.Errorf("invalid exp value: %v", err)
				}
			}
		default:
			return fmt.Errorf("unsupported type for exp: %T", val)
		}
	}

	return nil
}
