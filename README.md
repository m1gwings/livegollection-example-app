# livegollection-example-app
**livegollection-example-app** is a simple web-chat app that demonstrates how the Golang **[livegollection](https://github.com/m1gwings/livegollection)** library can be used for live data synchronization between multiple web clients and the server.
The app allows to exchange text messages inside a chat room, **livegollection** will take care of notifying each client when a new message has been sent and keep everything consistent with the collection of messages stored in a SQLite database by the server.
# Step-by-step guide
The following guide will explain step-by-step how to create this web-app and how to use **livegollection**.
## Project setup
Create the directory that will house project:
```bash
mkdir livegollection-example-app
cd livegollection-example-app
```
## Initialize the Golang module
Use go mod init to initialize the Golang module for the app:
```bash
go mod init module-name
```
In my case module-name is github.com/m1gwings/livegollection-example-app.
## Implement the Chat collection
In order to use the **livegollection** library we need to implement a collection that satisfies the **livegollection.Collection** interface.
We'll define this collection inside the **chat** package.

Create a directory for the chat package:
```bash
mkdir chat
cd chat
```
As we said our backend will store the messages of the chat inside a SQLite database, so, first of all,
add `queries.go` inside the chat package in order to define the needed SQL query templates:
```bash
cat queries.go
```
```go
package chat

const createChatTable = `CREATE TABLE IF NOT EXISTS chat (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	sender VARCHAR,
	sent_time DATETIME,
	text TEXT
);`

const insertMessageIntoChat = `INSERT INTO chat(sender, sent_time, text)
	VALUES(?, ?, ?);`

const updateMessageInChat = `UPDATE chat
	SET text = ?
	WHERE id = ?;`

const deleteMessageFromChat = `DELETE from chat
	WHERE id = ?;`

const getAllMessagesFromChat = `SELECT * FROM chat;`

const getMessageFromChat = `SELECT * FROM chat WHERE id = ?;`

```
To execute the queries above we need a SQLite driver. Let's download it and add to our dependencies:
```bash
go get github.com/mattn/go-sqlite3
```
Now it's time to implement the actual collection inside `chat.go`:
```bash
cat chat.go
```
```go
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
```
I'd like to emphasize (as mentioned in the comment) that it's IMPORTANT to set the message ID retreived from the database when we add a new message to the chat in Create.
## Implement server executable
We need to write backend logic in order to serve static content (that will be added to the project in the next step) and register livegollection handler to the HTTP server.

In particular for the second point we need to download and add to our dependencies the **livegollection** library:
```bash
go get github.com/m1gwings/livegollection
```

Create a directory for the server executable:
```bash
cd .. # (If you are inside chat directory)
mkdir server
cd server
```

Write the code for the server inside `main.go`:
```bash
cat main.go
```
```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/m1gwings/livegollection"
	"module-name/chat"
)

func main() {
	for r, fP := range map[string]string{
		"/":          "../static/index.html",
		"/bundle.js": "../static/bundle.js",
		"/style.css": "../static/style.css",
	} {
		// Prevents passing loop variables to a closure.
		route, filePath := r, fP
		http.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filePath)
		})
	}

	coll, err := chat.NewChat()
	if err != nil {
		log.Fatal(fmt.Errorf("error when creating new chat: %v", err))
	}

	// When we create the LiveGollection we need to specify the type parameters for items' id and items themselves.
	// (In this case int64 and *chat.Message)
	liveGoll := livegollection.NewLiveGollection[int64, *chat.Message](context.TODO(), coll, log.Default())

	// After we created liveGoll, to enable it we just need to register a route handled by liveGoll.Join.
	http.HandleFunc("/livegollection", liveGoll.Join)

	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
```
As you can see, once you have implemented the collection, it's very easy to start using **livegollection**: you just need to create an istance of LiveGollection (with NewLiveGollection factory function) and register a route handled by liveGoll.Join.
## Add static content
We need to add an HTML page and some CSS to display the chat to our users.

Create a directory for static content:
```bash
cd .. # (If you are inside server directory)
mkdir static
cd static
```
Add the following HTML page in `index.html`:
```bash
cat index.html
```
```html
<!DOCTYPE html>
<html>
<head>
    <title>livegollection Chat</title>
    <script src="bundle.js"></script>
    <link rel="stylesheet" href="style.css" />
</head>
<body>
    <div class="chat">
        <p>livegollection Chat</p>
        <div class="inbox" id="inbox-div"></div>
        <div class="send-bar">
            <input type="text" id="message-text-input" value="" />
            <input type="button" id="send-button" value="Send" />
        </div>
        <div class="bottom-white-space"></div>
    </div>
</body>
</html>
```
and some styling in `style.css`:
```bash
cat style.css
```
```css
* {
    color: #22223B;
}

html, body {
    width: 100%;
    height: 100%;
    margin: 0%;
}

body {
    background-color: #F2E9E4;
}

.chat {
    display: flex;
    flex-direction: column;
    height: 100%;
    width: 100%;
    align-items: center;
}

.inbox {
    display: flex;
    flex-direction: column;
    width: 500px;
    height: 100%;
    overflow: scroll;
}

.mine {
    align-self: flex-end;
}

.others {
    align-self: flex-start;
}

.message {
    width: fit-content;
    margin-bottom: 10px;
    border-radius: 5px;
    background-color:#C9ADA7;
    padding: 5px;
}

.sender {
    font-size: small;
    margin: 0px;
}

.time {
    font-size: xx-small;
    margin: 0px;
}

.send-bar {
    display: flex;
    flex-direction: row;
    height: 50px;
    width: 250px;
    justify-content: center;
    border-radius: 5px;
    background-color: #C9ADA7;
    padding: 5px;
}

.bottom-white-space {
    height: 50px;
}

input {
    border: 0px;
    margin-left: 10px;
    border-radius: 5px;
}

input[type=text] {
    background-color: transparent;
}

input[type=button] {
    background-color: #9A8C98;
}
```
## Implement client-side logic
To make the chat actually working we need to implement the JavaScript script that will interact with the server and manage the DOM.

To interact with the server we'll use **[livegollection-client](https://github.com/m1gwings/livegollection-client)** which is the JavaScript/TypeScript client library for **livegollection**.

**livegollection-client** is distributed on npm, furthermore we'll use webpack to bundle our script. So let's setup npm and add our dependencies:
```bash
cd .. # (If you are inside static directory)
npm init -y
npm install livegollection-client
npm install webpack webpack-cli --save-dev
```

Create the directory where we'll write our script:
```bash
mkdir src
cd src
```

and add the following JavaScript code inside `index.js`:
```bash
cat index.js
```
```javascript
import LiveGollection from "../node_modules/livegollection-client/dist/index.js";

function generateClientTag() {
    return Math.random().toString(36).substr(2, 6);
}

const clientTag = generateClientTag();
const me = `Client#${clientTag}`;

function getMessageDivId(id) {
    return  `message-${id}`;
}

let inboxDiv = null;
let liveGoll = null;

function addMessageToInbox(message) {
    const sentByMe = me == message.sender;

    const messageDiv = document.createElement("div");
    messageDiv.id = getMessageDivId(message.id);
    messageDiv.className = sentByMe ? "mine" : "others";
    messageDiv.className += " message";

    if (!sentByMe) {
        const senderP = document.createElement("p");
        senderP.className = "sender";
        senderP.innerHTML = message.sender;
        messageDiv.appendChild(senderP);
    }

    const messageTextInput = document.createElement("input");
    messageTextInput.type = "text";
    messageTextInput.value = message.text;
    messageDiv.appendChild(messageTextInput);

    if (sentByMe) {
        const editButton = document.createElement("input");
        editButton.type = "button";
        editButton.value = "Edit";
        editButton.onclick = () => {
            message.text = messageTextInput.value;
            liveGoll.update(message);
        };
        messageDiv.appendChild(editButton);

        const deleteButton = document.createElement("input");
        deleteButton.type = "button";
        deleteButton.value = "Delete";
        deleteButton.onclick = () => {
            liveGoll.delete(message);
        };
        messageDiv.appendChild(deleteButton);
    }

    const sentTimeP = document.createElement("p");
    sentTimeP.className = "time";
    sentTimeP.innerHTML = new Date(message.sentTime).toLocaleTimeString();
    messageDiv.appendChild(sentTimeP);

    inboxDiv.appendChild(messageDiv);
}

window.onload = () => {
    liveGoll = new LiveGollection("ws://localhost:8080/livegollection");

    const messageTextInput = document.getElementById("message-text-input");
    const sendButton = document.getElementById("send-button");

    sendButton.onclick = () => {
        liveGoll.create({
            sender: me,
            sentTime: new Date(),
            text: messageTextInput.value,
        });
    };

    inboxDiv = document.getElementById("inbox-div");

    liveGoll.oncreate = (message) => {
        addMessageToInbox(message, inboxDiv);
    };

    liveGoll.onupdate = (message) => {
        const messageToUpdateTextInput = document.getElementById(getMessageDivId(message.id))
            .getElementsByTagName('input')[0];
        messageToUpdateTextInput.value = message.text;
    };

    liveGoll.ondelete = (message) => {
        const messageToDeleteDiv = document.getElementById(getMessageDivId(message.id));
        messageToDeleteDiv.remove();
    };
};
```
**livegollection-client** is as simple to use as **livegollection**: we need to create a LiveGollection object passing to the constructor the url to the **livegollection** server-side handler (in this case ws://localhost:8080/livegollection).

After the LiveGollection object has been created we need to set the event handlers **oncreate**, **onupdate** and **ondelete** to handle the correspondant events relative to messages. In this case to handle the events we need to update the DOM by adding a div for the new message or by modifying/deleting the existing one.

We can send events to the server with the following methods: **create**, **update**, **delete**. In this case we use **create** to send a new message and **update**/**delete** to edit/delete an existing one sent by us.
## Run the app!
To run the app we need to bundle our script with webpack and start the server:
```bash
cd .. # (If you are inside src directory)
npx webpack --mode production --entry ./src/index.js --output-filename bundle.js -o ./static
cd server
go run main.go
```
