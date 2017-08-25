package main

import (
	"html/template"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/user"
)

// Piece is a single item in a delivery
type Piece struct {
	AddedBy     string
	Description string
	Quantity    int64
	Date        time.Time
}

// pieceKey returns the key used for all piece entries.
func pieceKey(c context.Context) *datastore.Key {
	// The string "default_delivery" here could be varied to have multiple guestbooks.
	return datastore.NewKey(c, "Piece", "default_piece", 0, nil)
}

func init() {
	http.HandleFunc("/", root)
	http.HandleFunc("/add", add)
}

func root(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	// Ancestor queries, as shown here, are strongly consistent with the High
	// Replication Datastore. Queries that span entity groups are eventually
	// consistent. If we omitted the .Ancestor from this query there would be
	// a slight chance that Greeting that had just been written would not
	// show up in a query.
	// [START query]
	q := datastore.NewQuery("Piece").Ancestor(pieceKey(c)).Order("-Date").Limit(10)
	// [END query]
	// [START getall]
	pieces := make([]Piece, 0, 10)
	if _, err := q.GetAll(c, &pieces); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// [END getall]
	inventoryTemplate, _ := template.ParseFiles("tmpl/index.html")
	if err := inventoryTemplate.Execute(w, pieces); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func add(w http.ResponseWriter, r *http.Request) {
	// [START new_context]
	c := appengine.NewContext(r)
	p := Piece{
		Description: r.FormValue("description"),
		Date:        time.Now(),
	}

	qty, err := strconv.ParseInt(r.FormValue("quantity"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	p.Quantity = qty

	// [START if_user]
	if u := user.Current(c); u != nil {
		p.AddedBy = u.String()
	}
	// We set the same parent key on every Greeting entity to ensure each Greeting
	// is in the same entity group. Queries across the single entity group
	// will be consistent. However, the write rate to a single entity group
	// should be limited to ~1/second.
	key := datastore.NewIncompleteKey(c, "Piece", pieceKey(c))
	_, err = datastore.Put(c, key, &p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
	// [END if_user]
}
