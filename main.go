package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/xjem/t38c"
)

var tile38 *t38c.Client

func main() {
	// Connect to Tile38
	var err error
	tile38, err = t38c.New(t38c.Config{
		Address: "localhost:9851",
		Debug:   true, // print queries to stdout
	})
	if err != nil {
		log.Fatalf("Failed to connect to Tile38: %v", err)
	}

	r := gin.Default()

	// Register API endpoints
	r.POST("/drivers", addDriver)
	r.POST("/deliveries", addDelivery)
	r.GET("/nearest-driver", findNearestDriver)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8089"
	}
	r.Run(":" + port)
}

// Add a driver to Tile38
func addDriver(c *gin.Context) {
	var req struct {
		ID  string  `json:"id"`
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	if err := tile38.Keys.Set("drivers", req.ID).Point(req.Lat, req.Lng).Do(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store driver"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Driver added"})
}

// Add a delivery location to Tile38
func addDelivery(c *gin.Context) {
	var req struct {
		ID  string  `json:"id"`
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	if err := tile38.Keys.Set("deliveries", req.ID).Point(req.Lat, req.Lng).Do(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store delivery"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Delivery added"})
}

// Find the nearest driver for a delivery
func findNearestDriver(c *gin.Context) {
	var req struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	// searchResult, err := tile38.Search.Nearby("drivers").Point(req.Lat, req.Lng, 10000).Limit(1).Do(ctx)
	searchResult, err := tile38.Search.Nearby("drivers", 33.462, -112.268, 6000).
		Where("speed", 0, 100).
		Match("truck*").
		Format(t38c.FormatPoints).
		Do(ctx)
	if err != nil || len(searchResult.Objects) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No drivers found"})
		return
	}

	driverID := searchResult.Objects[0].ID
	c.JSON(http.StatusOK, gin.H{"driver_id": driverID})
}

/*
# How to Use

## 1. Start Tile38 Server
Ensure you have Tile38 running before starting the application. You can use Docker:

```sh
docker run -d -p 9851:9851 tile38/tile38
```

Or run locally:
```sh
tile38-server
```

## 2. Run the Go Application
```sh
go run main.go
```

## 3. API Endpoints

### Add a Driver
```sh
curl -X POST "http://localhost:8089/drivers" -H "Content-Type: application/json" -d '{"id":"driver1", "lat":37.7749, "lng":-122.4194}'
```

### Add a Delivery
```sh
curl -X POST "http://localhost:8089/deliveries" -H "Content-Type: application/json" -d '{"id":"delivery1", "lat":37.7750, "lng":-122.4195}'
```

### Find the Nearest Driver
```sh
curl -X GET "http://localhost:8089/nearest-driver" -H "Content-Type: application/json" -d '{"lat":37.7750, "lng":-122.4195}'
```
*/
