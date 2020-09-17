package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/lng50k/booster-backend/models"
	"github.com/lng50k/booster-backend/config"
	"golang.org/x/crypto/ssh"
	// "golang.org/x/crypto/ssh/agent"
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"path/filepath"
    "runtime"
	"fmt"
	// "strings"
	"time"
	"encoding/json"
	"net"
	"net/http"
	"io/ioutil"
	"encoding/base64"
)

var (
    _, b, _, _ = runtime.Caller(0)
    basepath   = filepath.Dir(b)
)

type ViaSSHDialer struct {
	client *ssh.Client
}

func (self *ViaSSHDialer) Dial(addr string) (net.Conn, error) {
	return self.client.Dial("tcp", addr)
}

type WHMController struct{}

type Account struct {
	Domain    string    `json:"domain"`
	Username  string    `json:"username"`
	Password  string 	`json:"password"`
}

func (w WHMController) RetrieveAll(c *gin.Context) {
	url := "listaccts?api.version=1"

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

	// Create a CPanel account

	url := "createacct?api.version=1&username=" + account.Username + "&domain=" + account.Domain + "&bwlimit=unlimited&cgi=1&contactemail=xin.li@gazri.com&hasshell=1&cpmod=paper_lantern&password=" + account.Password

	data, err := doRequest(url, "Post")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Unable to create an account"})
		return
	}

	var res interface{}
	json.Unmarshal(data, &res)
	fmt.Println(res)
	// c.JSON(http.StatusOK, gin.H{"message": "Successfully Created"})

	message := "CPanel Account Successfully Created"

	// Create a DB

	dbPrefix := ""
	if len(account.Username) > 8 {
		dbPrefix = account.Username[0:8]
	} else {
		dbPrefix = account.Username
	}

	url = "Mysql/create_database?name=" + dbPrefix + "_db"
	data, err = doCpanelRequest(url, account.Username, account.Password, "Post")
	
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Unable to create a database"})
		return
	}
		
	json.Unmarshal(data, &res)
	fmt.Println(res)

	message += ", Database Successfully Created"

	// Create a DB user

	url = "Mysql/create_user?name=" + dbPrefix + "_dbuser&password=EcommerceHub@13"
	data, err = doCpanelRequest(url, account.Username, account.Password, "Post")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Unable to create a database user"})
		return
	}
		
	json.Unmarshal(data, &res)
	fmt.Println(res)

	message += ", Database User Successfully Created"

	// Assign created DB user to Database

	url = "Mysql/set_privileges_on_database?user=" + dbPrefix + "_dbuser&database=" + dbPrefix + "_db&privileges=ALL PRIVILEGES"
	data, err = doCpanelRequest(url, account.Username, account.Password, "Post")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Unable to add a user to database"})
		return
	}

	json.Unmarshal(data, &res)
	fmt.Println(res)

	message += ", User Successfully Added to Database"	

	// SSH Connection
	config := config.GetConfig()
	serverIP := config.GetString("server.remote_ip")
	fmt.Println("IP Address : " + serverIP)
	conn, err := models.Connect(serverIP + ":22", account.Username, account.Password)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Unable to connect via ssh"})
		return
	}
	output, err := conn.SendCommands("rm -rf /home/" + account.Username + "/public_html")
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Unable to execute ssh command"})
		return
	}

	fmt.Println(string(output))

	// Remote Clone

	sourceR := map[string]string{"remote_name": "origin", "url": "https://github.com/lng50k/m2-jumpstart.git"}
	jsonSourceR, _ := json.Marshal(sourceR)

	fmt.Println(string(jsonSourceR))
	
	url = "VersionControl/create?type=git&name=reboxit&repository_root=/home/" + account.Username + "/public_html&source_repository=" + string(jsonSourceR)

	data, err = doCpanelRequest(url, account.Username, account.Password, "Post")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Unable to clone a remote repo"})
		return
	}

	json.Unmarshal(data, &res)
	fmt.Println(res)

	message += ", Repository Successfully Cloned"

	// Write env file

	// conn, err = models.Connect(serverIP + ":22", account.Username, account.Password)
	// if err != nil {
	// 	fmt.Println(err)
	// 	c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Unable to connect via ssh"})
	// 	return
	// }

	time.Sleep(120 * time.Second)

	envPath := " >> /home/" + account.Username + "/public_html/app/etc/env.php"
	output, err = conn.SendCommands(
		"ls -la " + "/home/" + account.Username + "/public_html",
		"echo \"<?php\"" + envPath,
		"echo \"return [\"" + envPath,
			"echo \"'backend' => [\"" + envPath,
				"echo \"'frontName' => 'admin'\"" + envPath,
			"echo \"],\"" + envPath,
			"echo \"'queue' => [\"" + envPath,
				"echo \"'consumers_wait_for_messages' => 1\"" + envPath,
			"echo \"],\"" + envPath,
			"echo \"'crypt' => [\"" + envPath,
				"echo \"'key' => 'e7713aafdd788703036fbf24b575e1ec'\"" + envPath,
			"echo \"],\"" + envPath,
			"echo \"'db' => [\"" + envPath,
				"echo \"'table_prefix' => '',\"" + envPath,
				"echo \"'connection' => [\"" + envPath,
					"echo \"'default' => [\"" + envPath,
						"echo \"'host' => 'localhost',\"" + envPath,
						"echo \"'dbname' => '" + dbPrefix + "_db',\"" + envPath,
						"echo \"'username' => '" + dbPrefix + "_dbuser',\"" + envPath,
						"echo \"'password' => 'EcommerceHub@13',\"" + envPath,
						"echo \"'active' => '1'\"" + envPath,
					"echo \"]\"" + envPath,
				"echo \"]\"" + envPath,
			"echo \"],\"" + envPath,
			"echo \"'resource' => [\"" + envPath,
				"echo \"'default_setup' => [\"" + envPath,
					"echo \"'connection' => 'default'\"" + envPath,
				"echo \"]\"" + envPath,
			"echo \"],\"" + envPath,
			"echo \"'x-frame-options' => 'SAMEORIGIN',\"" + envPath,
			"echo \"'MAGE_MODE' => 'developer',\"" + envPath,
			"echo \"'session' => [\"" + envPath,
				"echo \"'save' => 'files'\"" + envPath,
			"echo \"],\"" + envPath,
			"echo \"'cache' => [\"" + envPath,
				"echo \"'frontend' => [\"" + envPath,
					"echo \"'default' => [\"" + envPath,
						"echo \"'id_prefix' => '2c1_'\"" + envPath,
					"echo \"],\"" + envPath,
					"echo \"'page_cache' => [\"" + envPath,
						"echo \"'id_prefix' => '2c1_'\"" + envPath,
					"echo \"]\"" + envPath,
				"echo \"]\"" + envPath,
			"echo \"],\"" + envPath,
			"echo \"'lock' => [\"" + envPath,
				"echo \"'provider' => 'db',\"" + envPath,
				"echo \"'config' => [\"" + envPath,
					"echo \"'prefix' => null\"" + envPath,
				"echo \"]\"" + envPath,
			"echo \"],\"" + envPath,
			"echo \"'cache_types' => [\"" + envPath,
				"echo \"'config' => 1,\"" + envPath,
				"echo \"'layout' => 1,\"" + envPath,
				"echo \"'block_html' => 1,\"" + envPath,
				"echo \"'collections' => 1,\"" + envPath,
				"echo \"'reflection' => 1,\"" + envPath,
				"echo \"'db_ddl' => 1,\"" + envPath,
				"echo \"'compiled_config' => 1,\"" + envPath,
				"echo \"'eav' => 1,\"" + envPath,
				"echo \"'customer_notification' => 1,\"" + envPath,
				"echo \"'config_integration' => 1,\"" + envPath,
				"echo \"'config_integration_api' => 1,\"" + envPath,
				"echo \"'google_product' => 1,\"" + envPath,
				"echo \"'full_page' => 1,\"" + envPath,
				"echo \"'config_webservice' => 1,\"" + envPath,
				"echo \"'translate' => 1,\"" + envPath,
				"echo \"'vertex' => 1\"" + envPath,
			"echo \"],\"" + envPath,
			"echo \"'downloadable_domains' => [\"" + envPath,
				"echo \"'hiiauto.gazri.com'\"" + envPath,
			"echo \"],\"" + envPath,
			"echo \"'install' => [\"" + envPath,
				"echo \"'date' => 'Thu, 09 Jul 2020 20:28:21 +0000'\"" + envPath,
			"echo \"]\"" + envPath,
		"echo \"];\"" + envPath,
	)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Unable to wrtie env file"})
		return
	}

	

	fmt.Println(string(output))

	message += ", Env File Successfully Created"

	fmt.Println(basepath)
	fmt.Println(dbPrefix + "_dbuser:EcommerceHub@13@tcp(" + serverIP + ":3306)/" + dbPrefix + "_db")

	mysql.RegisterDial("mysql+tcp", (&ViaSSHDialer{conn.Client}).Dial)

	db, err := sql.Open("mysql", dbPrefix + "_dbuser:EcommerceHub@13@mysql+tcp(localhost:3306)/" + dbPrefix + "_db?multiStatements=true&maxAllowedPacket=419430400")
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Unable to establish database connection"})
		return
	}

	defer db.Close()

	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
		fmt.Println(err.Error()) // proper error handling instead of panic in your app
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Unable to ping database"})
		return
	}
	
	file, err := ioutil.ReadFile(basepath + "/../hiiauto_db.sql")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Unable to open a sql file"})
		return
	}

	// requests := strings.Split(string(file), ";\n")

	result, err := db.Exec(string(file))

	if err != nil {
		fmt.Println(err)			
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Unable to import database" })
		return
	}
	result, err = db.Exec("SET FOREIGN_KEY_CHECKS=0;UPDATE `store` SET store_id = 0 WHERE code='admin';UPDATE `store_group` SET group_id = 0 WHERE name='Default';UPDATE `store_website` SET website_id = 0 WHERE code='admin';UPDATE `customer_group` SET customer_group_id = 0 WHERE customer_group_code='NOT LOGGED IN';SET FOREIGN_KEY_CHECKS=1;")

	if err != nil {
		fmt.Println(err)			
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Unable to reset store ids" })
		return
	}

	// for _, request := range requests {
	// 	result, err := db.Exec(request)

	// 	if err != nil {
	// 		fmt.Println(err)
	// 		fmt.Println(request)
	// 		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Unable to execute " + request })
	// 		return
	// 	}

	// 	fmt.Println(result)
	// }

	fmt.Println(result)

	message += ", Database Successfully Imported"

	projectPath := "/home/" + account.Username + "/public_html"

	output, err = conn.SendCommands(
		"php " + projectPath + "/bin/magento setup:store-config:set --base-url=\"http://" + account.Domain + "/\"", 
		"php " + projectPath + "/bin/magento setup:store-config:set --base-url-secure=\"https://" + account.Domain + "/\"",
	)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Unable to execute domain change command via ssh"})
		return
	}

	fmt.Println(string(output))

	output, err = conn.SendCommands(
		"cd " + projectPath, 
		"./deploy.sh", 
		"find . -type f -exec chmod 644 {} \\;", 
		"find . -type d -exec chmod 755 {} \\;", 
		"find ./var -type d -exec chmod 777 {} \\;", 
		"find ./pub/media -type d -exec chmod 777 {} \\;", 
		"find ./pub/static -type d -exec chmod 777 {} \\;", 
		"chmod 777 ./app/etc",
		"chmod 644 ./app/etc/*.xml",
	)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Unable to execute deploy command via ssh"})
		return
	}

	fmt.Println(string(output))

	c.JSON(http.StatusOK, gin.H{"message": message})
}



func (w WHMController) Delete(c *gin.Context) {
	username := c.Params.ByName("username")
	
	url := "removeacct?user=" + username

	data, err := doRequest(url, "Delete")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Unable to delete an account"})
		return
	}

	var res interface{}
	json.Unmarshal(data, &res)
	fmt.Println(res)
	c.JSON(http.StatusOK, gin.H{"message": "Successfully Deleted"})
	
}

func doRequest(actionUrl string, method string) ([]byte, error) {
	url := "https://gazri.net:2087/json-api/" + actionUrl
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

func doCpanelRequest(actionUrl string, username string, password string, method string) ([]byte, error) {
	url := "https://gazri.net:2083/execute/" + actionUrl
	req, err := http.NewRequest(method, url, nil)

	req.Header.Add("Authorization", "Basic " + base64.StdEncoding.EncodeToString([]byte(username + ":" + password)))
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