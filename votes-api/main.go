package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"votes-api/api"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	hostFlag    string
	portFlag    uint
	cacheURL    string
	voterAPIURL string
	pollAPIURL  string
)

func processCmdLineFlags() {
	flag.StringVar(&hostFlag, "h", "0.0.0.0", "Listen on all interfaces")
	flag.StringVar(&cacheURL, "c", "0.0.0.0:6379", "Default cache location")
	flag.StringVar(&voterAPIURL, "http://voter-api:1080", "http://localhost:1080", "Default endpoint for voter API")
	flag.StringVar(&pollAPIURL, "http://poll-api:2080", "http://localhost:2080", "Default endpoint for poll API")
	flag.UintVar(&portFlag, "p", 3080, "Default Port")

	flag.Parse()
}

func envVarOrDefault(envVar string, defaultVal string) string {
	envVal := os.Getenv(envVar)
	if envVal != "" {
		return envVal
	}
	return defaultVal
}

func setupParms() {
	processCmdLineFlags()

	cacheURL = envVarOrDefault("REDIS_URL", cacheURL)
	voterAPIURL = envVarOrDefault("VOTER_API_URL", voterAPIURL)
	pollAPIURL = envVarOrDefault("POLLS_API_URL", pollAPIURL)
	hostFlag = envVarOrDefault("POLL_API_HOST", hostFlag)

	pfNew, err := strconv.Atoi(envVarOrDefault("ELECTION_API_PORT", fmt.Sprintf("%d", portFlag)))
	if err == nil {
		portFlag = uint(pfNew)
	}

}

func main() {
	setupParms()
	log.Println("Init/cacheURL: " + cacheURL)
	log.Println("Init/VOTERAPIURL: " + voterAPIURL)
	log.Println("Init/POLLAPIURL: " + pollAPIURL)
	log.Println("Init/hostFlag: " + hostFlag)
	log.Printf("Init/portFlag: %d", portFlag)

	apiHandler, err := api.NewVotesAPI(cacheURL)

	if err != nil {
		panic(err)
	}

	r := gin.Default()
	r.Use(cors.Default())

	r.GET("/votes", apiHandler.GetAllVotes)
	r.POST("/votes", apiHandler.PostVote)
	r.GET("/votes/:id", apiHandler.GetVoteByID)

	serverPath := fmt.Sprintf("%s:%d", hostFlag, portFlag)
	r.Run(serverPath)
}
