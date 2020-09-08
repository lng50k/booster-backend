package controllers

import (
	"github.com/gin-gonic/gin"
	"fmt"
	"encoding/json"
	"net/http"
	"io/ioutil"
)

type WHMController struct{}

type Account struct {
	Domain    string    `json:"domain"`
	Username  string    `json:"username"`
	Password  string 	`json:"password"`
}

func (w WHMController) RetrieveAll(c *gin.Context) {
	url := "https://gazri.net:2087/json-api/listaccts?api.version=1"

	data, err := doRequest(url, "Get")
	if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Unable to fetch accounts data"})
	}

	var res interface{}
	json.Unmarshal(data, &res)

	c.JSON(http.StatusOK, gin.H{"message": "CPanel Accounts fetched!", "accounts": res})
}

func (w WHMController) Create(c *gin.Context) {
	account := Account{}

	err := c.BindJSON(&account)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Request Body is invalid!"})
		return
	}
	url := "https://gazri.net:2087/json-api/createacct?api.version=1&username=" + account.Username + "&domain=" + account.Domain + "&bwlimit=unlimited&cgi=1&contactemail=info@gazri.com&cpmod=paper_lantern&password=" + account.Password

	data, err := doRequest(url, "Post")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Unable to create an account"})
		return
	}

	var res interface{}
	json.Unmarshal(data, &res)
	fmt.Println(res)
	c.JSON(http.StatusOK, gin.H{"message": "Successfully created"})
	
}

func doRequest(url string, method string) ([]byte, error) {
	req, err := http.NewRequest(method, url, nil)

	req.Header.Add("Authorization", "whm root:EKDILJAVOIC2A159GX3F027T79NWXVHP")
	if err != nil {
		fmt.Errorf("Error when try to create the request", "HTTPMethod", method, "Request", req, "url", url, "error", err)
	}

	client := &http.Client{}
	response, err := client.Do(req)
	defer response.Body.Close()

	if err != nil {
        return nil, err
	}
	if response.StatusCode != 200 {
        return nil, fmt.Errorf("%s: %d", url, response.StatusCode)
    }
	

	return ioutil.ReadAll(response.Body)
}