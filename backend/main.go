package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/xjem/t38c"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	tile38   *t38c.Client
	mongoDB  *mongo.Client
	dbName         = "geoDB"
	collectionName = "locations"
)

func main() {
	// Connect to Tile38
	var err error
	tile38, err = t38c.New(t38c.Config{
		Address: "localhost:9851",
		Debug:   true,
	})
	if err != nil {
		log.Fatalf("Failed to connect to Tile38: %v", err)
	}

	// Connect to MongoDB
	mongoDB, err = mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	r := gin.Default()

	// Register API endpoints
	r.POST("/drivers", addDriver)
	r.POST("/drivers/update", updateDriverLocation)
	r.POST("/deliveries", addDelivery)
	r.POST("/stores", addStore)
	r.GET("/nearest-driver", findNearestDriver)
	r.GET("/locations", getAllLocations)
	r.DELETE("/drivers", deleteDriver)
	r.DELETE("/deliveries", deleteDelivery)
	r.DELETE("/stores", deleteStore)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8089"
	}
	r.Run(":" + port)
}

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

	// Update location if driver already exists
	tile38.Keys.Set("drivers", req.ID).Point(req.Lat, req.Lng).Do(ctx)
	collection := mongoDB.Database(dbName).Collection(collectionName)
	filter := bson.M{"id": req.ID}
	update := bson.M{"$set": bson.M{"lat": req.Lat, "lng": req.Lng, "type": "driver"}}
	_, err := collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store/update driver"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Driver added/updated"})
}

func updateDriverLocation(c *gin.Context) {
	addDriver(c) // Reuse the same logic
}

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
	collection := mongoDB.Database(dbName).Collection(collectionName)
	existing := collection.FindOne(ctx, bson.M{"id": req.ID})
	if existing.Err() == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID already exists for another entity"})
		return
	}

	tile38.Keys.Set("deliveries", req.ID).Point(req.Lat, req.Lng).Do(ctx)
	_, err := collection.InsertOne(ctx, bson.M{"id": req.ID, "lat": req.Lat, "lng": req.Lng, "type": "delivery"})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store delivery"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Delivery added"})
}

func addStore(c *gin.Context) {
	ctx := context.Background()
	storeID := "store1"
	storeLat, storeLng := 37.7749, -122.4194

	collection := mongoDB.Database(dbName).Collection(collectionName)
	existing := collection.FindOne(ctx, bson.M{"id": storeID})
	if existing.Err() == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID already exists for another entity"})
		return
	}

	tile38.Keys.Set("stores", storeID).Point(storeLat, storeLng).Do(ctx)
	_, err := collection.InsertOne(ctx, bson.M{"id": storeID, "lat": storeLat, "lng": storeLng, "type": "store"})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store store location"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Store location added"})
}

func deleteDriver(c *gin.Context) {
	var req struct {
		ID string `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()

	// Remove driver from Tile38
	err := tile38.Keys.Del(ctx, "drivers", req.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete driver from Tile38"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Driver deleted"})
}

func deleteDelivery(c *gin.Context) {
	var req struct {
		ID string `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()

	// Remove delivery from Tile38
	err := tile38.Keys.Del(ctx, "deliveries", req.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete delivery from Tile38"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Delivery deleted"})
}

func deleteStore(c *gin.Context) {
	var req struct {
		ID string `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()

	// Remove store from Tile38
	err := tile38.Keys.Del(ctx, "stores", req.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete store from Tile38"})
		return
	}

	// Remove store from MongoDB
	collection := mongoDB.Database(dbName).Collection(collectionName)
	_, err = collection.DeleteOne(ctx, bson.M{"id": req.ID, "type": "store"})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete store from MongoDB"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Store deleted"})
}

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
	searchResult, err := tile38.Search.Nearby("drivers", req.Lat, req.Lng, 10000).Limit(1).Do(ctx)
	if err != nil || len(searchResult.Objects) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No drivers found"})
		return
	}

	driverID := searchResult.Objects[0].ID
	c.JSON(http.StatusOK, gin.H{"driver_id": driverID})
}

func getAllLocations(c *gin.Context) {
	ctx := context.Background()
	collection := mongoDB.Database(dbName).Collection(collectionName)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve locations"})
		return
	}

	var locations []bson.M
	if err := cursor.All(ctx, &locations); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode locations"})
		return
	}

	c.JSON(http.StatusOK, locations)
}
