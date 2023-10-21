package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"forum/internal/models"
	"forum/internal/validator"
)

type postCreateForm struct {
	Title    string
	Content  string
	Created  time.Time
	UserId   string
	Category string
	Author   string
	validator.Validator
}

type userSignupForm struct {
	Name      string
	Email     string
	Password  string
	PasswordC string
	validator.Validator
}

type userLoginForm struct {
	UserID   int
	Email    string
	Password string
	validator.Validator
}

type commentForm struct {
	PostID  int
	Comment string
	validator.Validator
}

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		data := app.newTemplateData(r)
		data.Form = userSignupForm{}
		app.render(w, http.StatusOK, "signup.tmpl", data)
	case "POST":
		if err := r.ParseForm(); err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}

		form := userSignupForm{
			Name:      r.PostForm.Get("name"),
			Email:     r.PostForm.Get("email"),
			Password:  r.PostForm.Get("password"),
			PasswordC: r.PostForm.Get("passwordC"),
		}

		form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
		form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
		form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
		form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
		form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")
		form.CheckField(form.Password == form.PasswordC, "password", "Passwords must match")
		// If there are any errors, redisplay the signup form along with a 422
		// status code.
		if !form.Valid() {
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "signup.tmpl", data)
			return
		}
		hashedPassword, err := HashPassword(form.Password)
		if err != nil {
			log.Fatal(err)
		}
		err = app.users.Insert(form.Name, form.Email, hashedPassword)
		if err != nil {
			fmt.Println(form.Email)
			if errors.Is(err, models.ErrDuplicateEmail) {
				fmt.Println("takoy email est'")
				form.AddFieldError("email", "Email address is already in use")
				data := app.newTemplateData(r)
				data.Form = form
				app.render(w, http.StatusUnprocessableEntity, "signup.tmpl", data)
			}
			if errors.Is(err, models.ErrDuplicateUsername) {
				fmt.Println("ya tut bil")
				form.AddFieldError("username", "Username is already in use")
				data := app.newTemplateData(r)
				data.Form = form
				app.render(w, http.StatusUnprocessableEntity, "signup.tmpl", data)
			}
		}
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
	default:
		app.clientError(w, http.StatusMethodNotAllowed)
	}
}

func (app *application) homePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}
	// posts := []*models.post{}
	switch r.Method {
	case "GET":
		posts, err := app.posts.Latest()
		if err != nil {
			app.serverError(w, err)
			return
		}
		data := app.newTemplateData(r)
		data.AuthenticatedUser = app.authenticatedUser(r)
		data.Posts = posts
		app.render(w, http.StatusOK, "home.tmpl", data)

	case "POST":
		err := r.ParseForm()
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}
		categoryArray := r.Form["category"]
		// category := strings.Join(categoryArray, ", ")
		// categories := r.FormValue("category")
		posts, err := app.posts.LatestWithCategory(categoryArray)
		if err != nil {
			app.serverError(w, err)
			return
		}
		data := app.newTemplateData(r)
		data.AuthenticatedUser = app.authenticatedUser(r)
		data.Posts = posts
		app.render(w, http.StatusOK, "home.tmpl", data)

	default:
		app.clientError(w, http.StatusMethodNotAllowed)
	}
}

func (app *application) postView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}
	likes, err := app.reactions.GetLikes(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	dislikes, err := app.reactions.GetDislikes(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	switch r.Method {
	case "GET":
		post, err := app.posts.Get(id)
		if err != nil {
			if errors.Is(err, models.ErrNoRecord) {
				app.notFound(w)
			} else {
				app.serverError(w, err)
			}
			return
		}
		// comments, err :=
		data := app.newTemplateData(r)
		data.Post = post
		data.Post.Likes = likes
		data.Post.Dislikes = dislikes
		data.Comments, err = app.comments.GetComments(id)
		data.Post.Categories, _ = app.postTags.GetCategoriesByPostID(id)
		fmt.Println(data.Post.Categories)
		if err != nil {
			app.serverError(w, err)
			return
		}
		// data.Comment = comment
		data.AuthenticatedUser = app.authenticatedUser(r)
		data.Form = commentForm{}
		app.render(w, http.StatusOK, "view.tmpl", data)

	case "POST":
		fmt.Println("test")
		if app.authenticatedUser(r) == 0 {
			// http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			fmt.Println("tut")
			return
		}
		if err := r.ParseForm(); err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}
		post, err := app.posts.Get(id)
		if err != nil {
			if errors.Is(err, models.ErrNoRecord) {
				app.notFound(w)
			} else {
				app.serverError(w, err)
			}
			return
		}
		cookie, err := r.Cookie("session_token")
		if err != nil {
			if err == http.ErrNoCookie {
				return
			}
			if err == cookie.Valid() {
				return
			}
		}
		UserID := app.sessionManager.GetUserIDBySessionToken(cookie.Value)
		data := app.newTemplateData(r)
		if r.Form.Has("Like") {
			switch r.PostForm.Get("Like") {
			case "1":

				app.reactions.MakeReaction(UserID, id, 1)
			case "-1":
				app.reactions.MakeReaction(UserID, id, -1)

			}
		}
		if r.Form.Has("comment") {

			form := commentForm{
				Comment: r.PostForm.Get("comment"),
			}

			form.CheckField(validator.NotBlank(form.Comment), "comment", "This field cannot be blank")
			form.CheckField(validator.MaxChars(form.Comment, 10), "comment", "This field cannot be more than 200 characters long")
			// Update the post in your database, or wherever it's stored.
			// You need to have a method to update the snippet, something like app.posts.Update(snippet).
			// If there are any errors, redisplay the signup form along with a 422
			// status code.
			if !form.Valid() {
				fmt.Println("maybe")
				data := app.newTemplateData(r)
				data.Form = form
				data.AuthenticatedUser = app.authenticatedUser(r)
				data.Post = post
				data.Comments, err = app.comments.GetComments(id)
				data.Post.Likes = likes
				data.Post.Dislikes = dislikes
				app.render(w, http.StatusUnprocessableEntity, "view.tmpl", data)
				return
			}
			app.comments.Insert(UserID, id, form.Comment)

		}
		data.AuthenticatedUser = app.authenticatedUser(r)
		data.Post = post
		data.Comments, err = app.comments.GetComments(id)
		data.Post.Likes = likes
		data.Post.Dislikes = dislikes
		http.Redirect(w, r, fmt.Sprintf("/post/view?id=%d", id), 302)
		return
	default:
		app.clientError(w, http.StatusMethodNotAllowed)
	}
}

