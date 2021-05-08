package main

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
)



func hedder(cuser User, content string) []byte {
	var buf bytes.Buffer
	buf.Grow(0x10000)

	io.WriteString(&buf, `<!DOCTYPE html><html><head><meta http-equiv="Content-Type" content="text/html" charset="utf-8"><link rel="stylesheet" href="/css/bootstrap.min.css">
	<title>すごいECサイト</title></head><body><nav class="navbar navbar-inverse navbar-fixed-top"><div class="container"><div class="navbar-header">
	<a class="navbar-brand" href="/">すごいECサイトで爆買いしよう!</a></div><div class="header clearfix">`)
	if cuser.ID > 0 {
		io.WriteString(&buf, `<nav><ul class="nav nav-pills pull-right"><li role="presentation"><a href="/users/` + strconv.Itoa(cuser.ID) + `">` +
				cuser.Name + `さんの購入履歴</a></li><li role="presentation"><a href="/logout">Logout</a></li></ul></nav>`)
	} else {
		io.WriteString(&buf, `<nav><ul class="nav nav-pills pull-right"><li role="presentation"><a href="/login">Login</a></li></ul></nav>`)
	}
	io.WriteString(&buf, `</div></nav>`)

	//{{ template "content" . }}
	io.WriteString(&buf, content)

	io.WriteString(&buf, `</body></html>`)


	byte_html := buf.Bytes()

	return byte_html
}


func mypage_html(cuser User, user User, totalPay int, products []Product) string {
	var html string

	html = fmt.Sprintf(`<div class="jumbotron"><div class="container"><h2>%s さんの購入履歴</h2><h4>合計金額: %d円</h4></div></div><div class="container"><div class="row">`, user.Name, totalPay)
	
	for i, p := range products {
		if i >= 30 {
			break
		}

		s := fmt.Sprintf(`<div class="col-md-4"><div class="panel panel-default"><div class="panel-heading"><a href="/products/%d">%s</a></div>
		<div class="panel-body"><a href="/products/%d"><img src="%s" class="img-responsive" /></a>
		<h4>価格</h4><p>%d円</p><h4>商品説明</h4><p>%s</p><h4>購入日時</h4><p>%s</p></div>`, p.ID, p.Name, p.ID, p.ImagePath, p.Price, p.Description, p.CreatedAt)					
						
		if user.ID == cuser.ID{
			s += fmt.Sprintf(`<div class="panel-footer"><form method="POST" action="/comments/%d"><fieldset><div class="form-group"><input class="form-control" placeholder="Comment Here" name="content" value="">
							</div><input class="btn btn-success btn-block" type="submit" name="send_comment" value="コメントを送信" /></fieldset></form></div>`, p.ID)
		}
		s += `</div></div>`
		html += s
	}

	html += `</div></div>`

	return html
}


// original = cuser.ID > 0
func index_html(isAuth bool,products [50]ProductWithComments) string {
	html := `<div class="jumbotron"><div class="container"><h1>今日は大安売りの日です！</h1></div></div><div class="container"><div class="row">`

	for _, p := range products {
		pid := strconv.Itoa(p.ID)

		html += `<div class="col-md-4"><div class="panel panel-default"><div class="panel-heading"><a href="/products/`+ pid + `">` + p.Name + `</a></div>` +
			`<div class="panel-body"><a href="/products/` + pid + `"><img src="` + p.ImagePath + `" class="img-responsive" /></a><h4>価格</h4><p>` + 
			strconv.Itoa(p.Price) + `円</p><h4>商品説明</h4><p>` + p.Description + `</p><h4>` + strconv.Itoa(p.CommentCount) + `件のレビュー</h4><ul>`

		for _, cw := range p.Comments {
			html += (`<li>` + cw.Content + ` by ` + cw.Writer + `</li>`)
		}
		html += `</ul></div>`
		if isAuth {
			html += (`<div class="panel-footer"><form method="POST" action="/products/buy/` + pid + `"><fieldset>
				<input class="btn btn-success btn-block" type="submit" name="buy" value="購入" /></fieldset></form></div>`)
		}
		html += `</div></div>`
	}
	html += `</div></div>`

	return html
}


func product_html(cuser User, product Product, isBought bool) string {
	html := `<div class="jumbotron"><div class="container"><h2>`+ product.Name + `</h2>`
	if isBought {
		html += `<h4>あなたはすでにこの商品を買っています</h4>`
	}
	html += (`</div></div><div class="container"><div class="row"><div class="jumbotron"><img src="` + product.ImagePath + `" class="img-responsive" width="400"/><h2>価格</h2><p>` + strconv.Itoa(product.Price) + ` 円</p><h2>商品説明</h2><p>` + product.Description + `</p></div></div></div>`)

	return html
}