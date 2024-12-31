package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"poll-api/schema"

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

type PollAPI struct {
	cache
}

func NewPollAPI(location string) (*PollAPI, error) {
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

	return &PollAPI{
		cache: cache{
			client:  client,
			helper:  jsonHelper,
			context: ctx,
		},
	}, nil
}

func (p *PollAPI) GetPollByID(c *gin.Context) {
	pubid := c.Param("id")
	if pubid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No publication ID provided"})
		return
	}

	var pollItem schema.Poll
	value, err := p.client.Get(c, "poll-"+pubid).Result()
	if err == redis.Nil {
		msg := fmt.Sprintf("Key %s does not exist in Redis\n", pubid)
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	} else if err != nil {
		msg := fmt.Sprintf("Error getting key %s: %v", pubid, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	} else {
		valueBytes := []byte(value)
		err = json.Unmarshal(valueBytes, &pollItem)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not find publication in cache with id=" + pubid})
		}

		c.JSON(http.StatusOK, pollItem)
		return
	}
}

func (p *PollAPI) GetAllPolls(c *gin.Context) {
	var pollList []schema.Poll
	var pollItem schema.Poll

	//Lets query redis for all of the items
	pattern := "poll-*"
	ks, _ := p.client.Keys(p.context, pattern).Result()
	for _, key := range ks {

		value, err := p.client.Get(c, key).Result()
		if err == redis.Nil {
			fmt.Printf("Key %s does not exist in Redis\n", key)
		} else if err != nil {
			log.Fatalf("Error getting key %s: %v", key, err)
		} else {
			fmt.Printf("Value for key %s: %s\n", key, value)
			valueBytes := []byte(value)
			err = json.Unmarshal(valueBytes, &pollItem)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not find publication in cache with id=" + key})
			}

			pollList = append(pollList, pollItem)
		}

	}

	c.JSON(http.StatusOK, pollList)
}

func (p *PollAPI) PostPoll(c *gin.Context) {
	var newPoll schema.Poll

	if err := c.ShouldBindJSON(&newPoll); err != nil {
		log.Println("Error binding JSON: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	pollJSON, err := json.Marshal(newPoll)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize poll data"})
		return
	}

	// Generate a unique key for the poll in the cache, e.g., using a UUID

	pollKey := fmt.Sprintf("poll-%d", newPoll.PollID) // Adjust the key generation based on your needs

	// Set the key-value pair in the cache
	err = p.client.Set(p.context, pollKey, pollJSON, 0).Err() // 0 means no expiration
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store poll in cache"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Poll added to cache successfully"})
}
