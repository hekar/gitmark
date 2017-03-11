package gitmark

import (
	"fmt"
	"log"
	"time"

	"github.com/Rafflecopter/golang-relyq/relyq"
	"github.com/garyburd/redigo/redis"
	"github.com/spf13/viper"
)

func newPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", addr) },
	}
}

func createRelyQ(pool *redis.Pool, prefix string) *relyq.Queue {
	config := &relyq.Config{
		Prefix: prefix,
	}
	return relyq.NewRedisJson(pool, config)
}

type Task struct {
	relyq.StructuredTask

	Owner string `json:"owner"`
	Repo  string `json:"repo"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func addBookmark(owner, repo, title, url string) error {
	bookmark := Bookmark{
		Title: title,
		Url:   url,
	}

	githubClient, err := newGithubClient(owner, repo)
	if err != nil {
		return err
	}

	response, err := githubClient.commit(bookmark)
	if err != nil {
		return err
	}

	log.Println("Saved bookmark", bookmark, response)

	return nil
}

func setupViper() {
	viper.SetConfigName(".gitmarkrc")
	viper.AddConfigPath("$HOME")
	viper.AddConfigPath("$HOME/.config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}
}

func listenForBookmark(task *Task, listener *relyq.Listener, queue *relyq.Queue) {

	for task := range listener.Tasks {
		message := task.(*Task)
		log.Printf("Received a message: %s %s %s", task.Id(), message.Title, message.URL)
		err := addBookmark(message.Owner, message.Repo, message.Title, message.URL)
		if err != nil {
			queue.Fail(message)
			panic(fmt.Errorf("fatal error cannot save bookmark: %s", err))
		}

		err = queue.Finish(message)
		if err != nil {
			panic(fmt.Errorf("fatal error cannot complete message: %s", err))
		}
	}
}

// ListenToEvents Start event listener. Blocking call.
func ListenToEvents() {
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Println("Recovered: ", rec)
			err, _ := rec.(error)
			log.Panic("err\n", err.Error())
		}
	}()

	setupViper()

	redisPool := newPool(":6379")
	queue := createRelyQ(redisPool, "gitmark:bookmark")

	var genericTask *Task
	for {
		listener := queue.Listen(genericTask)
		go func() {
			for err := range listener.Errors {
				log.Printf("Received a with an error: %s", err.Error())
			}
		}()

		go func() {
			listenForBookmark(genericTask, listener, queue)
		}()
	}

	err := queue.Close()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}
}
