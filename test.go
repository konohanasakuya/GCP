package test

import (
	"net/http"
	"fmt"
	"html"
)

var (
	Data1  = "データ１"
	Data2  = "データ２"
	Data3  = "データ３"
	Data4  = "データ４"
	Data5  = "データ５"
	Data6  = "データ６"
	Data7  = "データ７"
	Data8  = "データ８"
	Data9  = "データ９"
)

func viewHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("run Handler")

	// テキストフォームに入力されたデータを引っこ抜き、入力値：XXXとして表示
    v := r.FormValue("input_value")

	// ここでDataを更新すると書き換わる
	// また、<tr>の部分のデータ量に合わせて繰り返しで生成するようにすればいい？

	// テスト(おまけ)
	if v != "" {Data1 = v; Data2 = v; Data3 = v}

	// HTML
	h := `
<html><head><title>Go表示ページ</title></head><body>
    <h1>ページ見出し</h1>
    <form  action="/" method="post">
        <input type="text" name="input_value">
        <input type="submit" name="送信"><br>
        入力値：`+html.EscapeString(v)+`
    </form>

    <table border="1" bordercolor=#9999FF cellspacing="0">
        <caption>データ</caption>
        <tr bgcolor=#CCCCFF>
            <th width="200">あああ</th>
            <th width="200">いいい</th>
            <th width="200">ううう</th>
        </tr>
        <tr>
            <td>`+html.EscapeString(Data1)+`</td>
            <td>`+html.EscapeString(Data2)+`</td>
            <td>`+html.EscapeString(Data3)+`</td>
        </tr>
        <tr>
            <td>`+html.EscapeString(Data4)+`</td>
            <td>`+html.EscapeString(Data5)+`</td>
            <td>`+html.EscapeString(Data6)+`</td>
        </tr>
        <tr>
            <td>`+html.EscapeString(Data7)+`</td>
            <td>`+html.EscapeString(Data8)+`</td>
            <td>`+html.EscapeString(Data9)+`</td>
        </tr>
    </table>
</body></html>
`
	// 表示更新
	fmt.Fprint(w, h)
}


// App Engine上では init 関数に変える必要あり
func init() {
	// HTTPハンドラ登録 (ページにアクセルされると走る)
	http.HandleFunc("/", viewHandler)

	// HTTPサーバ起動  (★★ http://localhost:8080/で表示される ★★)
	fmt.Println("http server start")
	http.ListenAndServe(":8080", nil)
}

