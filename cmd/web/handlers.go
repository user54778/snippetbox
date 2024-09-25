package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"snippetbox.adpollak.net/internal/models"
	"snippetbox.adpollak.net/internal/validator"

	"github.com/julienschmidt/httprouter"
)

// Represent the form data and validation errors for the
// form fields. Note all struct fields are deliberately EXPORTED
// (i.e., start w/ Capital letter). Struct fields must be exported
// in order to be read by the html/template package when rendering a template.
type snippetCreateForm struct {
	Title   string `form:"title"`
	Content string `form:"content"`
	Expires int    `form:"expires"`
	// FieldErrors map[string]string
	validator.Validator `form:"-"` // composition
}

// Handler
// NOTE: This signature was changed to be defined as a method against the *application type.
// This allows us to not depend on some specific type.
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// Before, we had to check the r.URL.Path != "/" for this handler.
	// Since httprouter matches the "/" path EXACTLY, we can remove this.

	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Call newTemplateData() helper to get a templateData struct containing the
	// 'default' data (for now curr year) and add the snippets slice to it.
	data := app.newTemplateData(r)
	data.Snippets = snippets

	// Pass data to the render() helper
	app.render(w, http.StatusOK, "home.tmpl", data)
}

// Handler
func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	// When parsing a request, any named parameters will be stored in the request context.
	params := httprouter.ParamsFromContext(r.Context())

	// Instead, use ByName() to get the value of id named parameter from the slice
	// and validate it as normal.
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.notFound(w)
		return
	}

	// Use SnippetModel's Get method to retrieve the data for a specific
	// record based on its ID. If no record is found, return a 404 Not Found response.
	snippet, err := app.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	/*
		// Retrieve and remove our flash message from the sessionManager
		flash := app.sessionManager.PopString(r.Context(), "flash")
	*/

	data := app.newTemplateData(r)
	data.Snippet = snippet

	/*
		// Now pass the flash message to the template
		data.Flash = flash
	*/

	// Use our new render helper.
	app.render(w, http.StatusOK, "view.tmpl", data)
}

// Render the html form from the GET method.
func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	// w.Write([]byte("Display the form for creating a new snippet here..."))
	data := app.newTemplateData(r)

	// NOTE: init a new createSnippetForm instance and pass it to the template.
	// Otherwise, upon visiting /snippet/create Go would try to eval some tmpl tag
	// such as .Form.FieldErrors.title which would be nil
	data.Form = snippetCreateForm{
		Expires: 365, // default value
	}

	app.render(w, http.StatusOK, "create.tmpl", data)
}

// Processes and uses form data of snippet. Upon completion
// redirects user to the view page to view their snippet.
func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	/*
		// Step 1: parse request body
		err := r.ParseForm()
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}
	*/
	var form snippetCreateForm // zero-valued instance of snippetCreateForm struct.

	// Step 2: retrieve PostForm() data. Expires is a integer
	// so convert with strconv.Atoi() and throw an error is fail
	// conversion.
	/*
		expires, err := strconv.Atoi(r.PostForm.Get("expires"))
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}

		form := snippetCreateForm{
			Title:   r.PostForm.Get("title"),
			Content: r.PostForm.Get("content"),
			Expires: expires,
		}
	*/

	// Use our helper to handle nil panic.
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Call CheckField to execute our validation checks.
	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters long")
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")
	// Use generic PermittedValue() instead of type-specific PermittedInt().
	form.CheckField(validator.PermittedValue(form.Expires, 1, 7, 365), "expires", "This field must equal 1, 7, or 365")

	// Use Valid() to see if any checks failed.
	// If so, re-render passing in the form as before.
	if !form.Valid() {
		// re-render
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "create.tmpl", data)
		return
	}

	// insert title, content, expiration into db
	id, err := app.snippets.Insert(form.Title, form.Content, form.Expires)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// NOTE: Use Put() to add a string value and corresponding flash key to session data.
	app.sessionManager.Put(r.Context(), "flash", "Snippet created successfully!")

	// w.Write([]byte("Create a new snippet...\"))
	// Redirect the user to the relevant page for the snippet.
	// Update redirect path to the new clean URL format.
	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}

// Hold form data for the user signup.
type userSignupForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

// Hold form data for the user login.
type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

// NOTE: User handlers

// Handler to display an HTML form for signing up a new user.
func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userSignupForm{} // no defaults
	app.render(w, http.StatusOK, "signup.tmpl", data)
}

