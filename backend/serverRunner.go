package root

import (
	"net/http"
	"text/template"
	"fmt"
)

func ServerRunner() {
	http.HandleFunc("/", RootHandler) // this is for the home page
	http.HandleFunc("/auth", Auth) // this is for the authentication page
	http.HandleFunc("/register", Register) // this is for the registeration form
	http.HandleFunc("/login", Login) // this is for the login form
	fs := http.FileServer(http.Dir("./frontend/css"))
	http.Handle("/frontend/css/", http.StripPrefix("/frontend/css/", fs))
	// now we are listening for the port 8080
	fmt.Print("The server is running on port :8080\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Server error:", err)
	}

}

func RootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Print("You entered the main page")
	fmt.Print(r.Method)
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
	func Auth(w http.ResponseWriter, r *http.Request) {
		fmt.Print("++++++++++++++++++++++++++++")  // Debug line
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

	func Register(w http.ResponseWriter, r *http.Request){
		if r.Method == "POST" {
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			username, email, password,secondPass := r.FormValue("username"),r.FormValue("email") , r.FormValue("password"), r.FormValue("secondpass")
			if len(username) > 50 || len(email) > 50 || len(password) > 50 || len(secondPass) > 50 {
				http.Error(w,  "Wrong number of values! you are allowed to enter up to 50 digit", http.StatusBadRequest)
				return
			}else if len(username) < 3 || len(email) < 8  || len(password)  < 8  || len(secondPass) < 8 {
				http.Error(w,  "Wrong number of values! you are allowed to enter less than 8 digit", http.StatusBadRequest)
				return
			}
			if username == "" || email == "" || password == "" || secondPass == ""{
				http.Error(w, "You can not keep any field empty you must fill them all!", http.StatusBadRequest)
				return
			}

			fmt.Printf("These are the information collected from the user:\nUserName: {%v}\nEmail: {%s}\nAnd the password: {%s}\n", username, email, password)
			
	}
}

	func Login(w http.ResponseWriter, r *http.Request){
		if r.Method == "POST" {
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			username, password := r.FormValue("username"), r.FormValue("password")
			if len(username) > 50 ||  len(password) > 50 ||  len(username) < 3  ||  len(password) < 8  {
				http.Error(w, "Wrong number of values! you are allowed to enter between 8-50 digit", http.StatusBadRequest)
				return
			}
			if username == "" || password == ""{
				http.Error(w, "You can not keep any field empty you must fill them all!", http.StatusBadRequest)
				return
			}

			fmt.Printf("These are the information collected from the user:\nUserName:{%s}\nAnd the password: {%s}\n", username,password)
			
	}	
	}