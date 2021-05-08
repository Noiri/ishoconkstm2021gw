package main

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/contrib/sessions"
)

// User model
type User struct {
	ID        int
	Name      string
	Email     string
	Password  string
	LastLogin string
}


// for auth
type PassAndUid struct {
	PassWord string
	Uid int
}

func authenticate(email string, password string) (User, bool) {
	result := (authenticateMap[email].PassWord == password)
	user := User{}
	if result {
		user = userMemoTable[authenticateMap[email].Uid - 1]
	}
	return user, result
}

func notAuthenticated(session sessions.Session) bool {
	uid := session.Get("uid")
	return !(uid.(int) > 0)
}


func getUser(uid int) User {
	return userMemoTable[uid-1]
}

func currentUser(session sessions.Session) User {

	u := User{}

	uID := session.Get("uid")
	uName := session.Get("uName")

	if uID != nil && uName != nil {
		u.ID = session.Get("uid").(int)
		u.Name = session.Get("uName").(string)
	}

	return u
}


func initBuyingHistoriy() [5000][]Product {
	// read json
	raw, err := ioutil.ReadFile("./initBuyingHistory.json")
	if err != nil {
		panic(err)
	}
	buyhist := [5000][]Product{}
	json.Unmarshal(raw, &buyhist)

	// 30件までしか使用しない.
	for uid := 1; uid <= 5000; uid++ {
		if len(buyhist[uid-1]) > 30 {
			buyhist[uid-1] = buyhist[uid-1][0:30]
		}

		for i, p := range buyhist[uid-1] {
			// shorten description and comment
			if utf8.RuneCountInString(p.Description) > 70 {
				p.Description = string([]rune(p.Description)[:70]) + "…"
				buyhist[uid-1][i] = p
			}
		}
	}

	return buyhist
}


// BuyingHistory : products which user had bought
func (u *User) BuyingHistory() ([]Product, int) {
	buyingHistoryLock.RLock()
	userBuyingHistory := buyingHistoryCache[u.ID-1]
	totalPay, ok := userTotalPay.Load(u.ID)
	//sum := userSumCache[u.ID-1]
	buyingHistoryLock.RUnlock()

	if !ok {
		totalPay = 0
	}
	

	return userBuyingHistory, totalPay.(int)
}


// BuyProduct : buy product
func (u *User) BuyProduct(pid string) {
	pid_int, err := strconv.Atoi(pid)
	if err != nil {
		panic(err)
	}

	buyingHistoryLock.Lock()
	{	
		// update total buy
		product := getProduct(pid_int)
		totalPay, ok := userTotalPay.Load(u.ID)
		if !ok {}
		userTotalPay.Store(u.ID, totalPay.(int) + product.Price)


		latest_p := []Product{productMemoTable[pid_int - 1]}
		latest_p[0].CreatedAt = (time.Now()).Format("2006-01-02 15:04:05")
		// shorten description and comment
		if utf8.RuneCountInString(latest_p[0].Description) > 70 {
			latest_p[0].Description = string([]rune(latest_p[0].Description)[:70]) + "…"
		}

		// 30件プールされていたら最後尾(最古)をpop
		if len(buyingHistoryCache[u.ID-1]) > 30 {
			buyingHistoryCache[u.ID-1] = buyingHistoryCache[u.ID-1][0:29]
		} 

		// 一番前にpush
		latest_p = append(latest_p, buyingHistoryCache[u.ID-1]...)
		buyingHistoryCache[u.ID-1] = latest_p
		
		isBoughtMemoTable[u.ID-1][pid_int - 1] = true

	}
	buyingHistoryLock.Unlock()


	// ここはベンチマーカーがDBを利用してるらしいので消せない.
	db.Exec( "INSERT INTO histories (product_id, user_id, created_at) VALUES (?, ?, ?)", pid, u.ID, time.Now())
}


// CreateComment : create comment to the product
func (u *User) CreateComment(pid string, content string) {
	//getCommentがいらないので、コメント件数にインクリメントするだけ.
	pid_int, e := strconv.Atoi(pid)
	if e != nil {
		panic(e)
	}
	cnt, ok := commentCount.Load(pid_int)
	if ok {}
	commentCount.Store(pid_int, cnt.(int) + 1)

	productsWithComments[pid_int - 1].CommentCount += 1
}
