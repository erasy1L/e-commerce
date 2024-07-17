package store

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

type OnlyDate string

const DateLayout = "2006-01-02"

// method of [driver.Valuer] interface
func (o *OnlyDate) Value() (interface{}, error) {
	nt, err := time.Parse(DateLayout, string(*o))
	if err != nil {
		return nil, err
	}

	return nt, nil
}

// method of [sql.Scanner] interface
func (o *OnlyDate) Scan(val interface{}) error {
	nt, ok := val.(time.Time)
	if !ok {
		return fmt.Errorf("expected time.Time, got %T", val)
	}

	*o = OnlyDate(nt.Format(DateLayout))

	return nil
}

func (o *OnlyDate) String() string {
	return string(*o)
}

func GenerateID() string {
	bytes := make([]byte, 12)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
