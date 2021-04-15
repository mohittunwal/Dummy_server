package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
	"github.com/google/uuid"
    //"go.mongodb.org/mongo-driver"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)


type User struct {
	Id      string `json:"id"`
	Firstname string `json:"firstname"`
	Lastname string `json:"lastname"`
	Email string `json:"email"`
	Teamname string `json:"teamname"`
	Role string `json:"role"`
	Password string `json:"password"`
	Token string `json:"token"`
}


type Credentials struct {
	Email string `json:"email"`
	Password string `json:"password"`
}


type Event struct {
	Id      string `json:"id"`
	Created_by string `json:"created_by"`
	Resource string `json:"resource"`
	Title string `json:"title"`
	Start string `json:"start"`
	End string `json:"end"`
}


type Server struct {
	Ip string `json:"ip"`
	Cpu string `json:"cpu"`
	Ram string `json:"ram"`
	Storage string `json:"storage"`
}


type Environment struct {
	Name string `json:"name"`
	Server []Server `json:"server"`
}


type TempServer struct {
	Name string `json:"name"`
	Server Server `json:"server"`
}


var ctx = context.TODO()
var PORT string = ":3000"
var Client *mongo.Client = nil
const dbName = "reservationTool"


// Connect establish a connection to database
func init() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("[MONGO DB INITIATED]")
	Client = client
	return
}


func main() {

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowMethods: []string{"GET", "POST", "OPTIONS", "PUT", "PATCH", "DELETE"},
		AllowHeaders: []string{"Origin", "Content-Length", "Content-Type", "User-Agent", "Referrer", "Host",
			"Token", "X-Requested-With", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowAllOrigins:  true,
		MaxAge:           86400,
	}))

	router.GET("/", index)


	environment_routes := router.Group("/environment")
	{
		environment_routes.GET("/", getAllEnvironment)
		environment_routes.POST("/", addEnvironment)
		environment_routes.DELETE("/:id", deleteEnvironment)
	}

	server_routes := router.Group("/server")
	{
		server_routes.POST("/", addServer)
		server_routes.DELETE("/:env/:id", deleteServer)
	}

	event_routes := router.Group("/event")
	{
		event_routes.GET("/", getAllEvent)
		event_routes.POST("/", addEvent)
		event_routes.DELETE("/:id", deleteEvent)
	}

	login_routes := router.Group("/login")
	{
		login_routes.POST("/", login)
	}

	register_routes := router.Group("/register")
	{
		register_routes.POST("/", register)
	}

	user_routes := router.Group("/user")
	{
		user_routes.GET("/", getAllUser)
		user_routes.GET("/:id", getUserById)
		user_routes.PUT("/:id", updateUser)
		user_routes.DELETE("/:id", deleteUser)
		user_routes.POST("/", addUser)
	}

	if err := router.Run(PORT); err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println("[ROUTER STARTED]")

}


func index(c *gin.Context) {
	c.JSON(200, gin.H{
		"Message": "Hello World",
	})
}



// login handler function
func login(c *gin.Context) {
	var temp Credentials
	err := c.BindJSON(&temp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	email := temp.Email
	filter := bson.M{"email": email}
	var user User
	collName := "user"
	collection := Client.Database(dbName).Collection(collName)
	err = collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound,  "Invalid request")
		return
	}
	if user.Email == temp.Email && user.Password == temp. Password {
		user.Token = "fake-jwt-token"
		c.JSON(http.StatusOK, user)
		return
	}
	c.JSON(404, "Invalid login request")
	return
}

/**************************************************/



// registration handler function
func register(c *gin.Context) {
	var temp User
	err := c.BindJSON(&temp)
	temp.Id = uuid.New().String()
	temp.Role = "user"
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	collName := "user"
	collection := Client.Database(dbName).Collection(collName)
	tempResult, err := collection.InsertOne(ctx, temp)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted a User into user Collection: ", tempResult.InsertedID)
	c.JSON(http.StatusCreated, "user registered")
	return	
}

/**************************************************/



// environment handler fucntions //
func getAllEnvironment(c *gin.Context) {
	fmt.Println("Header")
	fmt.Println(c.Request.Header["Authorization"])
	collName := "environment"
	collection := Client.Database(dbName).Collection(collName)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
    		log.Fatal(err)
	}
	var environments []Environment
	if err = cursor.All(ctx, &environments); err != nil {
    		log.Fatal(err)
	}
	c.JSON(http.StatusOK,  environments)
	return
}

