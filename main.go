package main

import (
	// "bytes"
	// "compress/gzip"
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ShadiestGoat/ImageServerApi/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	// "github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)


var submittionCache = map[string]models.Submition{}
var userCache = map[string]models.User{}
var admin = map[string]models.User{}

func mongoUrl() string {
	return "mongodb://" + os.Getenv("USERNAME") + ":" + os.Getenv("PASSWORD") + "@" + os.Getenv("LOCATION") + "/" + os.Getenv("DBNAME") + "?readPreference=primary&authSource=" + os.Getenv("DBNAME")
}

func setupSubmittionCache(db *mongo.Database, ctx context.Context) {
	col := db.Collection("submittions")
	cur, err := col.Find(ctx, bson.D{})
	
	if err != nil {
		log.Fatal(err)
	}

	for cur.Next(ctx) {
		// var b bytes.Buffer
		// gz, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
		if err != nil {
			log.Fatal(err)
		}
		var res models.Submition
		err = cur.Decode(&res)
		if err != nil {
			log.Fatal(err)
		}
		// if _, err := gz.Write([]byte(res.Content)); err != nil {
		// 	log.Fatal(err)
		// }	
		// if err := gz.Close(); err != nil {
		// 	log.Fatal(err)
		// }
		// res.Content = b.String()
		submittionCache[res.Id] = res
	}
	// fmt.Printf("%#v\n", res.Content)
	cur.Close(ctx)
}

func setupUserCache(db *mongo.Database, ctx context.Context) {
	col := db.Collection("users")
	cur, err := col.Find(ctx, bson.D{})
	
	if err != nil {log.Fatal(err)}

	for cur.Next(ctx) {
		var res models.User
		err := cur.Decode(&res)
		if err != nil {
			log.Fatal(err)
		}
		userCache[res.Id] = res
		if res.Admin {
			admin[res.Id] = res
		}
	}

	cur.Close(ctx)
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}
	PORT := os.Getenv("PORT")
	if (len(PORT) == 0) {
		PORT = "3000"
	}
	// Time to setup cache & mongodb

	client, err := mongo.NewClient(options.Client().ApplyURI(mongoUrl()))

	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)

	err = client.Connect(ctx)
	defer cancel()
	defer client.Disconnect(ctx)

	if err != nil {
		log.Fatal(err)
	}

	db := client.Database(os.Getenv("DBNAME"))

	setupSubmittionCache(db, ctx)
	setupUserCache(db, ctx)
	
	app := fiber.New(fiber.Config{
		AppName: "Image Server",
	})

	// app.Use(cache.New())

	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	app.Use(etag.New(etag.Config{
		Weak: false,
		Next: func(c *fiber.Ctx) bool {return false},
	}))



	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Get("/rawi/:id", func(c *fiber.Ctx) error {
		paramId := c.Params("id")
		splited := strings.Split(paramId, ".")
		id := strings.Join(splited[:(len(splited)-1)], ".")
		if len(splited) == 1 {
			id = splited[0]
		}
		item, ok := submittionCache[id]
		if !ok {return c.SendStatus(404)}
		format := "webp"
		if item.Gif {format = "gif"}
		c.Type(format)
		return c.SendString(item.Content)
	})

	app.Get("/i/:id", func(c *fiber.Ctx) error {
		paramId := c.Params("id")
		splited := strings.Split(paramId, ".")
		id := strings.Join(splited[:(len(splited)-1)], ".")
		if len(splited) == 1 {
			id = splited[0]
		}
		item, ok := submittionCache[id]
		format := ".webp"
		if !ok {return c.SendStatus(404)}
		if item.Gif {format = ".gif"}
		c.Type("html")
		return c.SendString(`<!DOCTYPE html>
<html lang="en">
<head>
<title> Sick ass epic image server </title>
<meta name="viewport" content="width=device-width,initial-scale=1">
<meta property="og:title" content="Shady's image server" />
<meta property="og:image" content="/rawi/` + id + format + `" />
<meta property="og:url" content="/i/` + id + `" />
<meta property="og:description" content="Forcefully shoved onto this by ` + userCache[item.Author].Username + " on " + time.UnixMilli(item.Timestamp).String() + `" />
<meta property="twitter:title" content="Shady's image server" />
<meta property="twitter:image" content="/rawi/` + id + format + `" />
<meta name="theme-color" content="#5655b0">
<meta name="twitter:card" content="summary_large_image">
<style>
:root {
	background-color: #202124 !important;
}
*, :after, :before {
	box-sizing: border-box;
	margin: 0 !important;
}
</style></head>
<body><img style="height: 100vh; margin: 0 auto !important; display: block;" src="https://ssltest.shadygoat.eu/rawi/` + id + format +`" /></body>`)
	})

	app.Listen(":" + PORT)
}
