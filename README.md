# livegollection-example-app
**livegollection-example-app** is a simple web-chat app that demonstrates how the Golang **[livegollection](https://github.com/m1gwings/livegollection)** library can be used for live data synchronization between multiple web clients and the server.
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