func addEnvironment(c *gin.Context) {
	var temp Environment
	err := c.BindJSON(&temp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	collName := "environment"
	collection := Client.Database(dbName).Collection(collName)
	tempResult, err := collection.InsertOne(ctx, temp)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted an Environment into environment Collection: ", tempResult.InsertedID)
	c.JSON(http.StatusCreated, "environment added")
	return
}

func deleteEnvironment(c *gin.Context) {
	envId := c.Param("id")
	filter := bson.M{"name": envId}
	collName := "environment"
	collection := Client.Database(dbName).Collection(collName)
	deleteResult, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Deleted Environment successfully from environmentn Collection: ", deleteResult)
	c.JSON(200, "deleted")
	return
}

/**************************************************/



// server handler functions //
func addServer(c *gin.Context) {
	var temp TempServer
	err := c.BindJSON(&temp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	filter := bson.M{"name": temp.Name}
	update := bson.M{"$push": bson.M{"server": temp.Server}}
	collName := "environment"
	collection := Client.Database(dbName).Collection(collName)
	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	fmt.Println("Added server in environment ", result)
	c.JSON(http.StatusOK, "server added")
	return
}


func deleteServer(c *gin.Context) {
	env_name := c.Param("env")
	ip := c.Param("id")
	fmt.Println(env_name)
	fmt.Println(ip)

	filter := bson.M{"name": env_name}
	update := bson.M{"$pull": bson.M{"server": bson.M{"ip": ip}}}
	collName := "environment"
	collection := Client.Database(dbName).Collection(collName)
	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	fmt.Println("Added server in environment ", result)
	c.JSON(http.StatusOK, "server deleted")
	return
}

/**************************************************/



// event handler functions //
func getAllEvent(c *gin.Context) {
	collName := "event"
	collection := Client.Database(dbName).Collection(collName)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
    		log.Fatal(err)
	}
	var events []Event
	if err = cursor.All(ctx, &events); err != nil {
    		log.Fatal(err)
	}
	c.JSON(http.StatusOK, events)
	return
}


func addEvent(c *gin.Context) {
	var temp Event
	err := c.BindJSON(&temp)
	temp.Id = uuid.New().String()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	collName := "event"
	collection := Client.Database(dbName).Collection(collName)
	tempResult, err := collection.InsertOne(ctx, temp)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted an Event into event Collection: ", tempResult.InsertedID)
	c.JSON(http.StatusCreated, "event added")
	return
}


func deleteEvent(c *gin.Context) {
	eventId := c.Param("id")
	filter := bson.M{"id": eventId}
	collName := "event"
	collection := Client.Database(dbName).Collection(collName)
	deleteResult, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Deleted Event successfully from event Collection: ", deleteResult)
	c.JSON(200, "event deleted")
	return
}

/**************************************************/


// user handler functions
func getAllUser(c *gin.Context) {
	collName := "user"
	collection := Client.Database(dbName).Collection(collName)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
    		log.Fatal(err)
	}
	var users []User
	if err = cursor.All(ctx, &users); err != nil {
    		log.Fatal(err)
	}
	c.JSON(http.StatusOK, users)
	return
}


func getUserById(c *gin.Context){
	userId := c.Param("id")
	var user User
	filter := bson.M{"id": userId}
	collName := "user"
	collection := Client.Database(dbName).Collection(collName)
	err := collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound,  "user not found!")
		return
	}
	c.JSON(http.StatusOK, user)

}

func addUser(c *gin.Context){
	var temp User
	err := c.BindJSON(&temp)
	temp.Id = uuid.New().String()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	collName := "user"
	collection := Client.Database(dbName).Collection(collName)
	tempResult, err := collection.InsertOne(ctx, temp)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted a User into user Collection: ", tempResult.InsertedID)
	c.JSON(http.StatusCreated, "user added")
	return	
}


func updateUser(c *gin.Context) {
	userId := c.Param("id")
	fmt.Println(userId)
	var temp User
	err := c.BindJSON(&temp)
	if err != nil {
		c.JSON(http.StatusNotFound, err)
		return
	} 
	temp.Id = userId
	filter := bson.M{"id": userId}
	update := bson.M{"$set": temp}
	collName := "user"
	collection := Client.Database(dbName).Collection(collName)
	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	fmt.Println("update result", result)
	c.JSON(http.StatusOK, "User updated successfully!")
}



func deleteUser(c *gin.Context) {
	userId := c.Param("id")
	filter := bson.M{"id": userId}
	collName := "user"
	collection := Client.Database(dbName).Collection(collName)
	deleteResult, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Deleted User successfully from user Collection: ", deleteResult)
	c.JSON(200, "deleted")
	return

}

/**************************************************/
