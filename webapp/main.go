package main

import (
	"database/sql"
	"html/template"
	"net/http"
	"os"
	"strconv"

	"sync"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

//mysql
var db *sql.DB


// Lock
var buyingHistoryLock sync.RWMutex

// Memo Tabels (imutable, read only)
var productsWithComments [10000]ProductWithComments // 最新五件を保存するテーブル
var productMemoTable [10000]Product // productの情報をpid -> productで保存するテーブル
var userMemoTable [5000]User //userの情報を uid -> user で保存するテーブル
var authenticateMap map[string]PassAndUid // auth用, email -> password,uid


// Cache (mutable)
var isBoughtMemoTable [5000][10000]bool
var buyingHistoryCache[5000][]Product
var userTotalPay sync.Map
var commentCount sync.Map // comment count


//var IndexHtml [200][2]string

var layoutPageHtml [5001]string

var isMyPageCached [5000][2]bool
var myPageHtml [5000][2]string


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
        

        r.GET("/users/:userId", func(c *gin.Context) {
                cUser := currentUser(sessions.Default(c))
                uid, _ := strconv.Atoi(c.Param("userId"))
                user := getUser(uid)

                myPage := user.BuyingHistory(cUser)

                html := hedder(cUser, myPage)
                c.Data(http.StatusOK, "text/html", html)
        })


        r.GET("/products/:productId", func(c *gin.Context) {
                //getCommentいらない.
                pid, _ := strconv.Atoi(c.Param("productId"))
                product := getProduct(pid)

                cUser := currentUser(sessions.Default(c))
                isBought := product.isBought(cUser.ID)

                html := hedder(cUser, product_html(cUser, product, isBought))
                c.Data(http.StatusOK, "text/html", html)
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

                buyingHistoryCache = initBuyingHistoriy()


                // NavigationBarを先に作っておく.
                for uid := 1; uid <= 5000; uid++ {
                        user := userMemoTable[uid-1]
                        layoutPageHtml[uid] = layout_html(user)
                }
                layoutPageHtml[0] = layout_html(User{})


                c.String(http.StatusOK, "Finish")
        })

        r.RunUnix("/run/go/webapp.sock")
}