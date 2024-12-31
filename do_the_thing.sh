#!/bin/bash

# Start Docker Compose
docker-compose up -d

# Wait for containers to start
sleep 5


### CURL COMMANDS:

# Add two polls
docker exec -it poll-api-1 curl -X POST -H "Content-Type: application/json" -d '{"PollID": 1,"PollTitle": "Color Poll","PollQuestion": "What is your favorite color?","PollOptions": [ { "PollOptionID": 1, "PollOptionText": "Red" },{ "PollOptionID": 2, "PollOptionText": "Blue" },{ "PollOptionID": 3, "PollOptionText": "Green" }] }' http://localhost:2080/polls
docker exec -it poll-api-1 curl -X POST -H "Content-Type: application/json" -d '{"PollID": 2,"PollTitle": "Poll Question Poll","PollQuestion": "How much time did you spend coming up with an interesting poll question?","PollOptions": [ { "PollOptionID": 1, "PollOptionText": "< 60s" },{ "PollOptionID": 2, "PollOptionText": "1-3 minutes" },{ "PollOptionID": 3, "PollOptionText": "too much" }] }' http://localhost:2080/polls

# Add a voter
docker exec -it voter-api-1 curl -X POST -H "Content-Type: application/json" -d '{"VoterID": 1,"FirstName": "John","LastName": "Doe","VoteHistory": []}' http://localhost:1080/voters
docker exec -it voter-api-1 curl -X POST -H "Content-Type: application/json" -d '{"VoterID": 2,"FirstName": "Amirali","LastName": "Sajadi","VoteHistory": []}' http://localhost:1080/voters
docker exec -it voter-api-1 curl -X POST -H "Content-Type: application/json" -d '{"VoterID": 3,"FirstName": "Aaron","LastName": "Swartz","VoteHistory": []}' http://localhost:1080/voters

# Add a vote by the above voter in the above poll
docker exec -it votes-api-1 curl -X POST -H "Content-Type: application/json" -d '{"VoteID": 1,"VoterID": "1","PollID": "1","VoteValue": 1}' http://localhost:3080/votes
docker exec -it votes-api-1 curl -X POST -H "Content-Type: application/json" -d '{"VoteID": 2,"VoterID": "2","PollID": "1","VoteValue": 2}' http://localhost:3080/votes
docker exec -it votes-api-1 curl -X POST -H "Content-Type: application/json" -d '{"VoteID": 3,"VoterID": "2","PollID": "2","VoteValue": 3}' http://localhost:3080/votes
docker exec -it votes-api-1 curl -X POST -H "Content-Type: application/json" -d '{"VoteID": 4,"VoterID": "3","PollID": "2","VoteValue": 1}' http://localhost:3080/votes

# Adding an existing voter - This user will not be added
docker exec -it voter-api-1 curl -X POST -H "Content-Type: application/json" -d '{"VoterID": 3,"FirstName": "Zhijie","LastName": "Wang","VoteHistory": []}' http://localhost:1080/voters