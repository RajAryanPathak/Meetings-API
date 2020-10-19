package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

//struct
type Meeting struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Title       string             `json:"title,omitempty" bson:"title,omitempty"`
	Participant string             `json:"participant,omitempty" bson:"participant,omitempty"`
	StartTime   string             `json:"starttime,omitempty" bson:"starttime,omitempty"`
	EndTime     string             `json:"endtime,omitempty" bson:"endtime,omitempty"`
	Timestamp   time.Time          `json:"timestamp,omitempty" bson:"timestamp,omitempty"`
}

type Participants struct {
	Name  string `json:"name,omitempty" bson:"name,omitempty"`
	Email string `json:"email,omitempty" bson:"email,omitempty"`
	RSVP  string `json:"rsvp,omitempty" bson:"rsvp,omitempty"`
}

func CreateMeetingEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	now := time.Now()
	var meeting Meeting
	_ = json.NewDecoder(request.Body).Decode(&meeting)
	meeting.Timestamp = now
	collection := client.Database("raj").Collection("meet")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result, _ := collection.InsertOne(ctx, meeting)
	meeting = meeting
	json.NewEncoder(response).Encode(result)
	json.NewEncoder(response).Encode(meeting)
}

func GetMeetingEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var meeting Meeting
	collection := client.Database("raj").Collection("meet")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	err := collection.FindOne(ctx, Meeting{ID: id}).Decode(&meeting)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(meeting)
}

func GetAllMeetings(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	keys, ok := request.URL.Query()["start"]
	keys1, ok := request.URL.Query()["end"]
	if !ok || len(keys[0]) < 1 {
		log.Println("Url Param 'key' is missing")
		return
	}
	st := keys[0]
	et := keys1[0]
	json.NewEncoder(response).Encode(st)
	json.NewEncoder(response).Encode(et)
	var people []Meeting
	collection := client.Database("raj").Collection("meet")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var ele Meeting
		cursor.Decode(&ele)
		people = append(people, ele)
	}
	if err := cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(people)
}
func main() {
	fmt.Println("Starting the application...")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb+srv://RAJ:Shefalijha1@cluster0.p8idi.mongodb.net/test?retryWrites=true&w=majority")
	client, _ = mongo.Connect(ctx, clientOptions)
	router := mux.NewRouter()
	router.HandleFunc("/meetings", CreateMeetingEndpoint).Methods("POST")
	router.HandleFunc("/meetings/", GetAllMeetings).Methods("GET")
	router.HandleFunc("/meeting/{id}", GetMeetingEndpoint).Methods("GET")
	http.ListenAndServe(":8080", router)
}
