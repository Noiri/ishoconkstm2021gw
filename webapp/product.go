package main

import (
	"strconv"
)


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


// Product Model
type Product struct {
	ID          int
	Name        string
	Description string
	ImagePath   string
	Price       int
	CreatedAt   string
}

// ProductWithComments Model
type ProductWithComments struct {
	ID           int
	Name         string
	Description  string
	ImagePath    string
	Price        int
	CreatedAt    string
	CommentCount int
	Comments     []CommentWriter
}

// CommentWriter Model
type CommentWriter struct {
	Content string
	Writer  string
}


func getProduct(pid int) Product {
	return productMemoTable[pid-1]
}


// ユーザの購入金額を取得する.
func getBuyingSum(uid int) string {
	sum_user := "sum_user" + strconv.Itoa(uid)
	sum, err := rdb.Get(ctx, sum_user).Result()
	if err != nil {
		panic(err)
	}

	return sum
}


func getProductsWithCommentsAt(page int) [50]ProductWithComments {
	products := [50]ProductWithComments{}
	start_idx := (199 - page) * 50
	for i := 0; i < 50; i++ {
			pid_string := strconv.Itoa(start_idx + i + 1)
			result, err := rdb.Get(ctx, "count_" + pid_string).Result()
			if err != nil {
					panic(err)
			}

			cnt, err := strconv.Atoi(result)
			productsWithComments[start_idx + i].CommentCount = cnt


			products[50 - i - 1] = productsWithComments[start_idx + i]
	}

	return products
}


// 
func (p *Product) isBought(uid int) bool {
	//集計処理が重いのでクエリを変更.

	// logout時にクエリ投げられたときのエスケープ
	if uid == 0 {
		return false
	}

	//fmt.Println(uid)
	/*
	if isBoughtMemoTable[uid-1][p.ID-1] {
		return true
	} else {
		var count int
		err := db.QueryRow(
			"SELECT * as count FROM histories WHERE product_id = ? AND user_id = ? LIMIT 1",
			p.ID, uid,
		).Scan(&count)

		isBoughtMemoTable[uid-1][p.ID-1] = true

		return err == nil
	}
	*/

	return isBoughtMemoTable[uid-1][p.ID-1]
	
	/*
	var count int
	err := db.QueryRow(
		"SELECT * as count FROM histories WHERE product_id = ? AND user_id = ? LIMIT 1",
		p.ID, uid,
	).Scan(&count)
	//isBoughtMemoTable[uid-1][p.ID-1] = true

	return err == nil
	*/
}
