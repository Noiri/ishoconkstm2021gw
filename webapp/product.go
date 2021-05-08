package main


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


func getProductsWithCommentsAt(page int) [50]ProductWithComments {
	products := [50]ProductWithComments{}
	start_idx := (199 - page) * 50
	for i := 0; i < 50; i++ {
			tmp, ok := commentCount.Load(start_idx + i + 1)
			if ok {
				productsWithComments[start_idx + i].CommentCount = tmp.(int)
			}
			
			products[50 - i - 1] = productsWithComments[start_idx + i]
	}

	return products
}


func (p *Product) isBought(uid int) bool {
	// logout時にクエリ投げられたときのエラーハンドリング
	if uid == 0 {
		return false
	}

	buyingHistoryLock.Lock()
	res := isBoughtMemoTable[uid-1][p.ID-1]
	buyingHistoryLock.Unlock()

	return res
}