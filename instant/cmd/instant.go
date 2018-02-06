// Sample instant demonstrates how to run a simple instant answers server.
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/jivesearch/jivesearch/config"
	"github.com/jivesearch/jivesearch/instant"
	"github.com/jivesearch/jivesearch/wikipedia"
	"github.com/spf13/viper"
)

type cfg struct {
	*instant.Instant
}

func (c *cfg) handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sol := c.Instant.Detect(r)

	if err := json.NewEncoder(w).Encode(sol); err != nil {
		http.Error(w, http.StatusText(500), 500)
	}
}

func favHandler(w http.ResponseWriter, r *http.Request) {}

func setup() (*sql.DB, error) {
	v := viper.New()
	v.SetEnvPrefix("jivesearch")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	config.SetDefaults(v)

	db, err := sql.Open("postgres",
		fmt.Sprintf(
			"user=%s password=%s host=%s database=%s sslmode=require",
			v.GetString("postgresql.user"),
			v.GetString("postgresql.password"),
			v.GetString("postgresql.host"),
			v.GetString("postgresql.database"),
		),
	)
	if err != nil {
		panic(err)
	}

	db.SetMaxIdleConns(0)

	return db, err
}

func main() {
	db, err := setup()
	if err != nil {
		panic(err)
	}

	defer db.Close()

	c := &cfg{
		&instant.Instant{
			QueryVar: "q",
			Fetcher: &wikipedia.PostgreSQL{
				DB: db,
			},
		},
	}

	port := 8000
	http.HandleFunc("/", c.handler)
	http.HandleFunc("/favicon.ico", favHandler)
	log.Printf("Listening at http://localhost:%d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
