package root

import (
	"net/http"
	"text/template"
	"fmt"
)

func ServerRunner() {
	http.HandleFunc("/", RootHandler) // this is for the home page
	http.HandleFunc("/auth", auth) // this is for the home page
	http.HandleFunc("/register", Register) // this is for the home page
	http.HandleFunc("/login", Login) // this is for the home page
	fs := http.FileServer(http.Dir("./frontend/css"))
	http.Handle("/frontend/css/", http.StripPrefix("/frontend/css/", fs))
	// now we are listening for the port 8080
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Server error:", err)
	}
}

func RootHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.URL.Path != "/" {
		http.Redirect(w, r, "/404", http.StatusSeeOther)
		return
	}
	t, err := template.ParseFiles("./frontend/home.html")
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}
	err2 := t.Execute(w, nil)
	if err2 != nil {
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}

}
	func auth(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		t, err := template.ParseFiles("./frontend/auth.html")
		if err != nil {
			http.Redirect(w, r, "/500", http.StatusSeeOther)
			return
		}
		err2 := t.Execute(w, nil)
		if err2 != nil {
			http.Redirect(w, r, "/500", http.StatusSeeOther)
			return
		}
	}




	
	func Login(w http.ResponseWriter, r *http.Request){

	}

	func Register(w http.ResponseWriter, r *http.Request){

	}