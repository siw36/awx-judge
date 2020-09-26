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
	// Default
	r.HandleFunc("/", loginForm).Methods("GET")
	// Login
	r.HandleFunc("/login", loginForm).Methods("GET")
	r.HandleFunc("/api/v1/login/internal", loginInternal).Methods("POST")
	r.HandleFunc("/logout", logout).Methods("GET")
	// Shop
	r.HandleFunc("/shop", shop).Methods("GET")
	// Templates import
	r.HandleFunc("/template-import", templateImport)
	r.HandleFunc("/template-import-form", templateImportForm).Methods("GET")
	r.HandleFunc("/api/v1/import/add", templateImportAdd).Methods("POST")
	r.HandleFunc("/api/v1/import/list", templateImportList).Methods("GET")
	r.HandleFunc("/api/v1/import/get", templateImportGet).Methods("POST")
	r.HandleFunc("/api/v1/import/survey/get", templateImportSurveyGet).Methods("POST")
	// Templates
	r.HandleFunc("/api/v1/templates/get", templateGet).Methods("POST")
	r.HandleFunc("/api/v1/templates/list", templateList).Methods("GET")
	r.HandleFunc("/api/v1/templates/remove", templateRemove).Methods("POST")
	// Cart
	r.HandleFunc("/cart", cart).Methods("GET")
	r.HandleFunc("/api/v1/cart/list", cartList).Methods("GET")
	r.HandleFunc("/api/v1/cart/add", cartAdd).Methods("POST")
	r.HandleFunc("/api/v1/cart/remove", cartRemove).Methods("POST")
	r.HandleFunc("/api/v1/cart/edit", cartEdit).Methods("POST")
	r.HandleFunc("/api/v1/cart/execute", cartToRequest).Methods("POST")
	// Request
	r.HandleFunc("/request", request).Methods("POST", "GET")
	// Requests (already existant)
	r.HandleFunc("/requests", requests).Methods("GET")
	r.HandleFunc("/api/v1/requests/reorder", requestReorder).Methods("POST")
	r.HandleFunc("/api/v1/requests/list", requestsList).Methods("GET")
	r.HandleFunc("/api/v1/requests/approve", requestsApprove).Methods("POST")
	r.HandleFunc("/api/v1/requests/deny", requestsDeny).Methods("POST")
	r.HandleFunc("/api/v1/requests/get", requestsGet).Methods("GET")
	// r.HandleFunc("/api/v1/requests/remove", requestRemove).Methods("POST") // Do not use this in v0.1 (too complicated)
	// r.HandleFunc("/api/v1/requests/edit", requestEdit).Methods("POST") // Do not use this in v0.1 (too complicated)

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
		log.Info("Starting server at ", srv.Addr)
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

func goTemplateLayout(templateFile string) (*template.Template, error) {
	// Construct the template
	// Template functions
	tFuncs := template.FuncMap{"StringsSplit": strings.Split}
	// New template with attached functions
	name := path.Base(templateFile)
	t := template.New(name).Funcs(tFuncs)
	t, err := t.ParseFiles(templateFile, "www/sources.gohtml", "www/header.gohtml", "www/footer.gohtml")
	return t, err
}

////
// Templates import
////

func templateImport(w http.ResponseWriter, r *http.Request) {
	// Check for active session
	activeSession := securePageHandler(w, r)
	if !activeSession {
		return
	}
	t, err := goTemplateLayout("www/template-import.gohtml")
	if err != nil {
		log.Error(err)
		return
	}
	err = t.Execute(w, nil)
	if err != nil {
		log.Error(err)
		return
	} else {
		log.Info("Parsed template-import.gohtml")
	}
}

func templateImportForm(w http.ResponseWriter, r *http.Request) {
	// Check for active session
	activeSession := securePageHandler(w, r)
	if !activeSession {
		return
	}

	// Render template
	t, err := goTemplateLayout("www/template-import-form.gohtml")
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Serve site
	err = t.Execute(w, nil)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		log.Info("Parsed www/template-import-form.gohtml")
	}
}

func templateImportList(w http.ResponseWriter, r *http.Request) {
	// Check for active session
	activeSession := securePageHandler(w, r)
	if !activeSession {
		return
	}
	templates, err := awxConnector.GetTemplates()
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helper.JsonResponse(w, templates)
}

