package main

import (
	"strconv"
	"time"

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


// BuyingHistory : products which user had bought
func (u *User) BuyingHistory() (products []Product) {
	// 30件だけ読み込めば短くする件数を抑えられる.
	rows, err := db.Query(
		"SELECT p.id, p.name, LEFT(p.description, 70), p.image_path, p.price, h.created_at "+
			"FROM histories as h "+
			"LEFT OUTER JOIN products as p "+
			"ON h.product_id = p.id "+
			"WHERE h.user_id = ? "+
			"ORDER BY h.id DESC LIMIT 30", u.ID)
	if err != nil {
		return nil
	}

	defer rows.Close()
	for rows.Next() {
		p := Product{}
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

	return
}


// BuyProduct : buy product
func (u *User) BuyProduct(pid string) {
	sum_user := "sum_user" + strconv.Itoa(u.ID)
	pid_int, err := strconv.Atoi(pid)
	if err != nil {
		panic(err)
	}

	product := getProduct(pid_int)
	incr_err := rdb.IncrBy(ctx, sum_user, int64(product.Price)).Err()
	if incr_err != nil {
		panic(incr_err)
	}

	db.Exec(
		"INSERT INTO histories (product_id, user_id, created_at) VALUES (?, ?, ?)",
		pid, u.ID, time.Now())
	

	isBoughtMemoTable[u.ID-1][pid_int - 1] = true
}


// CreateComment : create comment to the product
func (u *User) CreateComment(pid string, content string) {
	//getCommentがいらないので、コメント件数にインクリメントするだけ.

	//comment pid の値に1加算する.
	err := rdb.Incr(ctx, "count_"+pid).Err()
	if err != nil {
		panic(err)
	}

	pid_int, err := strconv.Atoi(pid)
	if err != nil {
			panic(err)
	}
	productsWithComments[pid_int - 1].CommentCount += 1
}
