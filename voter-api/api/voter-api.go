package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"voter-api/schema"

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

type VoterAPI struct {
	cache
}

func NewVoterAPI(location string) (*VoterAPI, error) {
	//Connect to redis.  Other options can be provided, but the defaults are OK
	client := redis.NewClient(&redis.Options{
		Addr: location,
	})

	ctx := context.Background()

	//Ensure that our redis connection is working
	err := client.Ping(ctx).Err()
	if err != nil {
		log.Println("Error connecting to redis" + err.Error())
		return nil, err
	}

	jsonHelper := rejson.NewReJSONHandler()
	jsonHelper.SetGoRedisClientWithContext(ctx, client)

	//Return a pointer to a new Voter struct
	return &VoterAPI{
		cache: cache{
			client:  client,
			helper:  jsonHelper,
			context: ctx,
		},
	}, nil
}

func (p *VoterAPI) PostVoter(c *gin.Context) {
	var newVoter schema.Voter

	if err := c.ShouldBindJSON(&newVoter); err != nil {
		log.Println("Error binding JSON: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	VoterJSON, err := json.Marshal(newVoter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize Voter data"})
		return
	}

	voterKey := fmt.Sprintf("voter-%d", newVoter.VoterID)

	// Check if the id already exists
	ks, _ := p.client.Keys(p.context, "voter-*").Result()
	for _, key := range ks {
		if voterKey == key {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Voter already exists (ID is not unique)"})
			return
		}
	}

	// Check if the vote history is empty
	if len(newVoter.VoteHistory) != 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "New voters cannot have previous votes."})
		return
	}

	err = p.client.Set(p.context, voterKey, VoterJSON, 0).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store Voter in cache"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Voter added to cache successfully"})
}

func (p *VoterAPI) GetVoterByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No voter ID provided"})
		return
	}

	var voterItem schema.Voter
	value, err := p.client.Get(c, "voter-"+id).Result()
	if err == redis.Nil {
		notExistMsg := fmt.Sprintf("Key %s does not exist in Redis", id)
		c.JSON(http.StatusBadRequest, gin.H{"error": notExistMsg})
		return
	} else if err != nil {
		log.Fatalf("Error getting key %s: %v", id, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error getting key"})
		return
	} else {
		valueBytes := []byte(value)
		err = json.Unmarshal(valueBytes, &voterItem)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not find voter in cache with id=" + id})
			return
		}

		c.JSON(http.StatusOK, voterItem)
		return
	}
}

func (p *VoterAPI) GetAllVoters(c *gin.Context) {
	var voterList []schema.Voter
	var voterItem schema.Voter

	pattern := "voter-*"
	ks, _ := p.client.Keys(p.context, pattern).Result()

	for _, key := range ks {
		value, err := p.client.Get(c, key).Result()
		if err == redis.Nil {
			fmt.Printf("Key %s does not exist in Redis\n", key)
		} else if err != nil {
			log.Fatalf("Error getting key %s: %v", key, err)
		} else {
			valueBytes := []byte(value)
			err = json.Unmarshal(valueBytes, &voterItem)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not find voter in cache with id=" + key})
			}

			voterList = append(voterList, voterItem)
		}
	}

	c.JSON(http.StatusOK, voterList)
}

func (p *VoterAPI) GetVoteHistory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No ID provided"})
		return
	}

	var voterItem schema.Voter
	value, err := p.client.Get(c, "voter-"+id).Result()
	if err == redis.Nil {
		fmt.Printf("Key %s does not exist in Redis\n", id)
	} else if err != nil {
		log.Fatalf("Error getting key %s: %v", id, err)
	} else {
		fmt.Printf("Value for key %s: %s\n", id, value)
		valueBytes := []byte(value)
		err = json.Unmarshal(valueBytes, &voterItem)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not find publication in cache with id=" + id})
			return
		}

		c.JSON(http.StatusOK, voterItem.VoteHistory)
	}
}

func (p *VoterAPI) PutVoteToVoteHistory(c *gin.Context) {
	// Make sure the voter exists (using the :id)
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No ID provided"})
		return
	}

	// Get the voter
	var voterItem schema.Voter

	value, err := p.client.Get(c, "voter-"+id).Result()

	if err == redis.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Voter does not exist in Redis"})
	} else if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error getting the voter"})
	} else {
		valueBytes := []byte(value)
		err = json.Unmarshal(valueBytes, &voterItem)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not find voter in cache with id=" + id})
		}

		// Read the payload from the request body (assuming it's a string)
		payload, _ := io.ReadAll(c.Request.Body)

		// Add the payload to the voter's VoteHistory
		voterItem.VoteHistory = append(voterItem.VoteHistory, string(payload))

		VoterJSON, err := json.Marshal(voterItem)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to Marshall the voter"})
			return
		}
		log.Println("Payload: \n" + string(payload))
		log.Println("VoterJSON: \n" + string(VoterJSON))

		err = p.client.Set(p.context, "voter-"+id, VoterJSON, 0).Err()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store Voter in cache"})
			return
		}
		value, err := p.client.Get(c, "voter-"+id).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store Voter in cache"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Vote added to the voter's VoteHistory successfully: " + string(value)})
		return
	}
}
