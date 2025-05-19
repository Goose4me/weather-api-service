# Weather API Service

A Go-based weather subscription platform that allows users to subscribe to weather updates for their city and receive periodic email notifications. The service integrates with external weather APIs and supports email confirmations and unsubscriptions.

---
## Service deployed

You can find simple static html to interact with service by [url](https://weather-api-front.onrender.com/main.html):

https://weather-api-front.onrender.com/main.html

---
## Features

- **User Subscription**: Users can subscribe to receive weather updates by providing their email, city, and preferred update frequency (hourly or daily).
- **Email Notifications**: Sends confirmation emails upon subscription and periodic weather updates.
- **Weather Data Integration**: Fetches current weather data from external APIs.
- **Unsubscription**: Users can unsubscribe from the service via a unique link.
- **Scheduler**: Periodically checks and sends weather updates based on user preferences.

---

## Room for code improve
- Separate transactional functions in database repository to be able to form different transactional requests with `BaseRepository.WithTransaction`
- Change architecture to have common errors, functions in one place and reduce dependencies
- Group mail sending in chunks by cities
- Move confirmation mail sending to `mail-sender`
- ~~Add cache for weather info to reduce API calls. Redis best option, but simple map should work~~
- Improve test coverage. Now around `50%`
- Improve logic for daily sending. Probably need other goroutine and ability to set something like start point
---

## Getting Started

### Prerequisites

- [Go](https://golang.org/dl/) 1.24.3 or later
- [Docker](https://www.docker.com/get-started) (optional, for containerized deployment)

### Installation

1. **Clone the repository**:

   ```bash
   git clone https://github.com/Goose4me/weather-api-service.git
   cd weather-api-service

2. **Set up environment variables**

Create a `.env` (or copy `.env.template`) file in the root directory and define the following variables:
``` bash
WEATHER_API={{WEATHER_API}}
WEATHER_API_ADDRESS=https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric
DB_USER={{DB_USER}}
DB_PASSWORD={{DB_PASSWORD}}
DB_NAME={{DB_NAME}}
DBDSN="host=db user=%s password=%s dbname=%s port=5432 sslmode=disable TimeZone=Asia/Shanghai"
MAILSENDER_API_KEY={{MAILSENDER_API_KEY}}
MAILSENDER_EMAIL={{MAILSENDER_EMAIL}}
BASE_URL=http://localhost:8081
WEATHER_APP_BASE_URL=http://weather-app:8080/
```

3. **Deploy the application**

``` bash
docker compose --env-file .\.env -f .\deployments\docker-compose.yml up -d --build

```
The service will start on http://localhost:8081.

4. **Run tests**
``` bash
go test -v ./...
```

## API Endpoints

- `GET /api/weather?city={city}`: Get current weather in the city.

- `POST /api/subscribe`: Subscribe to weather updates.
    
- `GET /api/confirm/{token}`: Confirm email subscription.

- `GET /api/unsubscribe/{token}`: Unsubscribe from weather updates.