package main

import (
	"fmt"
	"github.com/gorilla/sessions"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"src/api/user_operation"
)

var tpl *template.Template
var store = sessions.NewCookieStore([]byte("super-secret-key"))

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}

func init() {
	files, err := filepath.Glob("pages/*.html")
	if err != nil {
		log.Fatal(err)
	}

	if len(files) == 0 {
		log.Fatal("No HTML files found in pages directory")
	}

	tpl = template.Must(template.ParseFiles(files...))

	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   31536000, // кука живёт 100 час
		HttpOnly: true,     // нельзя читать куку из JS
		Secure:   false,    // ОБЯЗАТЕЛЬНО false для HTTP
	}
}

func existsFile(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")

	isLoggedIn := session.Values["username"] != nil

	data := map[string]interface{}{
		"isLoggedIn": isLoggedIn,                 // Если имя пользователя есть в сессии
		"username":   session.Values["username"], // Имя пользователя
		"rules":      session.Values["rules"],    // Роли
	}

	err := tpl.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		http.Error(w, "Ошибка при рендере страницы", http.StatusInternalServerError)
	}
}

func notFoundPage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)

	// опционально: передаём данные в шаблон
	data := struct {
		URL string
	}{
		URL: r.URL.Path,
	}

	// рендерим именно 404.html
	err := tpl.ExecuteTemplate(w, "404.html", data)
	if err != nil {
		// если шаблон упал — возвращаем простой текст
		http.Error(w, "Ошибка при рендере страницы 404", http.StatusInternalServerError)
	}
}

func forbiddenHandler(w http.ResponseWriter, r *http.Request) {
	err := tpl.ExecuteTemplate(w, "403.html", nil)
	if err != nil {
		http.Error(w, "Ошибка при рендере страницы", http.StatusInternalServerError)
	}
}

func regHandle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		err := tpl.ExecuteTemplate(w, "registration.html", nil)
		if err != nil {
			http.Error(w, "Ошибка при рендере страницы", http.StatusInternalServerError)
		}
	case http.MethodPost:
		name := r.FormValue("username")
		mail := r.FormValue("email")
		password := r.FormValue("password")

		err := user_operation.RegisterUser(name, password, mail)
		if err != nil {
			log.Println("Ошибка регистрации:", err)
			http.Error(w, "Ошибка регистрации: "+err.Error(), http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, "/auth", http.StatusSeeOther)
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

func authHandle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		err := tpl.ExecuteTemplate(w, "auth.html", nil)
		if err != nil {
			http.Error(w, "Ошибка при рендере страницы", http.StatusInternalServerError)
		}
	case http.MethodPost:
		name := r.FormValue("username")
		password := r.FormValue("password")

		rules, err := user_operation.AuthenticateUser(name, password)
		if err != nil {
			log.Println("Ошибка авторизации:", err)
			http.Error(w, "Неверный логин или пароль", http.StatusUnauthorized)
			return
		}

		// Сохраняем сессию
		session, _ := store.Get(r, "session")
		session.Values["username"] = name
		session.Values["rules"] = rules
		err = session.Save(r, w)
		if err != nil {
			log.Println("Ошибка сохранения сессии:", err)
			http.Error(w, "Ошибка сессии", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	fmt.Println("Server starting on", port)

	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/forbidden", forbiddenHandler)
	mux.HandleFunc("/notFound", notFoundPage)
	mux.HandleFunc("/register", regHandle)
	mux.HandleFunc("/auth", authHandle)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h, pattern := mux.Handler(r)

		switch {
		case pattern == "":
			// Ни один маршрут не подошел
			notFoundPage(w, r)
			return
		case pattern == "/" && r.URL.Path != "/":
			// Если бы мы ловили "/" — а это не /
			notFoundPage(w, r)
			return
		default:
			// Всё ок — передаём исполнению оригинальный хендлер
			h.ServeHTTP(w, r)
		}
	})

	log.Println("🚀 Сервер запущен на порту", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
