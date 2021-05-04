package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

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

type JsonProduct struct {
        ID          int `json:ID`
	Name        string `json:Name`
	Description string `json:Description`
	ImagePath   string `json:ImagePath`
	Price       int `json:Price`
	CreatedAt   string `json:CreatedAt`
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


        buyHists := [5000][]JsonProduct{}

        for uid := 1; uid <= 5000; uid++ {
                products := []JsonProduct{}

                //数字が小さいほど先に挿入されたことになるので、古い.
                //昇順に並べ替えることで、後ろに追加するだけで良いようにする.
                rows, err := db.Query(
                        "SELECT p.id, p.name, p.description, p.image_path, p.price, h.created_at "+
                                "FROM histories as h "+
                                "LEFT OUTER JOIN products as p "+
                                "ON h.product_id = p.id "+
                                "WHERE h.user_id = ? "+
                                "ORDER BY h.id ASC", uid)
                if err != nil {
                        panic(err)
                }
        
                defer rows.Close()
                for rows.Next() {
                        p := JsonProduct{}
                        var cAt string
                        fmt := "2006-01-02 15:04:05"
                        err = rows.Scan(&p.ID, &p.Name, &p.Description, &p.ImagePath, &p.Price, &cAt)
                        tmp, _ := time.Parse(fmt, cAt)
                        p.CreatedAt = (tmp.Add(9 * time.Hour)).Format(fmt)
                        if err != nil {
                                panic(err.Error())
                        }
                        products = append(products, p)
                }

                fmt.Println(uid)

                buyHists[uid-1] = products
        }



        //fmt.Println(jsonProductWithCommentsCllection.cllection)

        // jsonエンコード
        
        f, err := os.Create("initBuyingHistory.json")
        if err != nil {
                panic(err)
        }
        defer f.Close()

        err = json.NewEncoder(f).Encode(buyHists)
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