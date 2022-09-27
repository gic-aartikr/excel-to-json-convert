package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/FerdinaKusumah/excel2json"
)

const maxUploadSize = 10 * 1024 * 1024 // 10 mb
const destination = "test/download/"

var fileName string

func main() {
	http.HandleFunc("/excel-convert", excelToJsonConvert)

	log.Print("Server started on localhost:8080, use /upload for uploading files and /files/{fileName} for downloading")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func excelToJsonConvert(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 32 MB is the default used by FormFile()
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get a reference to the fileHeaders.
	// They are accessible only after ParseMultipartForm is called
	files := r.MultipartForm.File["file"]
	if len(files) != 1 {
		respondWithError(w, http.StatusBadRequest, "Please provide only one excel file")
		return
	}

	for _, fileHeader := range files {
		// Restrict the size of each uploaded file given size.
		// To prevent the aggregate size from exceeding
		// a specified value, use the http.MaxBytesReader() method
		// before calling ParseMultipartForm()
		if fileHeader.Size > maxUploadSize {
			http.Error(w, fmt.Sprintf("The uploaded image is too big: %s. Please use an image less than 1MB in size", fileHeader.Filename), http.StatusBadRequest)
			return
		}

		// Open the file
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer file.Close()

		buff := make([]byte, 512)
		_, err = file.Read(buff)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = os.MkdirAll(destination, os.ModePerm)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		f, err := os.Create(destination + fileHeader.Filename)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		defer f.Close()

		_, err = io.Copy(f, file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fileName = fileHeader.Filename
	}

	excelToJson(w, fileName)
}
func excelToJson(w http.ResponseWriter, file string) {

	var (
		result []*map[string]interface{}
		err    error
		//	path   = "D:/Go_Project/city-search-service/data/download/searchResult1_38_42_pm.xlsx"
		// select sheet name
		sheetName = "Sheet1"
		// select only selected field
		// if you want to show all headers just passing nil or empty list
		headers = []string{}
	)

	fmt.Println("destination+file:", destination+file)
	if result, err = excel2json.GetExcelFilePath(destination+file, sheetName, headers); err != nil {
		log.Fatalf(`unable to parse file, error: %s`, err)
	}
	for _, val := range result {
		result, _ := json.Marshal(val)
		fmt.Println(string(result))
	}
	respondWithJson(w, http.StatusAccepted, result)
}

func respondWithJson(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJson(w, code, map[string]string{"error": msg})
}
