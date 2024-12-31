package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"votes-api/schema"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/nitishm/go-rejson/v4"
)

const (
	RedisNilError        = "redis: nil"
	RedisDefaultLocation = "0.0.0.0:6379"
)

type cache struct {
	client  *redis.Client
	helper  *rejson.Handler
	context context.Context
}

type VotesAPI struct {
	cache
}

func NewVotesAPI(location string) (*VotesAPI, error) {

	client := redis.NewClient(&redis.Options{
		Addr: location,
	})

	ctx := context.Background()

	err := client.Ping(ctx).Err()
	if err != nil {
		log.Println("Error connecting to redis" + err.Error())
		return nil, err
	}

	jsonHelper := rejson.NewReJSONHandler()
	jsonHelper.SetGoRedisClientWithContext(ctx, client)

	return &VotesAPI{
		cache: cache{
			client:  client,
			helper:  jsonHelper,
			context: ctx,
		},
	}, nil

}

func (p *VotesAPI) PostVote(c *gin.Context) {
	// Read the payload
	var newVote schema.Vote

	if err := c.ShouldBindJSON(&newVote); err != nil {
		log.Println("Error binding JSON: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// modifying the voter and poll id to be in the right format for hyperlinks
	newVote.VoterID = "/voters/" + newVote.VoterID
	newVote.PollID = "/polls/" + newVote.PollID

	VoteJSON, err := json.Marshal(newVote)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize Vote data"})
		return
	}

	// Check if the vote id already exists
	voteKey := fmt.Sprintf("vote-%d", newVote.VoteID)

	ks, _ := p.client.Keys(p.context, "vote-*").Result()
	for _, key := range ks {
		if voteKey == key {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Vote already exists (ID is not unique)"})
			return
		}
	}

	// Check if the voter exists
	voterAPIURL := envVarOrDefault("VOTER_API_URL", "http://localhost:1080")

	vReq := voterAPIURL + newVote.VoterID
	vResp, err := http.Get(vReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	if vResp.StatusCode == http.StatusBadRequest {
		c.JSON(http.StatusBadRequest, gin.H{"error": "The voter doesn't exist."})
		return
	} else {
		log.Println("The voter exists.")
	}

	// Check if the poll exists
	pollAPIURL := envVarOrDefault("POLLS_API_URL", "http://localhost:2080")
	pReq := pollAPIURL + newVote.PollID
	pResp, err := http.Get(pReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	if pResp.StatusCode == http.StatusBadRequest {
		c.JSON(http.StatusBadRequest, gin.H{"error": "The poll doesn't exist."})
		return
	} else {
		log.Println("The poll exists.")
	}

	// Add the vote to the voter's VoteHistory
	payload := fmt.Sprintf("/votes/" + strconv.Itoa(int(newVote.VoteID)))
	targetURL := voterAPIURL + newVote.VoterID + "/history"

	requestBody := strings.NewReader(payload)
	request, _ := http.NewRequest("PUT", targetURL, requestBody)
	httpClient := &http.Client{}
	response, err := httpClient.Do(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add vote to the voter's VoteHistory"})
		return
	} else {
		log.Println(payload)
		log.Println(targetURL)
		log.Println(response)
	}

	// Add the vote to redis
	err = p.client.Set(p.context, voteKey, VoteJSON, 0).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store Vote in cache"})
		return
	} else {
		log.Println(voteKey)

		c.JSON(http.StatusOK, gin.H{"message": "Vote added to cache successfully"})
	}
}

func (p *VotesAPI) GetAllVotes(c *gin.Context) {
	var voteList []schema.Vote
	var voteItem schema.Vote

	pattern := "vote-*"
	ks, _ := p.client.Keys(p.context, pattern).Result()

	for _, key := range ks {
		value, err := p.client.Get(c, key).Result()
		if err == redis.Nil {
			fmt.Printf("Key %s does not exist in Redis\n", key)
		} else if err != nil {
			log.Fatalf("Error getting key %s: %v", key, err)
		} else {
			valueBytes := []byte(value)
			err = json.Unmarshal(valueBytes, &voteItem)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not find publication in cache with id=" + key})
			}

			voteList = append(voteList, voteItem)
		}
	}

	c.JSON(http.StatusOK, voteList)
}

func (p *VotesAPI) GetVoteByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No vote ID provided"})
		return
	}

	var voteItem schema.Vote
	value, err := p.client.Get(c, "vote-"+id).Result()
	if err == redis.Nil {
		fmt.Printf("Key %s does not exist in Redis\n", id)
	} else if err != nil {
		log.Fatalf("Error getting key %s: %v", id, err)
	} else {
		valueBytes := []byte(value)
		err = json.Unmarshal(valueBytes, &voteItem)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not find vote in cache with id=" + id})
		}

		c.JSON(http.StatusOK, voteItem)
	}
}

// Helper function:
func envVarOrDefault(envVar string, defaultVal string) string {
	envVal := os.Getenv(envVar)
	if envVal != "" {
		return envVal
	}
	return defaultVal
}
