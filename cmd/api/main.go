package main 

import ( 
	"fmt"
	"net/http" 
) 

func main() { 
	http.HandleFunc("/jobs", handleJobSubmission) 
	fmt.Println("API service listening on :8080") 
	http.ListenAndServe(":8080", nil) 
} 
	
	
func handleJobSubmission(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received a job submission!") 
	w.WriteHeader(http.StatusAccepted) 
	fmt.Fprintln(w, "Job received") 
}
