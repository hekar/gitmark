package gitmark

import (
	"fmt"
	"time"
	"log"

	"github.com/spf13/viper"
	"github.com/garyburd/redigo/redis"
	"github.com/Rafflecopter/golang-relyq/relyq"
)

func newPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxIdle: 3,
		IdleTimeout: 240 * time.Second,
		Dial: func () (redis.Conn, error) { return redis.Dial("tcp", addr) },
	}
}

func CreateRelyQ(pool *redis.Pool) *relyq.Queue {
	return relyq.NewRedisJson(pool, &relyq.Config{Prefix: "gitmark"})
}

type Task struct {
	relyq.StructuredTask

	//Id string `json:"id"`
	Owner string `json:"owner"`
	Repo string `json:"repo"`
	Title string `json:"title"`
	Url string `json:"url"`
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func addBookmark(owner, repo, title, url string) {
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Println("Recovered: ", rec)
			err, _ := rec.(error)
			log.Panic("err", err.Error())
		}
	}()

	bookmark := Bookmark{
		Title: title,
		Url: url,
	}

	githubClient, err := NewGithubClient(owner, repo)
	if err != nil {
		panic(err)
	}

	response, err := githubClient.Commit(bookmark)
	if err != nil {
		panic(err)
	}

	log.Println("Saved bookmark", bookmark, response)
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

	redisPool := newPool(":6379")
	q := CreateRelyQ(redisPool)

	var example *Task
	for {
		l := q.Listen(example)

		go func() {
			for err := range l.Errors {
				log.Printf("Received a with an error: %s", err.Error())
			}
		}()

		go func() {
			for task := range l.Tasks {
				message := task.(*Task)
				log.Printf("Received a message: %s", task.Id(), message.Title, message.Url)
				addBookmark(message.Owner, message.Repo, message.Title, message.Url)
				err := q.Finish(message)
				if err != nil {
					panic(fmt.Errorf("Fatal error config file: %s \n", err))
				}
			}
		}()
	}

	err := q.Close()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}

