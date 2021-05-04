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


        jsonProductWithCommentsCllection := []JsonProductWithComments{}

        for pid := 1; pid <= 10000; pid++ {
                p := JsonProductWithComments{}
                row := db.QueryRow("SELECT * FROM products WHERE id = ? LIMIT 1", pid)
                err := row.Scan(&p.ID, &p.Name, &p.Description, &p.ImagePath, &p.Price, &p.CreatedAt)
                if err != nil {
                        panic(err.Error())
                }

                // select comment count for the product
		var cnt int
		cnterr := db.QueryRow("SELECT count(*) as count FROM comments WHERE product_id = ?", p.ID).Scan(&cnt)
		if cnterr != nil {
			cnt = 0
		}
		p.CommentCount = cnt


                if cnt > 0 {
			// select 5 comments and its writer for the product
			var cWriters []CommentWriter

			subrows, suberr := db.Query("SELECT * FROM comments as c INNER JOIN users as u "+
				"ON c.user_id = u.id WHERE c.product_id = ? ORDER BY c.created_at DESC LIMIT 5", p.ID)
			if suberr != nil {
				subrows = nil
			}

			defer subrows.Close()
			for subrows.Next() {
				var i int
				var s string
				var cw CommentWriter
				subrows.Scan(&i, &i, &i, &cw.Content, &s, &i, &cw.Writer, &s, &s, &s)
				cWriters = append(cWriters, cw)
			}

			p.Comments = cWriters
		}

                fmt.Println(pid)

                jsonProductWithCommentsCllection = append(jsonProductWithCommentsCllection, p)
        }

        //fmt.Println(jsonProductWithCommentsCllection.cllection)

        // jsonエンコード
        
        f, err := os.Create("initProductWithComments.json")
        if err != nil {
                panic(err)
        }
        defer f.Close()

        err = json.NewEncoder(f).Encode(jsonProductWithCommentsCllection)
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