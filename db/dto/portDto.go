package db

import (
	"database/sql"
	"fmt"
)

type PortDTO struct {
	HeaderName string
	Value      string
	Port       int
}

func (p *PortDTO) AddPortRecord(db *sql.DB) error {
	query := `INSERT INTO headerinfo (headername, value, port) VALUES (?, ?, ?)`
	_, err := db.Exec(query, p.HeaderName, p.Value, p.Port)
	if err != nil {
		return fmt.Errorf("failed to insert record: %v", err)
	}
	return nil
}

func GetPortByHeaderAndValue(db *sql.DB, headerName, value string) (*PortDTO, error) {
	query := `SELECT headername, value, port FROM headerinfo WHERE headername = ? AND value = ?`
	row := db.QueryRow(query, headerName, value)

	var dto PortDTO
	err := row.Scan(&dto.HeaderName, &dto.Value, &dto.Port)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Если запись не найдена, возвращаем nil.
		}
		return nil, fmt.Errorf("failed to query record: %v", err)
	}

	return &dto, nil
}
