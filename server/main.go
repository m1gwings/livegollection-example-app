package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/m1gwings/livegollection"
	"github.com/m1gwings/livegollection-example-app/chat"
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
