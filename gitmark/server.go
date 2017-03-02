package gitmark

import (
	"os"
	"fmt"
	"path"
	"log"
	"io/ioutil"
	"strings"

	"github.com/streadway/amqp"
	"github.com/spf13/viper"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func addBookmark(token, owner, repo, title, url string) {
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Println("Recovered: ", rec)
			err, _ := rec.(error)
			log.Panic("err", err.Error())
		}
	}()

	repo = strings.Replace(repo, "%2F", "/", -1)
	bookmark := Bookmark{
		Repo: repo,
		Title: title,
		Url: url,
	}

	origin := viper.GetString("RepoUrl")
	branch := viper.GetString("Branch")
	rootFolder, err := ioutil.TempDir("", "gitmark-")
	if err != nil {
		panic(err)
	}

	defer os.RemoveAll(rootFolder)

	repoFolder := path.Join(rootFolder, repo)
	fmt.Println(repoFolder)

	provider, err := CreateOrOpenRepository(repoFolder, origin, branch)
	defer provider.Free()

	root := RootFolder{
		Repo: origin,
		Path: rootFolder,
	}

	filename, err := AppendBookmark(root, bookmark)
	if err != nil {
		panic(err)
	}

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(content))

	_, err = provider.commit(bookmark)
	if err != nil {
		panic(err)
	}

	log.Println("Saved bookmark", bookmark)
}

func SetupViper() {
	viper.SetConfigName(".gitmarkrc")
	viper.AddConfigPath("$HOME")
	viper.AddConfigPath("$HOME/.config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}

func SetupServer() {
	SetupViper()

	rabbitmqUrl := viper.GetString("rabbitmq_url")
	conn, err := amqp.Dial(rabbitmqUrl)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	rabbitmqQueue := viper.GetString("rabbitmq_queue")
	msgs, err := ch.Consume(
		rabbitmqQueue, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