// Handler to process the HTML so as to create a new user.
func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	var form userSignupForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Validate form contents using our helper functions.
	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "signup.tmpl", data)
		return
	}

	// Try to create a new user record in the database. If the email already
	// exists then add an error message to the form and re-display it.
	err = app.users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			log.Println("Duplicate Detected")
			form.AddFieldError("email", "Email address already in use")

			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "signup.tmpl", data)
		} else {
			// NOTE: There was an issue causing it to throw a serverError every time an email a duplicate.
			// I believe the issue was due to the Users table; perhaps index or data type was setup incorrectly?
			app.serverError(w, err)
		}

		return
	}

	// Otherwise add a confirmation flash message to the session confirming
	// that their signup worked.
	app.sessionManager.Put(r.Context(), "flash", "Your signup was successful. Please log in.")

	// Then redirect the user to the login page.
	http.Redirect(w, r, "/user/login", http.StatusSeeOther) // NOTE:
}

// Handler for displaying an HTML form for logging in a user.
func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userLoginForm{} // no defaults
	app.render(w, http.StatusOK, "login.tmpl", data)
}

// Handler to process the HTML form so as to authenticate and login the user.
func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	// 1) Parse the submitted login form data
	var form userLoginForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// 1a) validation form checks
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "login.tmpl", data)
		return
	}

	// 2) Call authenticate
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

	// 3) add to our session if correct authentication and then redirect
	// Use RenewToken() to generate a new session ID when the authentication
	// state/privelege levels change for the user. (i.e., login/logout operations)
	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)

	// Check the user's session for a URL path that they may have been attempting
	// to login to.
	urlPath := app.sessionManager.PopString(r.Context(), "redirectPathAfterLogin")
	if urlPath != "" {
		http.Redirect(w, r, urlPath, http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/snippet/create", http.StatusSeeOther)
}

// Handler to process the HTML form so as to logout the user.
func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Remove the authenticatedUserID from the session data so that the user is
	// logged out.
	app.sessionManager.Remove(r.Context(), "authenticatedUserID")

	app.sessionManager.Put(r.Context(), "flash", "You have been logged out successfully")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// NOTE: Account handlers

type accountPasswordUpdateForm struct {
	CurrentPassword         string `form:"currentPassword"`
	NewPassword             string `form:"newPassword"`
	NewPasswordConfirmation string `form:"newPasswordConfirmation"`
	validator.Validator     `form:"-"`
}

// Handler to display an HTML form containing an authenticated
// user's name, email, and join date.
func (app *application) accountView(w http.ResponseWriter, r *http.Request) {
	// 1) get the authenticatedUserID from the session
	userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
	// 2) fetch details of the relevant user from the database
	// dump in plain text HTTP response
	user, err := app.users.Get(userID)
	if err != nil {
		// Should redirect if not logged in
		if errors.Is(err, models.ErrNoRecord) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		} else {
			app.serverError(w, err)
		}
		return
	}

	// Render the account template
	data := app.newTemplateData(r)
	data.User = user // use the user data to fill out the template

	app.render(w, http.StatusOK, "account.tmpl", data)
}

// Handler for displaying an HTML form
// for allowing a user to change their password.
func (app *application) accountPasswordUpdate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = accountPasswordUpdateForm{}

	app.render(w, http.StatusOK, "password.tmpl", data)
}

// Handler that processes an HTML form for chaning a user's password.
func (app *application) accountPasswordUpdatePost(w http.ResponseWriter, r *http.Request) {
	var form accountPasswordUpdateForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Form validation checks: all three fields required,
	// newPassword at least 8 chars long, newPasswordConfirmation and newPassword MATCH.
	form.CheckField(validator.NotBlank(form.CurrentPassword), "currentPassword", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.NewPassword), "newPassword", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.NewPasswordConfirmation), "newPasswordConfirmation", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.NewPassword, 8), "newPassword", "This field must be at least 8 characters long")
	form.CheckField(form.NewPassword == form.NewPasswordConfirmation, "newPasswordConfirmation", "Passwords must match")

	// Re-display form
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "password.tmpl", data)
		return
	}

	// 2) Call PasswordUpdate
	userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")

	err = app.users.PasswordUpdate(userID, form.CurrentPassword, form.NewPassword)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {

			form.AddNonFieldError("Email or password is incorrect")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "password.tmpl", data)
		} else {
			app.serverError(w, err)
		}
		return
	}

	// 3) Flash a message to user session saying password updated
	app.sessionManager.Put(r.Context(), "flash", "Your password has been updated.")

	http.Redirect(w, r, "/account/view", http.StatusSeeOther)
}

// NOTE: miscellaneous handlers

// Handler to display an HTML form of the about page.
func (app *application) about(w http.ResponseWriter, r *http.Request) {
	// Simply render the template and nothing else.
	data := app.newTemplateData(r)
	app.render(w, http.StatusOK, "about.tmpl", data)
}

// A simple ping handler to demonstrate testing handlers and middleware.
func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
