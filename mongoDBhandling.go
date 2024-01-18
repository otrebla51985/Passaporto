package main

func LogToDatabase(message string) {
	/*
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered from panic inside LogToDatabase:", r)
			}
		}()

		mongoURL := os.Getenv("MONGO_URL")

		client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURL))
		if err != nil {
			log.Fatal(err)
		}
		defer client.Disconnect(context.Background())

		collection := client.Database("passaporto").Collection("passaportoLogs")

		_, err = collection.InsertOne(context.Background(), bson.D{
			{"timestamp", time.Now()},
			{"message", message},
		})
		if err != nil {
			log.Fatal(err)
		}

	*/
}
