package service

import "goRedisLock/goProjectLearning/model"

type PostService interface {
	ListPosts() ([]*model.Post, error)
}

func ListPosts(serv PostService) ([]*model.Post, error) {
	return serv.ListPosts()
}
