package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"zjuici.com/tablegpt/jkpmanager/src/common"
	"zjuici.com/tablegpt/jkpmanager/src/models"
)

func GetKernelsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	conn, ok := r.Context().Value(common.DBConnKey).(*pgxpool.Conn)
	if !ok {
		http.Error(w, "Failed to get database connection", http.StatusInternalServerError)
		return
	}
	var session models.Session
	var sessionList []models.Session

	rows, err := conn.Query(r.Context(), "select * from sessions;")
	if err != nil {
		log.Println("Cannot get the kernels", err)
	}

	for rows.Next() {
		rows.Scan(&session.ID, &session.KernelInfo)

		sessionList = append(sessionList, session)
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
	conn, ok := r.Context().Value(common.DBConnKey).(*pgxpool.Conn)

	if !ok {
		http.Error(w, "Failed to get database connection", http.StatusInternalServerError)
		return
	}
	var session models.Session

	err := conn.QueryRow(r.Context(), "select * from sessions where id=$1;", kernelId).Scan(&session.ID, &session.KernelInfo)
	if err != nil {
		log.Println("Cannot get the kernel "+kernelId, err)
		http.Error(w, "Cannot get the kernel: "+kernelId, http.StatusNotFound)
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
	conn, ok := r.Context().Value(common.DBConnKey).(*pgxpool.Conn)
	if !ok {
		http.Error(w, "Failed to get database connection", http.StatusInternalServerError)
		return
	}
	placeholders := make([]string, len(kernelIDs))
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf("DELETE FROM sessions WHERE id IN (%s)", strings.Join(placeholders, ","))

	args := make([]interface{}, len(kernelIDs))
	for i, id := range kernelIDs {
		args[i] = id
	}

	_, err = conn.Exec(r.Context(), query, args...)
	if err != nil {
		log.Println("Failed to delete kernels: ", err)
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

	kernelInfoJSON := pgtype.JSONB{
		Bytes:  body,
		Status: pgtype.Present,
	}

	conn, ok := r.Context().Value(common.DBConnKey).(*pgxpool.Conn)

	if !ok {
		http.Error(w, "Failed to get database connection", http.StatusInternalServerError)
		return
	}

	_, err = conn.Exec(r.Context(), "INSERT INTO sessions (id, kernel_info) VALUES ($1, $2);", kernelId, kernelInfoJSON)
	if err != nil {
		log.Println("Failed to insert kernel: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)

}
