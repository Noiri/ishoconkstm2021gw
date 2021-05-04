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


        isBought := [5000][10000]bool{}

        for uid := 1; uid <= 5000; uid++ {
                fmt.Println(uid)
                for pid := 1; pid <= 10000; pid++ {
                        var count int
                        err := db.QueryRow(
                                "SELECT count(*) as count FROM histories WHERE product_id = ? AND user_id = ?", pid, uid).Scan(&count)
                        if err != nil {
                                panic(err.Error())
                        }

                        isBought[uid-1][pid-1] = (count > 0)
                }
        }

        // jsonエンコード
        
        f, err := os.Create("initIsBought.json")
        if err != nil {
                panic(err)
        }
        defer f.Close()

        err = json.NewEncoder(f).Encode(isBought)
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