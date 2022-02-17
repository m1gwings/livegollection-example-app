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
