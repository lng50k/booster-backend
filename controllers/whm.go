package controllers

import (
	"github.com/gin-gonic/gin"
	"fmt"
	"encoding/json"
	"net/http"
	"io/ioutil"
)

type WHMController struct{}

// type Todo struct {
// 	UserID    int    `json:"userId"`
// 	ID        int    `json:"id"`
// 	Title     string `json:"title"`
// 	Completed bool   `json:"completed"`
// }

func (w WHMController) RetrieveAll(c *gin.Context) {
	url := "https://gazri.net:2087/json-api/listaccts?api.version=1"
	req, err := http.NewRequest(http.MethodGet, url, nil)

	req.Header.Add("Authorization", "whm root:EKDILJAVOIC2A159GX3F027T79NWXVHP")
	if err != nil {
		fmt.Errorf("Error when try to create the request", "HTTPMethod", http.MethodGet, "Request", req, "url", url, "error", err)
	}

	client := &http.Client{}
	response, err := client.Do(req)
	// defer response.Body.Close()
	
	if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": err})
	} 

	data, _ := ioutil.ReadAll(response.Body)		
	
	var res interface{}
	json.Unmarshal(data, &res)

	c.JSON(http.StatusOK, gin.H{"message": "CPanel Accounts fetched!", "accounts": res})
}
