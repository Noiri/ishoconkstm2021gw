package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

type JsonProductWithCommentsCllection struct {
        cllection []JsonProductWithComments `json:JsonProductWithComments`
}

type  JsonProductWithComments struct {
        ID           int `json:ID`
	Name         string `json:Name`
	Description  string `json:Description`
	ImagePath    string `json:ImagePath`
	Price        int `json: Price`
	CreatedAt    string `json:CreatedAt`
	CommentCount int `json:CommentCount`
	Comments     []CommentWriter `json:Comments`
}


//mysql
var db *sql.DB

func getEnv(key, fallback string) string {
        if value, ok := os.LookupEnv(key); ok {
                return value
        }
        return fallback
}

func main() {
        // database setting
        user := getEnv("ISHOCON1_DB_USER", "ishocon")
        pass := getEnv("ISHOCON1_DB_PASSWORD", "ishocon")
        dbname := getEnv("ISHOCON1_DB_NAME", "ishocon1")
        db, _ = sql.Open("mysql", user+":"+pass+"@/"+dbname)
        db.SetMaxIdleConns(5)


        jsonComments := [10000][]Comment{}

        for pid := 1; pid <= 10000; pid++ {
		comments := []Comment{}
		rows, err := db.Query("SELECT * FROM comments WHERE product_id = ? ", pid)
		if err != nil {
			panic(err)
		}

		defer rows.Close()
		for rows.Next() {
			c := Comment{}
			err = rows.Scan(&c.ID, &c.ProductID, &c.UserID, &c.Content, &c.CreatedAt)
			comments = append(comments, c)
		}

                fmt.Println(pid)

                jsonComments[pid-1] = comments
	}

        // json file に出力.
        f, err := os.Create("initComment.json")
        if err != nil {
                panic(err)
        }
        defer f.Close()

        err = json.NewEncoder(f).Encode(jsonComments)
        if err != nil {
                panic(err)
        }
        



        // read json
        /*
        raw, err := ioutil.ReadFile("./initProductWithComments.json")
        if err != nil {
                fmt.Println(err.Error())
                os.Exit(1)
        }

        tmp := []JsonProductWithComments{}

        json.Unmarshal(raw, &tmp)

        fmt.Println(tmp[0].Name)
        */
}