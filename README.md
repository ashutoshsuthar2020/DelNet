# DelNet
# Driver Tracking System with Tile38

## Overview

This is a Go-based driver tracking system using [Tile38](https://tile38.com/) as a geospatial database. The system allows:

- Adding and updating driver locations
- Storing store and delivery locations
- Finding the nearest driver for a given delivery location

## Technologies Used

- **Golang** - Backend implementation
- **Gin** - HTTP framework for handling API requests
- **Tile38** - Geospatial database for storing and querying locations
- **Docker** - Containerization

## Prerequisites

Ensure you have the following installed:

- Go (1.23.2 or later)
- Docker

## Installation

1. Clone the repository:
   ```sh
   git clone <repository-url>
   cd driver-tracking-system
   ```
2. Install dependencies:
   ```sh
   cd backend
   go mod tidy
   ```
3. Start the Tile38 server using Docker:
   ```sh
   docker pull tile38/tile38
   docker run -p 9851:9851 tile38/tile38
   ```
4. Build and run the application using Docker:
   ```sh
   cd backend
   docker build -t mapping-container .
   docker run -p 8089:8089 --network=host mapping-container
   ```
5. Get mongodb docker image
    ```sh
    docker pull mongo
    docker run -d -p 27017:27017 --name mongodb mongo
    ```
## API Endpoints

### 1. Add a Driver

**Endpoint:** `POST /drivers`

**Request Body:**

```json
{
  "id": "driver1",
  "lat": 37.7749,
  "lng": -122.4194
}
```

**Response:**

```json
{
  "message": "Driver added"
}
```

### 2. Update Driver Location

**Endpoint:** `POST /drivers/update`

**Request Body:**

```json
{
  "id": "driver1",
  "lat": 37.7750,
  "lng": -122.4195
}
```

**Response:**

```json
{
  "message": "Driver location updated"
}
```

### 3. Add a Delivery Location

**Endpoint:** `POST /deliveries`

**Request Body:**

```json
{
  "id": "delivery1",
  "lat": 37.7760,
  "lng": -122.4200
}
```

**Response:**

```json
{
  "message": "Delivery added"
}
```

### 4. Add a Store Location

**Endpoint:** `POST /stores`

**Response:**

```json
{
  "message": "Store location added"
}
```

(Note: The store location is static.)

### 5. Find the Nearest Driver

**Endpoint:** `GET /nearest-driver`

**Request Body:**

```json
{
  "lat": 37.7760,
  "lng": -122.4200
}
```

**Response:**

```json
{
  "driver_id": "driver1"
}
```

## Environment Variables

- `PORT` - The port on which the server runs (default: `8089`)

## Notes

- Ensure Tile38 is running before making API requests.
- The `findNearestDriver` function searches within a 10 km radius.

## License

This project is licensed under the MIT License.