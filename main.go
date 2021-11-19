package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/twilio/twilio-go/client/jwt"
)

type Response map[string]interface{}

func main() {
	fmt.Println("Starting.....")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "5678"
	}

	log.SetFormatter(&log.JSONFormatter{})

	log.Info("logs r us up and running")
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.Get("/token", TwilioTokenHandler)
	router.Post("/voice", TwilioVoiceTwimlHandler)

	FileServer(router)

	err = http.ListenAndServe(fmt.Sprintf(":%s", port), router)
	if err != nil {
		panic(err)
	}
	log.WithField("port", port).Info("http ListenAndServe")

}

func FileServer(router *chi.Mux) {
	root := "./public"
	fs := http.FileServer(http.Dir(root))

	router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat(root + r.RequestURI); os.IsNotExist(err) {
			http.StripPrefix(r.RequestURI, fs).ServeHTTP(w, r)
		} else {
			fs.ServeHTTP(w, r)
		}
	})
}

// https://jojnts.loca.lt/voice
func TwilioVoiceTwimlHandler(w http.ResponseWriter, r *http.Request) {
	// TODO read number from params
	log.WithField("To", r.URL.Query().Get("To")).Info("voice handler")
	number := os.Getenv("TWILIO_CALLER_ID")
	twiml := fmt.Sprintf(
		`<?xml version="1.0" encoding="UTF-8"?>
<Response>
<Dial callerId="%s">
<Number>+46707402392</Number>
</Dial>
</Response>`, number)

	w.Write([]byte(twiml))

}

func TwilioTokenHandler(w http.ResponseWriter, r *http.Request) {
	accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
	apiKey := os.Getenv("API_KEY")
	apiSecret := os.Getenv("API_SECRET")
	outgoingApplicationSid := os.Getenv("TWILIO_TWIML_APP_SID")

	id := rand.Int()
	identity := fmt.Sprintf("%d", id)

	params := jwt.AccessTokenParams{
		AccountSid:    accountSid,
		SigningKeySid: apiKey,
		Secret:        apiSecret,
		Identity:      identity,
	}

	jwtToken := jwt.CreateAccessToken(params)
	voiceGrant := &jwt.VoiceGrant{
		Incoming: jwt.Incoming{Allow: false},
		Outgoing: jwt.Outgoing{
			ApplicationSid: outgoingApplicationSid,
		},
	}

	jwtToken.AddGrant(voiceGrant)
	token, err := jwtToken.ToJwt()
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("Error: %v", err)))
		return
	}

	response := &Response{
		"identity": identity,
		"token":    token,
	}

	bytes, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("Error: %v", err)))
		return
	}
	w.Write(bytes)
}
