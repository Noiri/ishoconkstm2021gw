package main

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
	"unicode/utf8"
)

// productをメモリに載せる.
//productMemoTableに値を割り当て.
func initProduct() [10000]Product {
	products := [10000]Product{}
	for pid := 1; pid <= 10000; pid++ {
		p := Product{}
		row := db.QueryRow("SELECT * FROM products WHERE id = ? LIMIT 1", pid)
		err := row.Scan(&p.ID, &p.Name, &p.Description, &p.ImagePath, &p.Price, &p.CreatedAt)
		if err != nil {
			panic(err.Error())
		}
		products[pid-1] = p
	}
	return products
}



func initProductsWithCommentsAt() [10000]ProductWithComments {
	products := [10000]ProductWithComments{}

	// read json
	raw, err := ioutil.ReadFile("./initProductWithComments.json")
	if err != nil {
		panic(err)
	}
	json_pwc := []JsonProductWithComments{}
	json.Unmarshal(raw, &json_pwc)

	for pid := 0; pid < 10000; pid++ {
		products[pid].ID = json_pwc[pid].ID
		products[pid].Name = json_pwc[pid].Name
		products[pid].Description = json_pwc[pid].Description
		products[pid].ImagePath = json_pwc[pid].ImagePath
		products[pid].Price = json_pwc[pid].Price
		products[pid].CreatedAt = json_pwc[pid].CreatedAt
		products[pid].CommentCount = json_pwc[pid].CommentCount
		products[pid].Comments = json_pwc[pid].Comments

		// shorten description and comment
		if utf8.RuneCountInString(products[pid].Description) > 70 {
			products[pid].Description = string([]rune(products[pid].Description)[:70]) + "…"
		}

		var newCW []CommentWriter
		for _, c := range products[pid].Comments {
			if utf8.RuneCountInString(c.Content) > 25 {
				c.Content = string([]rune(c.Content)[:25]) + "…"
			}
			newCW = append(newCW, c)
		}
		products[pid].Comments = newCW
	}

	return products
}


//ユーザごとの総購入金額をredisにのせる.
func initBuyingSum() {
	// read json
	raw, err := ioutil.ReadFile("./initBuySum.json")
	if err != nil {
		panic(err)
	}
	json_buysum := [5000]int{}
	json.Unmarshal(raw, &json_buysum)

	// redis set
	for uid := 1; uid <= 5000; uid++ {
		sum_user := "sum_user" + strconv.Itoa(uid)
		aerr := rdb.Set(ctx, sum_user, json_buysum[uid-1], 0)
		if aerr == nil {
			panic(aerr)
		}
	}
}




//redis: uid+++pid -> bool(int)
func initIsBought() [5000][10000]bool {
	// read json
	raw, err := ioutil.ReadFile("./initIsBought.json")
	if err != nil {
		panic(err)
	}
	json_isBought := [5000][10000]bool{}
	json.Unmarshal(raw, &json_isBought)

	return json_isBought
}



func initAuthenticateMap() map[string]PassAndUid {

	Email2Pass := map[string]PassAndUid{}

	for uid := 1; uid <= 5000; uid++ {
		u := User{}
		err := db.QueryRow("SELECT * FROM users WHERE id = ? LIMIT 1", uid).Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.LastLogin)
		if err != nil {
			panic(err)
		}

		Email2Pass[u.Email] = PassAndUid{u.Password, uid}
	}

	return Email2Pass
}




func initUsers() [5000]User {
	users := [5000]User{}
	for uid := 1; uid <= 5000; uid++ {
		u := User{}
		r := db.QueryRow("SELECT * FROM users WHERE id = ? LIMIT 1", uid)
		err := r.Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.LastLogin)
		if err != nil {
			panic(err)
		}
		users[uid-1] = u
	}

	return users
}



//comment-countをredisにのせる
func initCommentsCount() {
	for i := 1; i<=10000; i++ {
		pid_string := strconv.Itoa(i)
		aerr := rdb.Set(ctx, "count_" + pid_string, 20, 0)
		if aerr == nil {
				panic(aerr)
		}
	}

}
