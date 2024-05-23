package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type PushReport struct {
	ProjectID string `bson:"projectid" json:"projectid"`
	Username  string `bson:"username" json:"username"`
	Email     string `bson:"email,omitempty" json:"email,omitempty"`
	Repo      string `bson:"repo" json:"repo"`
	Ref       string `bson:"ref" json:"ref"`
	Message   string `bson:"message" json:"message"`
	Modified  string `bson:"modified,omitempty" json:"modified,omitempty"`
}

type Project struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Secret      string             `bson:"secret" json:"secret"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Owner       Userdomyikado      `bson:"owner" json:"owner"`
	Member      Userdomyikado      `bson:"member" json:"member"`
	Closed      bool               `bson:"closed,omitempty" json:"closed,omitempty"`
}

type Userdomyikado struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Name            string             `bson:"name" json:"name"`
	PhoneNumber     string             `bson:"phonenumber" json:"phonenumber"`
	Email           string             `bson:"email,omitempty" json:"email,omitempty"`
	GithubUsername  string             `bson:"githubusername,omitempty" json:"githubusername,omitempty"`
	GitlabUsername  string             `bson:"gitlabusername,omitempty" json:"gitlabusername,omitempty"`
	GitHostUsername string             `bson:"githostusername,omitempty" json:"githostusername,omitempty"`
}

type Task struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	ProjectID string             `bson:"projectid" json:"projectid"`
	Name      string             `bson:"name" json:"name"`
	PIC       Userdomyikado      `bson:"pic" json:"pic"`
	Done      bool               `bson:"done,omitempty" json:"done,omitempty"`
}