func templateImportGet(w http.ResponseWriter, r *http.Request) {
	// Check for active session
	activeSession := securePageHandler(w, r)
	if !activeSession {
		return
	}
	var templateQuery model.Template
	err := json.NewDecoder(r.Body).Decode(&templateQuery)
	if err != nil {
		log.Error("Parsing json for templateImportGet request failed with error: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	template, err := awxConnector.GetTemplate(templateQuery.ID)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helper.JsonResponse(w, template)
}

func templateImportSurveyGet(w http.ResponseWriter, r *http.Request) {
	// Check for active session
	activeSession := securePageHandler(w, r)
	if !activeSession {
		return
	}
	var templateQuery model.Template
	err := json.NewDecoder(r.Body).Decode(&templateQuery)
	if err != nil {
		log.Error("Parsing json for templateImportSurveyGet request failed with error: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	survey, err := awxConnector.GetSurvey(templateQuery.ID)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helper.JsonResponse(w, survey)
}

func templateImportAdd(w http.ResponseWriter, r *http.Request) {
	// Check for active session
	activeSession := securePageHandler(w, r)
	if !activeSession {
		return
	}
	err := r.ParseForm()
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	templateID, _ := strconv.Atoi(r.PostFormValue("template_import_form_id"))
	template, err := awxConnector.GetTemplate(templateID)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	template.Survey, err = awxConnector.GetSurvey(templateID)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info("Parsing import form variables")
	for key, value := range r.Form {
		switch key {
		case "template_name":
			template.Name = strings.Join(value, "")
		case "template_description":
			template.Description = strings.Join(value, "")
		case "template_icon_link":
			template.IconLink = strings.Join(value, "")
		default:
			for i, spec := range template.Survey {
				if spec.Variable == key {
					template.Survey[i].RegEx = strings.Join(value, "")
				}
			}
		}
	}
	// Download icon
	if helper.ValidUrl(template.IconLink) {
		log.Info("Downloading icon")
		err, template.Icon = helper.DownloadIcon(template.ID, template.IconLink)
		if err != nil {
			log.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		log.Info("Skipping icon download due to malformed URL")
	}
	// Write the object to DB
	err = mongoConnector.DBCreateTemplate(template)
	if err != nil {
		log.Error("Import failed")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}

////
// Template
////

func templateList(w http.ResponseWriter, r *http.Request) {
	activeSession := securePageHandler(w, r)
	if !activeSession {
		return
	}
	var templates []model.Template
	var err error
	templates, err = mongoConnector.DBGetTemplateAll()
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helper.JsonResponse(w, templates)
}

func templateRemove(w http.ResponseWriter, r *http.Request) {
	activeSession := securePageHandler(w, r)
	if !activeSession {
		return
	}

	// Parse json
	log.Info("Parsing json for templateRemove request")
	var template model.Template
	err := json.NewDecoder(r.Body).Decode(&template)
	if err != nil {
		log.Error("Parsing json for templateRemove request failed with error: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = mongoConnector.DBRemoveTemplate(template.ID)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// ToDo: delete the icon
	w.WriteHeader(http.StatusOK)
	return
}

func templateGet(w http.ResponseWriter, r *http.Request) {
	activeSession := securePageHandler(w, r)
	if !activeSession {
		return
	}

	// Parse json
	log.Info("Parsing json for templateGet request")
	var templateQuery model.Template
	err := json.NewDecoder(r.Body).Decode(&templateQuery)
	if err != nil {
		log.Error("Parsing json for templateGet request failed with error: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	template, err := mongoConnector.DBGetTemplate(templateQuery.ID)
	if err != nil {
		log.Error(err)
		helper.JsonResponse(w, template)
		return
	}
	helper.JsonResponse(w, template)
	return
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

func shop(w http.ResponseWriter, r *http.Request) {
	// Check for active session
	activeSession := securePageHandler(w, r)
	if !activeSession {
		return
	}

	// Render template
	//t := template.New("shop")
	t, err := goTemplateLayout("www/shop.gohtml")
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

func request(w http.ResponseWriter, r *http.Request) {
	activeSession := securePageHandler(w, r)
	if !activeSession {
		return
	}

	t, err := goTemplateLayout("www/request.gohtml")
	if err != nil {
		log.Error(err)
		return
	}

	// Serve site
	err = t.Execute(w, nil)
	if err != nil {
		log.Error(err)
	} else {
		log.Info("Parsed www/request.gohtml")
	}
}

func requestSpec(w http.ResponseWriter, r *http.Request) {
	activeSession := securePageHandler(w, r)
	if !activeSession {
		return
	}
	id, err := strconv.Atoi(r.PostFormValue("id"))
	if err != nil {
		msg := "Request temaplte spec failed: id is undefined or malformed"
		log.Error(msg)
		http.Error(w, msg, 400)
		return
	}
	template, err := mongoConnector.DBGetTemplate(id)
	if err != nil {
		log.Error(err)
		return
	}
	helper.JsonResponse(w, template)
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
	t, err := goTemplateLayout("www/request-template-form-edit.gohtml")
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
	t, err := goTemplateLayout("www/cart.gohtml")
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
	request.TemplateID, err = strconv.Atoi(r.PostFormValue("template_id"))
	if err != nil {
		log.Error(err)
		return
	}
	request.RequestReason = r.PostFormValue("request_reason")
	// Get the survey spec from imported template
	// and set it for this request
	request.Template, err = mongoConnector.DBGetTemplate(request.TemplateID)
	if err != nil {
		log.Error(err)
		return
	}

	// Fill in the variables
	log.Info("Parsing request form variables")
	for index, item := range request.Template.Survey {
		for key, value := range r.Form {
			switch key {
			case item.Variable:
				request.Template.Survey[index].Value = strings.Join(value, "")
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
	request.TemplateID, err = strconv.Atoi(r.PostFormValue("template_id"))
	if err != nil {
		log.Error(err)
		return
	}
	request.RequestReason = r.PostFormValue("request_reason")
	request.ID = guuid.MustParse(r.PostFormValue("request_id"))
	// Get the survey spec from imported template
	// and set it for this request
	request.Template, err = mongoConnector.DBGetTemplate(request.TemplateID)
	if err != nil {
		log.Error(err)
		return
	}

	// Fill in the variables
	log.Info("Parsing request form variables")
	for index, item := range request.Template.Survey {
		for key, value := range r.Form {
			switch key {
			case item.Variable:
				request.Template.Survey[index].Value = strings.Join(value, "")
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
	t, err := goTemplateLayout("www/requests.gohtml")
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

func requestReorder(w http.ResponseWriter, r *http.Request) {
	activeSession := securePageHandler(w, r)
	if !activeSession {
		return
	}

	// Parse json
	log.Info("Parsing json for requestReorder request")
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

	// Get the request
	var newRequest model.Request
	newRequest, err = mongoConnector.DBGetRequest(userID, request.ID)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Add the new request to the users cart
	err = mongoConnector.DBUpdateCartAdd(userID, newRequest)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func requestsList(w http.ResponseWriter, r *http.Request) {
	activeSession := securePageHandler(w, r)
	if !activeSession {
		return
	}
	userID := getUserID(r)
	requests, err := mongoConnector.DBGetRequests(userID)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helper.JsonResponse(w, requests)
}

func requestsGet(w http.ResponseWriter, r *http.Request) {
	activeSession := securePageHandler(w, r)
	if !activeSession {
		return
	}
	userID := getUserID(r)
	vars := mux.Vars(r)
	requestID := guuid.MustParse(vars["requestID"])
	requests, err := mongoConnector.DBGetRequest(userID, requestID)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helper.JsonResponse(w, requests)
}

func requestsApprove(w http.ResponseWriter, r *http.Request) {
	activeSession := securePageHandler(w, r)
	if !activeSession {
		return
	}
	userID := getUserID(r)

	// Parse json
	log.Info("Parsing json for requestsApprove request")
	var changes model.Request
	err := json.NewDecoder(r.Body).Decode(&changes)
	if err != nil {
		log.Error("Parsing json for requestsApprove request failed with error: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the user request
	request, err := mongoConnector.DBGetRequest(userID, changes.ID)

	if request.State != "pending" {
		log.Error("Request state is not pending. Approval canceled.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Apply the changes to the request
	request.Reason = changes.Reason
	request.State = "approved"
	message := "Approved by " + userID
	request.LastMessage = message
	request.Messages = append(request.Messages, message)
	request.JudgeID = userID

	// Write the updated request to DB
	err = mongoConnector.DBUpdateRequest(userID, request)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}

func requestsDeny(w http.ResponseWriter, r *http.Request) {
	activeSession := securePageHandler(w, r)
	if !activeSession {
		return
	}
	userID := getUserID(r)

	// Parse json
	log.Info("Parsing json for requestsDeny request")
	var request model.Request
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Error("Parsing json for requestsDeny request failed with error: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Apply the changes to the request
	request.State = "denied"
	message := "Denied by " + userID
	request.LastMessage = message
	request.Messages = append(request.Messages, message)
	request.JudgeID = userID

	// create db function for that
	err = mongoConnector.DBUpdateRequest(userID, request)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}

// func reorder(w http.ResponseWriter, r *http.Request) {
// 	activeSession := securePageHandler(w, r)
// 	if !activeSession {
// 		return
// 	}
// 	// Get the request
// 	r.ParseForm()
// 	requestID := guuid.MustParse(r.PostFormValue("request_id"))
// 	userID := getUserID(r)
// 	log.Info("Getting request ", requestID)
// 	request, err := mongoConnector.DBGetRequest(userID, requestID)
// 	if err != nil {
// 		log.Error(err)
// 		return
// 	}
// 	// Add the request to the users cart
// 	log.Info("Saving a copy of request ", requestID)
// 	err = mongoConnector.DBUpdateCartAdd(userID, request)
// 	if err != nil {
// 		log.Error(err)
// 		w.WriteHeader(http.StatusInternalServerError)
// 		return
// 	}
// 	w.WriteHeader(http.StatusOK)
// 	return
// }

func loginForm(w http.ResponseWriter, r *http.Request) {
	// Check for active session
	if userID := getUserID(r); userID != "" {
		http.Redirect(w, r, "/shop", 301)
		return
	}

	// Render template
	t, err := goTemplateLayout("www/login.gohtml")
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