func (app *application) postCreate(w http.ResponseWriter, r *http.Request) {
	if app.authenticatedUser(r) == 0 {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	switch r.Method {
	case "GET":
		data := app.newTemplateData(r)
		data.Form = postCreateForm{}
		data.AuthenticatedUser = app.authenticatedUser(r)
		app.render(w, http.StatusOK, "create.tmpl", data)
	case "POST":
		err := r.ParseForm()
		if err != nil {
			fmt.Println("err tut")
			app.clientError(w, http.StatusBadRequest)
			return
		}
		form := postCreateForm{
			Title:    r.PostForm.Get("title"),
			Content:  r.PostForm.Get("content"),
			Category: r.PostForm.Get("category"),
		}
		form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
		form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters long")
		form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")
		form.CheckField(validator.NotBlank(form.Category), "category", "This field cannot be blank")
		// Use the Valid() method to see if any of the checks failed. If they did,
		// then re-render the template passing in the form in the same way as
		// before.
		if !form.Valid() {
			fmt.Println("ili tut")
			data := app.newTemplateData(r)
			data.Form = form
			data.AuthenticatedUser = app.authenticatedUser(r)
			app.render(w, http.StatusBadRequest, "create.tmpl", data)
			return
		}
		cookie, err := r.Cookie("session_token")
		if err != nil {
			if err == http.ErrNoCookie {
				return
			}
			if err == cookie.Valid() {
				return
			}
		}

		// Получение значения массива строк из формы
		values := r.Form["category"]
		form.Category = strings.Join(values, ", ")

		UserID := app.sessionManager.GetUserIDBySessionToken(cookie.Value)
		id, err := app.posts.Insert(form.Title, form.Content, UserID)
		if err != nil {
			app.serverError(w, err)
			return
		}

		categories := strings.Split(form.Category, ", ")
		fmt.Println(categories, " - Categories")
		for _, category := range categories {
			fmt.Println(category)
			err := app.postTags.InsertCategory(id, category)
			fmt.Println("OK")
			if err != nil {
				app.serverError(w, err)
				return
			}
		}
		http.Redirect(w, r, fmt.Sprintf("/post/view?id=%d", id), http.StatusSeeOther)
	default:
		app.clientError(w, http.StatusMethodNotAllowed)
	}
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		data := app.newTemplateData(r)
		data.Form = userLoginForm{}
		app.render(w, http.StatusOK, "login.tmpl", data)
	case "POST":
		if err := r.ParseForm(); err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}
		form := userLoginForm{
			Email:    r.PostForm.Get("email"),
			Password: r.PostForm.Get("password"),
		}

		form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
		form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
		form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
		if !form.Valid() {
			data := app.newTemplateData(r)
			data.Form = form
			data.AuthenticatedUser = app.authenticatedUser(r)
			app.render(w, http.StatusBadRequest, "login.tmpl", data)
			return
		}
		// Check whether the credentials are valid. If they're not, add a generic
		// non-field error message and re-display the login page.
		id, err := app.users.Authenticate(form.Email, form.Password)
		if err != nil {
			if errors.Is(err, models.ErrInvalidCredentials) {
				form.AddNonFieldError("Email or password is incorrect")
				data := app.newTemplateData(r)
				data.Form = form
				app.render(w, http.StatusUnprocessableEntity, "login.tmpl", data)
			} else {
				app.serverError(w, err)
			}
			return
		}
		app.sessionManager.CreateSession(w, r, id)
		// Получаем значение UserID из контекста

		http.Redirect(w, r, "/post/create", http.StatusSeeOther)
	default:
		app.clientError(w, http.StatusMethodNotAllowed)
	}
}

func (app *application) userLogout(w http.ResponseWriter, r *http.Request) {
	// Получите сессию пользователя, если она существует
	sessionCookie, err := r.Cookie("session_token")

	if err == nil {
		// Удаление сессии
		err := app.sessionManager.DeleteSession(sessionCookie.Value)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}

	// Очистка куков и перенаправление на главную страницу
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   "",
		Expires: time.Unix(0, 0),
		Path:    "/",
		MaxAge:  -1,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) userPosts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		cookie, err := r.Cookie("session_token")
		if err != nil {
			if err == http.ErrNoCookie {
				return
			}
			if err == cookie.Valid() {
				return
			}
		}
		userID := app.sessionManager.GetUserIDBySessionToken(cookie.Value)
		posts, err := app.posts.GetUserPosts(userID)
		if err != nil {
			app.serverError(w, err)
			return
		}

		data := app.newTemplateData(r)
		data.AuthenticatedUser = app.authenticatedUser(r)
		data.Posts = posts
		app.render(w, http.StatusOK, "userPosts.tmpl", data)
	default:
		app.clientError(w, http.StatusMethodNotAllowed)
	}
}
