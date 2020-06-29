package webServer

import (
	"context"
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	guuid "github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"

	awxConnector "../awxConnector"
	helper "../helper"
	model "../model"
	mongoConnector "../mongoConnector"
)

var Config model.Config

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

func Serve() {
	r := mux.NewRouter()
	// Staic
	fsStatic := http.FileServer(http.Dir("www/static"))
	fsLogos := http.FileServer(http.Dir("/tmp"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fsStatic))
	r.PathPrefix("/icons/").Handler(http.StripPrefix("/tmp/icons/", fsLogos))
	// Login
	r.HandleFunc("/login", loginForm).Methods("GET")
	r.HandleFunc("/login-internal", loginInternal).Methods("POST")
	r.HandleFunc("/logout", logout).Methods("GET")
	// Shop
	r.HandleFunc("/shop", shop)
	r.HandleFunc("/api/v1/shop/list", shopList).Methods("GET")
	// Import
	r.HandleFunc("/templates", templates)
	r.HandleFunc("/import-template-form", importTemplateForm)
	r.HandleFunc("/import-template", importTemplate)
	// Cart
	r.HandleFunc("/cart", cart).Methods("GET")
	r.HandleFunc("/api/v1/cart/list", cartList).Methods("GET")
	r.HandleFunc("/cart-add", cartAdd).Methods("POST")
	r.HandleFunc("/api/v1/cart/remove", cartRemove).Methods("POST")
	r.HandleFunc("/cart-edit", cartEdit).Methods("POST")
	r.HandleFunc("/cart-to-request", cartToRequest).Methods("POST")
	// Request
	r.HandleFunc("/request-template-form", requestTemplateForm).Methods("POST")
	r.HandleFunc("/request-template-form-edit", requestTemplateFormEdit).Methods("POST")
	r.HandleFunc("/request-template-form-reorder", requestTemplateFormReorder).Methods("POST")
	r.HandleFunc("/requests", requests).Methods("GET")
	r.HandleFunc("/reorder", reorder).Methods("POST")
	r.HandleFunc("/api/v1/requests/list", requestsList).Methods("GET")

	srv := &http.Server{
		Addr: "0.0.0.0:8080",
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Info("Shutting down server gracefully")
	// Disconnect from DB
	mongoConnector.DBDisconnect(mongoConnector.Client)
	os.Exit(0)

}

func templateLayout(templateFile string) (*template.Template, error) {
	// Construct the template
	// Template functions
	tFuncs := template.FuncMap{"StringsSplit": strings.Split}
	// New template with attached functions
	name := path.Base(templateFile)
	t := template.New(name).Funcs(tFuncs)
	t, err := t.ParseFiles(templateFile, "www/sources.gohtml", "www/header.gohtml", "www/footer.gohtml")
	return t, err
}

func templates(w http.ResponseWriter, r *http.Request) {
	// Check for active session
	activeSession := securePageHandler(w, r)
	if !activeSession {
		return
	}
	// Show spinner and build the real website asynchonous
	// TODO: Check if the job template is already imported
	// If so: Create update/re-import button
	t, err := templateLayout("www/templates.gohtml")
	if err != nil {
		log.Error(err)
		return
	}
	err = t.Execute(w, awxConnector.GetTemplates())
	if err != nil {
		log.Error(err)
		return
	} else {
		log.Info("Parsed templates.gohtml")
	}
}

// // Get all inputs
// func request(w http.ResponseWriter, r *http.Request) {
// 	// Show spinner and build the real website asynchonous
// 	t := template.Must(template.New("request.gohtml").ParseFiles("www/request.gohtml"))
// 	if r.Method != http.MethodPost {
// 		t.Execute(w, nil)
// 		return
// 	}
//
// 	var request model.RequestDetails
// 	m := make(map[string]string)
//
// 	// Append known and unknown parameters to the request struct
// 	r.ParseForm()
// 	for key, value := range r.Form {
// 		if key == "id" {
// 			request.ID = strings.Join(value, "")
// 		} else if key == "name" {
// 			request.Name = strings.Join(value, "")
// 		} else {
// 			m[key] = strings.Join(value, "")
// 		}
// 	}
// 	request.Parameters = m
//
// 	// Call function to write request to DB
// 	log.Info(request)
//
// 	t.Execute(w, struct{ Success bool }{true})
// }

func importTemplateForm(w http.ResponseWriter, r *http.Request) {
	// Check for active session
	activeSession := securePageHandler(w, r)
	if !activeSession {
		return
	}
	err := r.ParseForm()
	if err != nil {
		log.Error(err)
		return
	}

	id := r.PostFormValue("id")
	if _, err := strconv.Atoi(id); err != nil {
		msg := "Request template failed: id is undefined or malformed"
		log.Error(msg)
		http.Error(w, msg, 400)
		return
	}

	// Get data
	survey, err := awxConnector.GetSurveySpec(id)
	if err != nil {
		log.Error(err)
		return
	}
	survey.Name = r.PostFormValue("name")
	survey.Description = r.PostFormValue("description")

	// Render template
	t, err := templateLayout("www/import-template-form.gohtml")
	if err != nil {
		log.Error(err)
		return
	}

	// Serve site
	err = t.Execute(w, survey)
	if err != nil {
		log.Error(err)
	} else {
		log.Info("Parsed www/import-template-form.gohtml")
	}

}

func importTemplate(w http.ResponseWriter, r *http.Request) {
	// Check for active session
	activeSession := securePageHandler(w, r)
	if !activeSession {
		return
	}
	err := r.ParseForm()
	if err != nil {
		log.Error(err)
		return
	}

	survey, err := awxConnector.GetSurveySpec(r.PostFormValue("template_id"))
	if err != nil {
		log.Error(err)
		return
	}
	log.Info("Parsing import form variables")
	for key, value := range r.Form {
		switch key {
		case "name":
			survey.Name = strings.Join(value, "")
		case "description":
			survey.Description = strings.Join(value, "")
		case "icon-link":
			survey.IconLink = strings.Join(value, "")
		default:
			for i, spec := range survey.Spec {
				if spec.Variable == key {
					survey.Spec[i].RegEx = strings.Join(value, "")
				}
			}
		}
	}
	// Download icon
	if survey.IconLink != "null" {
		log.Info("Downloading icon")
		err, survey.Icon = helper.DownloadIcon(survey.ID, survey.IconLink)
		if err != nil {
			log.Error(err)
			return
		}
	}
	// Write the object to DB
	err = mongoConnector.DBCreateJobTemplate(survey)
	if err != nil {
		log.Error("Import failed")
		return
	}
	http.Redirect(w, r, "/shop", 201)
	return
}

func shop(w http.ResponseWriter, r *http.Request) {
	// Check for active session
	activeSession := securePageHandler(w, r)
	if !activeSession {
		return
	}

	// Render template
	//t := template.New("shop")
	t, err := templateLayout("www/shop.gohtml")
	if err != nil {
		log.Error(err)
		return
	}

	// Serve site
	err = t.Execute(w, nil)
	if err != nil {
		log.Error(err)
	} else {
		log.Info("Parsed www/shop.gohtml")
	}
}

func shopList(w http.ResponseWriter, r *http.Request) {
	activeSession := securePageHandler(w, r)
	if !activeSession {
		return
	}
	var data []model.Survey
	var err error
	data, err = mongoConnector.DBGetJobTemplateAll()
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helper.JsonResponse(w, data)
}

func requestTemplateForm(w http.ResponseWriter, r *http.Request) {
	activeSession := securePageHandler(w, r)
	if !activeSession {
		return
	}
	id := r.PostFormValue("id")
	if _, err := strconv.Atoi(id); err != nil {
		msg := "Request template failed: id is undefined or malformed"
		log.Error(msg)
		http.Error(w, msg, 400)
		return
	}

	// Render template
	// Not using templateLayout because we need a addition function
	//https://www.calhoun.io/intro-to-templates-p3-functions/
	t, err := templateLayout("www/request-template-form.gohtml")
	if err != nil {
		log.Error(err)
		return
	}

	// Get template survey
	templateSurvey, err := mongoConnector.DBGetJobTemplate(id)
	if err != nil {
		log.Error(err)
		return
	}

	// Serve site
	err = t.Execute(w, templateSurvey)
	if err != nil {
		log.Error(err)
	} else {
		log.Info("Parsed www/request-template-form.gohtml")
	}
}

func requestTemplateFormEdit(w http.ResponseWriter, r *http.Request) {
	activeSession := securePageHandler(w, r)
	if !activeSession {
		return
	}
	// Get the user's cart
	userID := getUserID(r)
	cart, err := mongoConnector.DBGetCart(userID)
	if err != nil {
		log.Error(err)
		return
	}

	// Parse request_id
	r.ParseForm()
	if _, err := guuid.Parse(r.PostFormValue("request_id")); err != nil {
		msg := "Edit request failed: request_id is undefined or malformed"
		log.Error(msg)
		http.Error(w, msg, 400)
		return
	}

	var requestID guuid.UUID
	requestID = guuid.MustParse(r.PostFormValue("request_id"))
	log.Info("Received edit request for ", requestID)

	// Get the request out of the cart
	var request model.Request
	for _, item := range cart.Requests {
		if item.ID == requestID {
			request = item
			break
		}
	}

	// Render template
	t, err := templateLayout("www/request-template-form-edit.gohtml")
	if err != nil {
		log.Error(err)
		return
	}

	// Serve site
	err = t.Execute(w, request)
	if err != nil {
		log.Error(err)
	} else {
		log.Info("Parsed www/request-template-form-edit.gohtml")
	}
}

func requestTemplateFormReorder(w http.ResponseWriter, r *http.Request) {
	activeSession := securePageHandler(w, r)
	if !activeSession {
		return
	}
	// Get the user's cart
	userID := getUserID(r)
	// Create a cart for the user
	err := mongoConnector.DBCreateCart(userID)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Parse request_id
	r.ParseForm()
	if _, err := guuid.Parse(r.PostFormValue("request_id")); err != nil {
		msg := "Edit request failed: request_id is undefined or malformed"
		log.Error(msg)
		http.Error(w, msg, 400)
		return
	}
	// Get the request
	var requestID guuid.UUID
	requestID = guuid.MustParse(r.PostFormValue("request_id"))
	log.Info("Received reorder request for ", requestID)
	request, err := mongoConnector.DBGetRequest(userID, requestID)

	// Render template
	t, err := templateLayout("www/request-template-form-edit.gohtml")
	if err != nil {
		log.Error(err)
		return
	}

	// Serve site
	err = t.Execute(w, request)
	if err != nil {
		log.Error(err)
	} else {
		log.Info("Parsed www/request-template-form-edit.gohtml")
	}
}

func cart(w http.ResponseWriter, r *http.Request) {
	activeSession := securePageHandler(w, r)
	if !activeSession {
		return
	}
	// Render template
	t, err := templateLayout("www/cart.gohtml")
	if err != nil {
		log.Error(err)
		return
	}
	// Serve site
	err = t.Execute(w, nil)
	if err != nil {
		log.Error(err)
	} else {
		log.Info("Parsed www/cart.gohtml")
	}
}

func cartList(w http.ResponseWriter, r *http.Request) {
	activeSession := securePageHandler(w, r)
	if !activeSession {
		return
	}
	// Get the user's cart
	userID := getUserID(r)
	cart, err := mongoConnector.DBGetCart(userID)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helper.JsonResponse(w, cart)
}

func cartAdd(w http.ResponseWriter, r *http.Request) {
	activeSession := securePageHandler(w, r)
	if !activeSession {
		return
	}
	err := r.ParseForm()
	if err != nil {
		log.Error(err)
		return
	}
	// Create a request object from form values
	var request model.Request
	r.ParseForm()
	request.TemplateID = r.PostFormValue("template_id")
	request.RequestReason = r.PostFormValue("request_reason")
	// Get the survey spec from imported template
	// and set it for this request
	request.Survey, err = mongoConnector.DBGetJobTemplate(request.TemplateID)
	if err != nil {
		log.Error(err)
		return
	}

	// Fill in the variables
	log.Info("Parsing request form variables")
	for index, item := range request.Survey.Spec {
		for key, value := range r.Form {
			switch key {
			case item.Variable:
				request.Survey.Spec[index].Value = strings.Join(value, "")
			}
		}
	}

	// Cache userID
	userID := getUserID(r)
	// Create a cart for the user
	err = mongoConnector.DBCreateCart(userID)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Add the request object to the users cart
	err = mongoConnector.DBUpdateCartAdd(userID, request)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/cart", 301)
	return
}

func cartRemove(w http.ResponseWriter, r *http.Request) {
	activeSession := securePageHandler(w, r)
	if !activeSession {
		return
	}

	// Parse json
	log.Info("Parsing json for cartRemove request")
	var request model.Request
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Error("Parsing json for cartRemove request failed with error: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Cache userID
	userID := getUserID(r)
	// Create a cart for the user
	err = mongoConnector.DBCreateCart(userID)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Remove the request object from the users cart
	err = mongoConnector.DBUpdateCartRemove(userID, request.ID)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}

func cartEdit(w http.ResponseWriter, r *http.Request) {
	activeSession := securePageHandler(w, r)
	if !activeSession {
		return
	}
	err := r.ParseForm()
	if err != nil {
		log.Error(err)
		return
	}

	// Cache userID
	userID := getUserID(r)
	// Create a cart for the user
	err = mongoConnector.DBCreateCart(userID)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Create a request object from form values
	var request model.Request
	r.ParseForm()
	request.TemplateID = r.PostFormValue("template_id")
	request.RequestReason = r.PostFormValue("request_reason")
	request.ID = guuid.MustParse(r.PostFormValue("request_id"))
	// Get the survey spec from imported template
	// and set it for this request
	request.Survey, err = mongoConnector.DBGetJobTemplate(request.TemplateID)
	if err != nil {
		log.Error(err)
		return
	}

	// Fill in the variables
	log.Info("Parsing request form variables")
	for index, item := range request.Survey.Spec {
		for key, value := range r.Form {
			switch key {
			case item.Variable:
				request.Survey.Spec[index].Value = strings.Join(value, "")
			}
		}
	}
	// Apply the changed request
	err = mongoConnector.DBUpdateCartEdit(userID, request)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/cart", 301)
	return
}

func cartToRequest(w http.ResponseWriter, r *http.Request) {
	activeSession := securePageHandler(w, r)
	if !activeSession {
		return
	}

	// Cache userID
	userID := getUserID(r)
	// Create a request for each item in the users cart
	err := mongoConnector.DBCartToRequest(userID)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}

func requests(w http.ResponseWriter, r *http.Request) {
	activeSession := securePageHandler(w, r)
	if !activeSession {
		return
	}
	// Get the requests
	userID := getUserID(r)
	requests, err := mongoConnector.DBGetRequests(userID)
	if err != nil {
		log.Error(err)
		return
	}
	// Render template
	t, err := templateLayout("www/requests.gohtml")
	if err != nil {
		log.Error(err)
		return
	}

	// Serve site
	err = t.Execute(w, requests)
	if err != nil {
		log.Error(err)
	} else {
		log.Info("Parsed www/requests.gohtml")
	}
}

func reorder(w http.ResponseWriter, r *http.Request) {
	activeSession := securePageHandler(w, r)
	if !activeSession {
		return
	}
	// Get the request
	r.ParseForm()
	requestID := guuid.MustParse(r.PostFormValue("request_id"))
	userID := getUserID(r)
	log.Info("Getting request ", requestID)
	request, err := mongoConnector.DBGetRequest(userID, requestID)
	if err != nil {
		log.Error(err)
		return
	}
	// Add the request to the users cart
	log.Info("Saving a copy of request ", requestID)
	err = mongoConnector.DBUpdateCartAdd(userID, request)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}

func loginForm(w http.ResponseWriter, r *http.Request) {
	// Check for active session
	if userID := getUserID(r); userID != "" {
		http.Redirect(w, r, "/shop", 301)
		return
	}

	// Render template
	t, err := templateLayout("www/login.gohtml")
	if err != nil {
		log.Error(err)
		return
	}

	// Render aditional identity providers (OIDC) into this template

	// Serve site
	err = t.Execute(w, nil)
	if err != nil {
		log.Error(err)
	} else {
		log.Info("Parsed www/login.gohtml")
	}
}
