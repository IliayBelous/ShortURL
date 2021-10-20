package main

import (
	"fmt"
	"strconv"
	"net/http"
	"database/sql"
	"github.com/speps/go-hashids/v2"
	_ "github.com/lib/pq"
)

const (
    host     = "localhost"
    port     = 5432
    user     = "postgres"
    password = "qwerty"
    dbname   = "myDB"
)

var fullURLFormTmpl = []byte(`
<html>
	<head>
		<style>
			@import url("//fonts.googleapis.com/css?family=Pacifico&text=Вжух");
			@import url("//fonts.googleapis.com/css?family=Roboto:700");
			@import url("//fonts.googleapis.com/css?family=Kaushan+Script");
			body {
				min-height: 450px;
				height: 100vh;
				margin: 0;
				background: radial-gradient(circle, #0077ea, #1f4f96, #1b2949, #000);
			}
			.layer {
				font: 85px/0.10 "Pacifico", "Kaushan Script", Futura, "Roboto", "Trebuchet MS", Helvetica, sans-serif;
				white-space: pre;
				text-align: center;
				margin-top: -20px;
				color: #FF1212;
				letter-spacing: -2px;
				text-shadow: 2px 2px 2px #ffff37, 3px 3px 3px #FF871F, 6px 6px 6px #33199c;				
			}
			.words {
				color: whitesmoke;
				font-size: 25px;
				font-style: italic;
				letter-spacing: 3px;
				margin-top: -15px;
				position: relative;
				text-shadow: 2px 0 2px #00366b, 5px 5px 5px #002951, 0 4px 4px #00366b;
			}
			.short {
				color: whitesmoke;
				font-size: 20px;
				font-style: italic;
				letter-spacing: 3px;
				position: relative;
				text-shadow: 2px 0 2px #750E41, 5px 5px 5px #340E75, 0 4px 4px #750E41;
			}
			input[type="text"] {
			   border: 3px solid #5321B6;
			   border-radius: 10px;
			}
			input[type="submit"] {
			   border: 3px solid #123F83;
			   border-radius: 10px;
			   background: #2169D4;
			   color: #FFFFFF;
			   font-size: 18px;
			   font-family: Futura;
			}			
			</style>
	</head>
	<body>
		<div style="text-align: center; margin-top: 30px; padding-top: 10px;">
			<img src="/data/img/gopher.png" />
		</div>	
		<div class="layer">
			<h2 style="z-index: 2; position: relative;">Вжух !</h2>
			<p class="words">и коротка&#x301 длинная строка</p>
		</div>
		<form action="/" method="post" style="display: block; text-align: center">
			<input type="text" name="fullURL">
			<input type="submit" value="Post">
		</form>
		<form action="/get" method="get" style="display: block; text-align: center">
			<input type="text" name="originURL">
			<input type="submit" value="Get">
		</form>		
	</body>
</html>
`)



func CheckError(err error) {
    if err != nil {
        panic(err)
    }
}


func main() {
	// connection string
    psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	// open database
    db, err := sql.Open("postgres", psqlconn)
    CheckError(err)
     
    defer db.Close()

	// check db
    err = db.Ping()
    CheckError(err)
 
    fmt.Println("Connected!")
	
	
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Write(fullURLFormTmpl)
			return
		}
		
		
		inputFullURL := r.FormValue("fullURL")
		
		w.Write(fullURLFormTmpl)
		

		hd := hashids.NewData()
		hd.Salt = inputFullURL
		hd.MinLength = 10
		h, _ := hashids.NewWithData(hd)
		token := func(some []byte) (result []int) {
			for _, char := range some {
				elem, _ := strconv.Atoi(string(char))
				result = append(result, elem)
			}
			return result
		}([]byte(inputFullURL))
		e, _ := h.Encode(token)
		
		shortURL := `https://127.0.0.1:8080/` + string(e[:10])
		
		// insert
		// как избавиться от дубликатов???
		insert := `insert into "urls"("fullURL", "shortURL") values($1, $2) on conflict do nothing`
		db.Exec(insert, inputFullURL, shortURL)
		
		
		fmt.Fprintln(w, "<div style=\"text-align: center;\"><h3 class=\"short\">", shortURL, "</h3></div>")
	
	
	
})

	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		inputFullURL := r.FormValue("originURL")

		
		w.Write(fullURLFormTmpl)
		
		// find	
		rows, _ := db.Query(`select "fullURL" from "urls" where "shortURL" = $1 limit 1`, inputFullURL)
		 
		defer rows.Close()
		var fullURL string
		for rows.Next() { 
			rows.Scan(&fullURL)	 
			fmt.Println(fullURL)
		}	 	
		
		fmt.Fprintln(w, "<div style=\"text-align: center;\"><h3 class=\"short\">", fullURL, "</h3></div>")
})


	staticHandler := http.StripPrefix(
		"/data/",
		http.FileServer(http.Dir("./static")),
	)
	http.Handle("/data/", staticHandler)

	fmt.Println("starting server at :8080")
	http.ListenAndServe(":8080", nil)
}
