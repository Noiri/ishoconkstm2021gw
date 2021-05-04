package main

import (
	"database/sql"
	"html/template"
	"net/http"
	"os"
	"strconv"

	"context"

	"github.com/go-redis/redis/v8"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

//mysql
var db *sql.DB

//redis
var rdb *redis.Client
var ctx context.Context


// Memo Tabels
var productsWithComments [10000]ProductWithComments // 最新五件を保存するテーブル
var productMemoTable [10000]Product // productの情報をpid -> productで保存するテーブル
var userMemoTable [5000]User //userの情報を uid -> user で保存するテーブル
var authenticateMap map[string]PassAndUid // auth用, email -> password,uid

var isBoughtMemoTable [5000][10000]bool


func getEnv(key, fallback string) string {
        if value, ok := os.LookupEnv(key); ok {
                return value
        }
        return fallback
}

func main() {
        //redis setting
        ctx = context.Background()
        rdb = redis.NewClient(&redis.Options{
                Network:  "unix",
                Addr:     "/var/run/redis/redis-server.sock",
                Password: "",
                DB:       0,
        })


        // database setting
        user := getEnv("ISHOCON1_DB_USER", "ishocon")
        pass := getEnv("ISHOCON1_DB_PASSWORD", "ishocon")
        dbname := getEnv("ISHOCON1_DB_NAME", "ishocon1")
        //db, _ = sql.Open("mysql", user+":"+pass+"@/"+dbname)
        db, _ = sql.Open("mysql", user+":"+pass+"@unix(/var/run/mysqld/mysqld.sock)/"+dbname)
        db.SetMaxIdleConns(5)

        r := gin.Default()
        // load templates
        r.Use(static.Serve("/css", static.LocalFile("public/css", true)))
        r.Use(static.Serve("/images", static.LocalFile("public/images", true)))
        layout := "templates/layout.tmpl"

        // session store
        store := sessions.NewCookieStore([]byte("mysession"))
        store.Options(sessions.Options{HttpOnly: true})
        r.Use(sessions.Sessions("showwin_happy", store))


        // GET /login
        r.GET("/login", func(c *gin.Context) {
                session := sessions.Default(c)
                session.Clear()
                session.Save()

                tmpl, _ := template.ParseFiles("templates/login.tmpl")
                r.SetHTMLTemplate(tmpl)
                c.HTML(http.StatusOK, "login", gin.H{
                        "Message": "ECサイトで爆買いしよう！！！！",
                })
        })

        // POST /login
        r.POST("/login", func(c *gin.Context) {
                email := c.PostForm("email")
                pass := c.PostForm("password")

                session := sessions.Default(c)
                user, result := authenticate(email, pass)
                if result {
                        // 認証成功
                        session.Set("uid", user.ID)
                        session.Set("uName", user.Name)

                        session.Save()
                        c.Redirect(http.StatusSeeOther, "/")
                } else {
                        // 認証失敗
                        tmpl, _ := template.ParseFiles("templates/login.tmpl")
                        r.SetHTMLTemplate(tmpl)
                        c.HTML(http.StatusOK, "login", gin.H{
                                "Message": "ログインに失敗しました",
                        })
                }
        })


        // GET /
        r.GET("/", func(c *gin.Context) {
                cUser := currentUser(sessions.Default(c))

                page, err := strconv.Atoi(c.Query("page"))
                if err != nil {
                        page = 0
                }
                products := getProductsWithCommentsAt(page)

                //sProducts
                r.SetHTMLTemplate(template.Must(template.ParseFiles(layout, "templates/index.tmpl")))
                c.HTML(http.StatusOK, "base", gin.H{
                        "CurrentUser": cUser,
                        "Products":    products,
                })
        })

        // GET /users/:userId
        r.GET("/users/:userId", func(c *gin.Context) {
                cUser := currentUser(sessions.Default(c))

                uid, _ := strconv.Atoi(c.Param("userId"))
                user := getUser(uid)

                // shorten description
                //左から70文字とるだけなので、sqlでやってしまう.
                sdProducts := user.BuyingHistory()
                totalPay := getBuyingSum(uid)

                r.SetHTMLTemplate(template.Must(template.ParseFiles(layout, "templates/mypage.tmpl")))
                c.HTML(http.StatusOK, "base", gin.H{
                        "CurrentUser": cUser,
                        "User":        user,
                        "Products":    sdProducts,
                        "TotalPay":    totalPay,
                })
        })

        // GET /products/:productId
        r.GET("/products/:productId", func(c *gin.Context) {
                //getCommentいらない.
                pid, _ := strconv.Atoi(c.Param("productId"))
                product := getProduct(pid)

                cUser := currentUser(sessions.Default(c))
                bought := product.isBought(cUser.ID)

                r.SetHTMLTemplate(template.Must(template.ParseFiles(layout, "templates/product.tmpl")))
                c.HTML(http.StatusOK, "base", gin.H{
                        "CurrentUser":   cUser,
                        "Product":       product,
                        "AlreadyBought": bought,
                })
        })

        // POST /products/buy/:productId
        r.POST("/products/buy/:productId", func(c *gin.Context) {
                // need authenticated
                if notAuthenticated(sessions.Default(c)) {
                        tmpl, _ := template.ParseFiles("templates/login.tmpl")
                        r.SetHTMLTemplate(tmpl)
                        c.HTML(http.StatusForbidden, "login", gin.H{
                                "Message": "先にログインをしてください",
                        })
                } else {
                        // buy product
                        cUser := currentUser(sessions.Default(c))
                        cUser.BuyProduct(c.Param("productId"))

                        // redirect to user page
                        tmpl, _ := template.ParseFiles("templates/mypage.tmpl")
                        r.SetHTMLTemplate(tmpl)
                        c.Redirect(http.StatusFound, "/users/"+strconv.Itoa(cUser.ID))
                }
        })

        // POST /comments/:productId
        r.POST("/comments/:productId", func(c *gin.Context) {
                // need authenticated
                if notAuthenticated(sessions.Default(c)) {
                        tmpl, _ := template.ParseFiles("templates/login.tmpl")
                        r.SetHTMLTemplate(tmpl)
                        c.HTML(http.StatusForbidden, "login", gin.H{
                                "Message": "先にログインをしてください",
                        })
                } else {
                        // create comment
                        cUser := currentUser(sessions.Default(c))
                        cUser.CreateComment(c.Param("productId"), c.PostForm("content"))

                        // redirect to user page
                        tmpl, _ := template.ParseFiles("templates/mypage.tmpl")
                        r.SetHTMLTemplate(tmpl)
                        c.Redirect(http.StatusFound, "/users/"+strconv.Itoa(cUser.ID))
                }
        })

        // GET /initialize
        r.GET("/initialize", func(c *gin.Context) {
                db.Exec("DELETE FROM users WHERE id > 5000")
                db.Exec("DELETE FROM products WHERE id > 10000")
                db.Exec("DELETE FROM comments WHERE id > 200000")
                db.Exec("DELETE FROM histories WHERE id > 500000")

                // redisで管理
                initCommentsCount()
                initBuyingSum()

                initIsBought()

                // MemoTableの初期化.
                productsWithComments = initProductsWithCommentsAt()
                productMemoTable = initProduct()
                userMemoTable = initUsers()
                authenticateMap = initAuthenticateMap()

                isBoughtMemoTable = initIsBought()

                c.String(http.StatusOK, "Finish")
        })

        r.Run(":8080")
}