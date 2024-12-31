# Poll Voting System APIs

## Description
This project is a containerized application built with **Go**, **Docker**, and **Redis** that implements a voting system with three microservices: Poll, Voter, and Votes APIs. Each API manages a distinct functionality, interacting with Redis for caching and data storage.

### Architecture:
```
               ┏━━━━━━━━━━━┓               
      ┌───────▶┃ Votes API ┃◀────────┐     
      ▼        ┗━━━━━━━━━━━┛         ▼     
┌───────────┐        │         ┌──────────┐
│ Voter API │        │         │ Poll API │
└───────────┘        │         └──────────┘
      │              │                │    
      ▼              ▼                ▼    
┌─────────────────────────────────────────┐
│              Cache (Redis)              │
└─────────────────────────────────────────┘
```

### Features
- **Poll API** allows us to register new polls.
- **Voter API** allows us to create new voters without prior votes.
- **Votes API** allows us to create new votes when provided with existing voters and polls

### Tech Stack
- **Go:** For developing the APIs.
- **Redis:** In-memory data store for caching.
- **Docker:** For containerization.
- **Docker Compose:** To orchestrate the services.

### Overall Structure
```
**project-directory/**</br>
|-- poll-api/</br>
|         &emsp;&emsp;|-- Dockerfile</br>
|         &emsp;&emsp;|-- build-docker.sh</br>
|         &emsp;&emsp;|-- ...</br>
|-- voter-api/</br>
|         &emsp;&emsp;|-- Dockerfile</br>
|         &emsp;&emsp;|-- build-docker.sh</br>
|         &emsp;&emsp;|-- ...</br>
|-- votes-api/ </br>
|         &emsp;&emsp;|-- Dockerfile</br>
|         &emsp;&emsp;|-- build-docker.sh</br>
|         &emsp;&emsp;|-- ...</br>
|-- docker-compose.yml</br>
|-- do_the_thing.sh</br>
|-- ...
```

The three APIs are available in the three following folders:
- poll-api
- voter-api
- votes-api

In each API folder there is a *Dockerfile* and a *build-docker.sh* script which builds the API and then the container. There is also a docker-compose.yaml file in the root directory, used for configuring and running all the containers.

## Run It
The easiest way to run the containerized APIs along with the Redis container is to use the provided script *do_the_thing.sh*. The script will use the *curl* tool (already added to the containers via Dockerfile) to send *http* requests to insert some sample data.


## Make Changes
If you need to make changes to any of the three APIs all you need to do afterward is to run:
```bash
./build-docker.sh
```
inside the corresponding API directory. This script will rebuild the go project and our container to make sure all the changes will be reflected in the container.