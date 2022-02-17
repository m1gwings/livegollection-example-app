package chat

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Message struct {
	Id       int64     `json:"id,omitempty"`
	Sender   string    `json:"sender,omitempty"`
	SentTime time.Time `json:"sentTime,omitempty"`
	Text     string    `json:"text,omitempty"`
}

func (mess *Message) ID() int64 {
	return mess.Id
}

type Chat struct {
	db *sql.DB
}

const dbFilePath = "./chat.db"

func NewChat() (*Chat, error) {
	db, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		return nil, fmt.Errorf("error while opening the database: %v", err)
	}

	_, err = db.Exec(createChatTable)
	if err != nil {
		return nil, fmt.Errorf("error while creating chat table: %v", err)
	}

	return &Chat{db: db}, nil
}

func (c *Chat) All() ([]*Message, error) {
	rows, err := c.db.Query(getAllMessagesFromChat)
	if err != nil {
		return nil, fmt.Errorf("error while executing query in All: %v", err)
	}

	messages := make([]*Message, 0)
	for rows.Next() {
		var id int64
		var sender string
		var sentTime time.Time
		var text string
		if err := rows.Scan(&id, &sender, &sentTime, &text); err != nil {
			return nil, fmt.Errorf("error while scanning a row in All: %v", err)
		}

		messages = append(messages, &Message{
			Id:       id,
			Sender:   sender,
			SentTime: sentTime,
			Text:     text,
		})
	}

	return messages, nil
}

func (c *Chat) Item(ID int64) (*Message, error) {
	row := c.db.QueryRow(getMessageFromChat, ID)
	var id int64
	var sender string
	var sentTime time.Time
	var text string
	if err := row.Scan(&id, &sender, &sentTime, &text); err != nil {
		return nil, fmt.Errorf("error while scanning the row in Item: %v", err)
	}

	return &Message{
		Id:       id,
		Sender:   sender,
		SentTime: sentTime,
		Text:     text,
	}, nil
}

func (c *Chat) Create(mess *Message) (*Message, error) {
	res, err := c.db.Exec(insertMessageIntoChat, mess.Sender, mess.SentTime, mess.Text)
	if err != nil {
		return nil, fmt.Errorf("error while inserting the message in Create: %v", err)
	}

	// It's IMPORTANT to set the ID of the message before returning it
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("error while retreiving the last id in Create: %v", err)
	}

	mess.Id = id

	return mess, nil
}

func (c *Chat) Update(mess *Message) error {
	_, err := c.db.Exec(updateMessageInChat, mess.Text, mess.Id)
	if err != nil {
		return fmt.Errorf("error while updating the message in Update: %v", err)
	}

	return nil
}

func (c *Chat) Delete(ID int64) error {
	_, err := c.db.Exec(deleteMessageFromChat, ID)
	if err != nil {
		return fmt.Errorf("error while deleting the message in Delete: %v", err)
	}

	return nil
}
