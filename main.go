package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
)

var actions = []string{"logged in", "logged out", "created record", "deleted record", "updated account"}

type logItem struct {
	action    string
	timestamp time.Time
}

type User struct {
	id    int
	email string
	logs  []logItem
}

func (u User) getActivityInfo() string {
	output := fmt.Sprintf("UID: %d; Email: %s;\nActivity Log:\n", u.id, u.email)
	for index, item := range u.logs {
		output += fmt.Sprintf("%d. [%s] at %s\n", index, item.action, item.timestamp.Format(time.RFC3339))
	}

	return output
}

func main() {
	rand.Seed(time.Now().Unix())

	startTime := time.Now()

	wg := new(sync.WaitGroup)

	const usersCount, workerCount = 100, 100

	usersJobs := make(chan int, usersCount)
	users := make(chan User, usersCount)

	// because we need wait with Add for finish usersJobs
	for i := 0; i < usersCount; i++ {
		wg.Add(1)
		usersJobs <- i
	}

	// because we need generate users
	for i := 0; i < workerCount; i++ {
		go generateUsers(usersJobs, users)
	}

	// because we need save users
	for i := 0; i < workerCount; i++ {
		go func() {
			err := saveUserInfo(users, wg)
			if err != nil {
				log.Fatal(err)
			}
		}()
	}

	// waiting for all wg.Add counter
	wg.Wait()

	fmt.Printf("DONE! Time Elapsed: %.2f seconds\n", time.Since(startTime).Seconds())
}

func saveUserInfo(users <-chan User, wg *sync.WaitGroup) error {
	for user := range users {
		fmt.Printf("WRITING FILE FOR UID %d\n", user.id)

		filename := fmt.Sprintf("users/uid%d.txt", user.id)
		file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("saveUserInfo failed: %v", err)
		}

		file.WriteString(user.getActivityInfo())
		time.Sleep(time.Second)
		wg.Done()
	}

	return nil
}

func generateUsers(jobs <-chan int, users chan<- User) {
	for job := range jobs {
		users <- User{
			id:    job + 1,
			email: fmt.Sprintf("user%d@company.com", job+1),
			logs:  generateLogs(rand.Intn(1000)),
		}
		fmt.Printf("generated user %d\n", job+1)
		time.Sleep(time.Millisecond * 100)
	}

	// for i := 0; i < count; i++ {
	// 	usersChannel <- User{
	// 		id:    i + 1,
	// 		email: fmt.Sprintf("user%d@company.com", i+1),
	// 		logs:  generateLogs(rand.Intn(1000)),
	// 	}
	// 	fmt.Printf("generated user %d\n", i+1)
	// 	time.Sleep(time.Millisecond * 100)
	// }
}

func generateLogs(count int) []logItem {
	logs := make([]logItem, count)

	for i := 0; i < count; i++ {
		logs[i] = logItem{
			action:    actions[rand.Intn(len(actions)-1)],
			timestamp: time.Now(),
		}
	}

	return logs
}
