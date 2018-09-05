package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var addr = flag.String("addr", ":1234", "http service address")
var dbconnection = ""

//ResultETH struct
type ResultETH struct {
	Private string
	Public  string
	Address string
}

//ETHResponse struct
type ETHResponse struct {
	Result ResultETH
}

//CryptoConfig struct
type CryptoConfig struct {
	Cid        int    `json:"cid"`
	CryptoName string `json:"crypto_name"`
	Endpoint   string `json:"endpoint"`
	Params     string `json:"params"`
}

//Configuration struct
type Configuration struct {
	Port             string `json:"Port"`
	DatabaseAddress  string `json:"DATABASE_ADDRESS"`
	DatabaseName     string `json:"DATABASE_NAME"`
	DatabaseUser     string `json:"DATABASE_USER"`
	DatabasePassword string `json:"DATABASE_PASSWORD"`
	APIBlockcypher   string `json:"API_BLOCKCYPHER"`
}

var configuration = Configuration{}

func main() {
	file, err := os.Open("./config/config.json")
	if err != nil {
		log.Fatal(err)
		return
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&configuration)
	fmt.Print(`Server listening on 0.0.0.0 port ` + configuration.Port + `
Initializing Firebase Configuration
------------------------------------------------------------------------------
PORT: ` + configuration.Port + `
DatabaseAddress: ` + configuration.DatabaseAddress + `
DatabaseName: ` + configuration.DatabaseName + `
DatabaseUser: ` + configuration.DatabaseUser + `
DatabasePassword: ` + configuration.DatabasePassword + `
APIBlockcypher: ` + configuration.APIBlockcypher + `
------------------------------------------------------------------------------
Press CTRL+C to exit
`)

	dbconnection = configuration.DatabaseUser + ":" + configuration.DatabasePassword + "@tcp(" + configuration.DatabaseAddress + ")/" + configuration.DatabaseName
	initDatabase(configuration)
	if err != nil {
		log.Fatal(err)
		return
	}

	flag.Parse()
	http.HandleFunc("/v1/create/address/", addHandler)
	if err := http.ListenAndServe(configuration.Port, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	//////////////////

	//////////////////
}
func initDatabase(configuration Configuration) {
	db, err := sql.Open("mysql", dbconnection)

	if err != nil {
		log.Fatal(err)
		return
	}

	_, err = db.Exec("SELECT * FROM `crypto_config`")

	if err != nil {
		go func() {
			creatMemberTable()
		}()
		go func() {
			log.Println("Create Table crypto_config Completed!")
		}()

	}

}

func creatMemberTable() {
	log.Println("Creating table crypto_config....")
	db, err := sql.Open("mysql", dbconnection)
	_, err = db.Exec("CREATE TABLE `crypto_config` (`cid` int(11) NOT NULL,`crypto_name` varchar(10) NOT NULL,`endpoint` text NOT NULL,`params` text NOT NULL ) ENGINE=InnoDB DEFAULT CHARSET=utf8;")

	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("Alter table crypto_config Add PRIMARY KEY....")
	_, err = db.Exec("ALTER TABLE `crypto_config` ADD PRIMARY KEY (`cid`);")

	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("Alter table crypto_config Modify KEY....")
	_, err = db.Exec("ALTER TABLE `crypto_config` MODIFY `cid` int(11) NOT NULL AUTO_INCREMENT;")

	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("Commit....")
	_, err = db.Exec("COMMIT;")

	if err != nil {
		log.Fatal(err)
		return
	}
	insert, err := db.Query("INSERT INTO `crypto_config` (`cid`, `crypto_name`, `endpoint`, `params`) VALUES (1, 'BTC', 'https://api.blockcypher.com/v1/btc/main/wallets/hd?token=" + configuration.APIBlockcypher + "', '[\"name\",\"extended_public_key\"]'),(2, 'ETH', 'https://api.blockcypher.com/v1/eth/main/addrs?token=" + configuration.APIBlockcypher + "', '');")
	if err != nil {
		panic(err.Error())
	}
	// be careful deferring Queries if you are using transactions
	defer insert.Close()
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	const feedbackPath = "/v1/create/address/"
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var slug string
	if strings.HasPrefix(r.URL.Path, feedbackPath) {
		slug = r.URL.Path[len(feedbackPath):]
	}
	fmt.Println("the slug is: ", slug)

	result := &ResultETH{}
	url := getParamsFromSlug(slug).Endpoint
	spaceClient := http.Client{
		Timeout: time.Second * 200,
	}

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, getErr := spaceClient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	body, readErr := ioutil.ReadAll(res.Body)
	log.Println(body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	jsonErr := json.Unmarshal(body, &result)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	output := &ETHResponse{
		Result: ResultETH{
			Private: result.Private,
			Public:  result.Private,
			Address: result.Address,
		},
	}
	b, _ := json.Marshal(output)
	setHeader(w, b)
}

func setHeader(w http.ResponseWriter, b []byte) {
	w.Header().Set("Content-Type", "application/json")
	// w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(b)
	a := string(b)
	log.Println(a)
	log.Print(b)
	log.Print("Handle is ok")
}

func getParamsFromSlug(slug string) (result *CryptoConfig) {
	db, err := sql.Open("mysql", dbconnection)
	if err != nil {
		log.Fatal(err)
		return
	}
	var cryptoConfigValue CryptoConfig
	QueryRowErr := db.QueryRow("SELECT * FROM crypto_config WHERE crypto_name = ?", slug).Scan(&cryptoConfigValue.Cid, &cryptoConfigValue.CryptoName, &cryptoConfigValue.Endpoint, &cryptoConfigValue.Params)
	if QueryRowErr != nil {
		log.Fatal(QueryRowErr)
	}
	log.Println(cryptoConfigValue.CryptoName)
	log.Println(cryptoConfigValue.Endpoint)
	log.Println(cryptoConfigValue.Params)
	return &cryptoConfigValue
}
