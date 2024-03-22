package controllers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"zjuici.com/tablegpt/jkpmanager/src/common"
	"zjuici.com/tablegpt/jkpmanager/src/storage"
)

func GetKernelsHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	sessionClient, ok := r.Context().Value(common.SessionClientKey).(storage.SessionClient)
	if !ok {
		http.Error(w, "Failed to get database connection", http.StatusInternalServerError)
		return
	}
	sessionList, err := sessionClient.GetSessions()
	if err != nil {
		log.Println("Cannot get session lists: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(sessionList)
	if err != nil {
		log.Panicln("Cannot marshal sessionList")
	}

	w.Write(res)

}

func GetKernelByIdHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	kernelId, ok := vars["kernelId"]
	if !ok {
		http.Error(w, "Failed to get kernelId", http.StatusInternalServerError)
		return
	}

	sessionClient, ok := r.Context().Value(common.SessionClientKey).(storage.SessionClient)

	if !ok {
		http.Error(w, "Failed to get database connection", http.StatusInternalServerError)
		return
	}
	session, err := sessionClient.GetSessionByID(kernelId)

	if err != nil {
		log.Printf("Failed to get session by id: %v", kernelId)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(session)
	if err != nil {
		log.Panicln("Cannot marshal sessionList")
	}

	w.Write(res)

}

func DeleteKernelsHandler(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Failed to read request body: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var kernelIDs []string

	err = json.Unmarshal(body, &kernelIDs)
	if err != nil {
		log.Println("Failed to unmarshal request body: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sessionClient, ok := r.Context().Value(common.SessionClientKey).(storage.SessionClient)

	if !ok {
		http.Error(w, "Failed to get database connection", http.StatusInternalServerError)
		return
	}
	_, err = sessionClient.DeleteSessionByIDS(kernelIDs)
	if err != nil {
		log.Println("Failed to delete session: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)

}

func PostKernelByIdHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	kernelId, ok := vars["kernelId"]
	if !ok {
		log.Println("Failed to get kernelId from vars")
		http.Error(w, "Failed to get kernelId", http.StatusInternalServerError)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Failed to read request body: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sessionClient, ok := r.Context().Value(common.SessionClientKey).(storage.SessionClient)

	if !ok {
		http.Error(w, "Failed to get database connection", http.StatusInternalServerError)
		return
	}

	err = sessionClient.SaveSession(kernelId, body)
	if err != nil {
		log.Println("Failed to save session: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)

}
