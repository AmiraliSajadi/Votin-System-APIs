package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"voter-api/api"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	hostFlag string
	portFlag uint
	cacheURL string
	// voterAPIURL string
	// votesAPIURL string
)

func processCmdLineFlags() {
	flag.StringVar(&hostFlag, "h", "0.0.0.0", "Listen on all interfaces")
	flag.StringVar(&cacheURL, "c", "0.0.0.0:6379", "Default cache location")
	// flag.StringVar(&voterAPIURL, "voterapi", "http://localhost:1080", "Default endpoint for voter API")
	// flag.StringVar(&votesAPIURL, "votesapi", "http://localhost:3080", "Default endpoint for votes API")
	flag.UintVar(&portFlag, "p", 1080, "Default Port")

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
	//first process any command line flags
	processCmdLineFlags()

	//now process any environment variables
	cacheURL = envVarOrDefault("REDIS_URL", cacheURL)
	// voterAPIURL = envVarOrDefault("ELECTION_VOTER_API_URL", voterAPIURL)
	// votesAPIURL = envVarOrDefault("ELECTION_VOTES_API_URL", votesAPIURL)
	hostFlag = envVarOrDefault("POLL_API_HOST", hostFlag)

	pfNew, err := strconv.Atoi(envVarOrDefault("ELECTION_API_PORT", fmt.Sprintf("%d", portFlag)))
	//only update the port if we were able to convert the env var to an int, else
	//we will use the default we got from the command line, or command line defaults
	if err == nil {
		portFlag = uint(pfNew)
	}

}

func main() {
	//this will allow the user to override key parameters and also setup defaults
	setupParms()
	log.Println("Init/cacheURL: " + cacheURL)
	// log.Println("Init/VOTERAPIURL: " + voterAPIURL)
	// log.Println("Init/VOTESAPIURL: " + votesAPIURL)
	log.Println("Init/hostFlag: " + hostFlag)
	log.Printf("Init/portFlag: %d", portFlag)

	apiHandler, err := api.NewVoterAPI(cacheURL)

	if err != nil {
		panic(err)
	}

	r := gin.Default()
	r.Use(cors.Default())

	r.GET("/voters", apiHandler.GetAllVoters)
	r.POST("/voters", apiHandler.PostVoter)
	r.GET("/voters/:id", apiHandler.GetVoterByID)
	r.PUT("/voters/:id/history", apiHandler.PutVoteToVoteHistory)
	r.GET("/voters/:id/history", apiHandler.GetVoteHistory)
	// We may need more???

	serverPath := fmt.Sprintf("%s:%d", hostFlag, portFlag)
	r.Run(serverPath)
}
