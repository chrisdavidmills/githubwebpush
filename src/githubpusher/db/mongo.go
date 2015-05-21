package db

import (
	"github.com/google/go-github/github"
	"githubpusher/config"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
)

type User struct {
	Name         string
	GitHubToken  string
	Repos        []string
	PushEndpoint string
	PushEvents   []github.WebHookPayload
}

var (
	MgoSession *mgo.Session
)

func getSession() *mgo.Session {
	if MgoSession != nil {
		return MgoSession
	}

	config := config.ReadConfig()

	var err error
	MgoSession, err = mgo.Dial(config.MongoAddress)
	if err != nil {
		log.Println("Connect to mongo error: %s", err.Error())
		panic("could not connect to mongo.")
	}
	return MgoSession
}

func AddUsersToDB(user *User) {
	session := getSession()
	if session == nil {
		return
	}

	c := session.DB("githubpush").C("users")
	err := c.Insert(user)
	if err != nil {
		log.Println("AddUsersToDB error: ", err.Error())
	}
}

func RemoveUserByToken(token string) {
	session := getSession()
	if session == nil {
		return
	}

	c := session.DB("githubpush").C("users")
	err := c.Remove(bson.M{"githubtoken": token})
	if err != nil {
		log.Println("RemoveUserByToken error: ", err.Error(), token)
	}
}

// todo make the token field unique
func GetUserByToken(token string) *User {
	session := getSession()
	if session == nil {
		return nil
	}

	c := session.DB("githubpush").C("users")

	result := &User{}
	err := c.Find(bson.M{"githubtoken": token}).One(result)
	if err != nil {
		log.Println("GetUserByToken error: ", err.Error(), token)
	}
	return result
}

func GetUsersByStarredRepoURL(url string) []User {
	session := getSession()
	if session == nil {
		return nil
	}

	c := session.DB("githubpush").C("users")

	var results []User
	err := c.Find(bson.M{"repos": url}).All(&results)
	if err != nil {
		log.Println("GetUsersByStarredRepoURL error: ", err.Error(), url)
	}
	return results
}

func AddPayloadToUser(payload github.WebHookPayload, u *User) {
	session := getSession()
	if session == nil {
		return
	}

	c := session.DB("githubpush").C("users")

	err := c.Update(bson.M{"githubtoken": u.GitHubToken}, bson.M{"$push": bson.M{"pushevents": payload}})

	if err != nil {
		log.Println("Update error: ", err.Error())
	}
}
