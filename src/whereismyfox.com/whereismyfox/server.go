package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type DeviceInformation struct {
	DeviceName string `json: "name"`
	PushURL    string `json: "pushURL"`
}

type DeviceListResponse struct {
	Devices []DeviceInformation `json: "devices"`
}

func deviceListHandler(w http.ResponseWriter, r *http.Request) {

	if !IsLoggedIn(r) {
		log.Println("deviceListHandler: user not logged in")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("."))
		return
	}

	loginName := GetLoginName(r)
	if loginName == "" {
		log.Println("deviceListHandler: user does not have an email address")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("."))
		return
	}

	deviceList := devicesForUser(loginName)

	if deviceList == nil {
		log.Println("deviceListHandler: device list is empty for user")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("."))
		return
	}

	var data []byte
	var err error

	response := DeviceListResponse{deviceList}

	data, err = json.Marshal(response)
	if err != nil {
		log.Println("deviceListHandler: could not marshal data")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("."))
		return
	}

	w.Write(data)
}

func deviceAddHandler(w http.ResponseWriter, r *http.Request) {

	if !IsLoggedIn(r) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("."))
		return
	}

	loginName := GetLoginName(r)
	if loginName == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("."))
		return
	}

	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte("Bad Request."))
		return
	}

	pushURL := r.FormValue("pushURL")
	if pushURL == "" {
		w.WriteHeader(400)
		w.Write([]byte("Bad Request."))
		return
	}

	deviceName := r.FormValue("deviceName")
	if deviceName == "" {
		w.WriteHeader(400)
		w.Write([]byte("Bad Request."))
		return
	}

	if addDevice(loginName, deviceName, pushURL) {
		w.Write([]byte("ok"))
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("."))
}

func deviceDeleteHandler(w http.ResponseWriter, r *http.Request) {

	if !IsLoggedIn(r) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("."))
		return
	}

	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte("Bad Request."))
	}

	pushURL := r.FormValue("pushURL")
	if pushURL == "" {
		log.Println(err)
		w.WriteHeader(400)
		w.Write([]byte("Bad Request."))
	}

	if deleteDevice(pushURL) {
		w.Write([]byte("ok"))
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("."))
}

func deviceLocationHandler(w http.ResponseWriter, r *http.Request) {

	if !IsLoggedIn(r) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("."))
		return
	}

	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte("Bad Request."))
	}

	pushURL := r.FormValue("pushURL")

	lat, err1 := strconv.ParseFloat(r.FormValue("lat"), 64)
	lon, err2 := strconv.ParseFloat(r.FormValue("lon"), 64)

	if err1 != nil || err2 != nil || pushURL == "" {
		w.WriteHeader(400)
		w.Write([]byte("Bad Request."))
		return
	}

	if updateDeviceLocation(pushURL, lat, lon) {
		w.Write([]byte("ok"))
		return
	}

	w.WriteHeader(400)
	w.Write([]byte("Bad Request."))
	return
}

func serveSingle(pattern string, filename string) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		log.Println("serving file " + filename);
		http.ServeFile(w, r, filename)
	})
}

func main() {

	readConfig()
	openDb()

	http.HandleFunc("/device/update/", deviceLocationHandler)

	// device management
	http.HandleFunc("/device/list", deviceListHandler)
	http.HandleFunc("/device/add/", deviceAddHandler)
	http.HandleFunc("/device/delete/", deviceDeleteHandler)

	// personas
	http.HandleFunc("/auth/check", loginCheckHandler)
	http.HandleFunc("/auth/login", loginHandler)
	http.HandleFunc("/auth/logout", logoutHandler)


	serveSingle("/",                "./static/index.html")
	serveSingle("/index.html",      "./static/index.html")
	serveSingle("/install.html",    "./static/install.html")
	serveSingle("/push.html",       "./static/push.html")
	serveSingle("/manifest.webapp", "./static/manifest.webapp")
	serveSingle("/style.css",       "./static/style.css")
	serveSingle("/logos/64.png",    "./static/logos/64.png")
	serveSingle("/logos/128.png",   "./static/logos/128.png")

	serveSingle("/package.manifest",   "./static/package.manifest")
	serveSingle("/package.zip",        "./static/package.zip")


	log.Println("Listening on", gServerConfig.Hostname+":"+gServerConfig.Port)

	var err error
	if gServerConfig.UseTLS {
		err = http.ListenAndServeTLS(gServerConfig.Hostname+":"+gServerConfig.Port,
			gServerConfig.CertFilename,
			gServerConfig.KeyFilename,
			nil)
	} else {
		log.Println("This is a really unsafe way to run the push server.  Really.  Don't do this in production.")
		err = http.ListenAndServe(gServerConfig.Hostname+":"+gServerConfig.Port, nil)
	}

	log.Println("Exiting... ", err)
	closeDb()
}
