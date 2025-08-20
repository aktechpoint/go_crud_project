package main

import (
	"database/sql"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type Student struct {
	ID    int
	Name  string
	Email string
	Age   int
	Image string
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("mysql", "root:abhi1234@tcp(127.0.0.1:3306)/studentdb")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		panic(err)
	}

	// Static files (images)
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	// Routes
	http.HandleFunc("/", listStudents)
	http.HandleFunc("/add", addStudent)
	http.HandleFunc("/edit", editStudent)
	http.HandleFunc("/delete", deleteStudent)

	// Start server
	http.ListenAndServe(":8080", nil)
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	t := template.Must(template.ParseFiles("templates/" + tmpl))
	t.Execute(w, data)
}

func listStudents(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name, email, age, image FROM students")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var students []Student
	for rows.Next() {
		var s Student
		err := rows.Scan(&s.ID, &s.Name, &s.Email, &s.Age, &s.Image)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		students = append(students, s)
	}
	renderTemplate(w, "index.html", students)
}

func addStudent(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseMultipartForm(10 << 20)

		name := r.FormValue("name")
		email := r.FormValue("email")
		age, _ := strconv.Atoi(r.FormValue("age"))

		file, handler, err := r.FormFile("image")
		imagePath := ""
		if err == nil && handler != nil {
			defer file.Close()

			// Optional: replace spaces in file name
			safeFilename := strings.ReplaceAll(handler.Filename, " ", "_")

			// Save image using forward slashes for URL
			localPath := filepath.ToSlash(filepath.Join("uploads", safeFilename))
			dst, err := os.Create(localPath)
			if err != nil {
				http.Error(w, "Unable to save image", http.StatusInternalServerError)
				return
			}
			defer dst.Close()

			_, _ = dst.ReadFrom(file)

			// Save web-accessible image path (for browser)
			imagePath = "/" + localPath
		}

		_, err = db.Exec("INSERT INTO students (name, email, age, image) VALUES (?, ?, ?, ?)", name, email, age, imagePath)
		if err != nil {
			http.Error(w, "Failed to insert student", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	renderTemplate(w, "add.html", nil)
}

func editStudent(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))

	if r.Method == http.MethodPost {
		r.ParseMultipartForm(10 << 20)
		name := r.FormValue("name")
		email := r.FormValue("email")
		age, _ := strconv.Atoi(r.FormValue("age"))

		// Check if a new image is uploaded
		file, handler, err := r.FormFile("image")
		if err == nil && handler != nil {
			defer file.Close()

			imagePath := filepath.Join("uploads", handler.Filename)
			dst, err := os.Create(imagePath)
			if err != nil {
				http.Error(w, "Unable to save image", http.StatusInternalServerError)
				return
			}
			defer dst.Close()
			_, _ = dst.ReadFrom(file)

			imageWebPath := "/uploads/" + handler.Filename

			// Update with image
			_, err = db.Exec("UPDATE students SET name=?, email=?, age=?, image=? WHERE id=?", name, email, age, imageWebPath, id)
			if err != nil {
				http.Error(w, "Failed to update student with image", http.StatusInternalServerError)
				return
			}
		} else {
			// Update without changing image
			_, err = db.Exec("UPDATE students SET name=?, email=?, age=? WHERE id=?", name, email, age, id)
			if err != nil {
				http.Error(w, "Failed to update student", http.StatusInternalServerError)
				return
			}
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var s Student
	row := db.QueryRow("SELECT id, name, email, age, image FROM students WHERE id=?", id)
	err := row.Scan(&s.ID, &s.Name, &s.Email, &s.Age, &s.Image)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, "edit.html", s)
}


func deleteStudent(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	_, err := db.Exec("DELETE FROM students WHERE id=?", id)
	if err != nil {
		http.Error(w, "Failed to delete student", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
