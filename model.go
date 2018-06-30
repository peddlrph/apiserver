package main

import (
	"database/sql"
	//"errors"
	//	"net/http"
)

type message struct {
	Id       int    `json:"id"`
	Body     string `json:"body"`
	Msg_Box  string `json:"msg_box"`
	Address  string `json:"address"`
	DateTime string `json:"datetime"`
}

func (m *message) getLastMessage(db *sql.DB) error {
	return db.QueryRow("SELECT id,body,msg_box,address,datetime FROM messages ORDER by ID desc LIMIT 1").Scan(&m.Id, &m.Body, &m.Msg_Box, &m.Address, &m.DateTime)
}

func (m *message) getMessage(db *sql.DB) error {
	return db.QueryRow("SELECT body,msg_box,address,datetime FROM messages WHERE id=?", m.Id).Scan(&m.Body, &m.Msg_Box, &m.Address, &m.DateTime)
}

func (m *message) createMessage(db *sql.DB) error {
	_, err := db.Exec("INSERT INTO messages(id,body,msg_box,address,datetime) VALUES(?,?,?,?,?)", m.Id, m.Body, m.Msg_Box, m.Address, m.DateTime)

	if err != nil {
		return err
	}

	return nil
}

func getMessages(db *sql.DB, start, count int) ([]message, error) {
	rows, err := db.Query(
		"SELECT id, body,msg_box,address,datetime FROM messages LIMIT ? OFFSET ?",
		count, start)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	messages := []message{}

	for rows.Next() {
		var m message
		if err := rows.Scan(&m.Id, &m.Body, &m.Msg_Box, &m.Address, &m.DateTime); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}

	return messages, nil
}
