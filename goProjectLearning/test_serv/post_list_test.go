package main

import (
	"goRedisLock/goProjectLearning/model"
	"goRedisLock/goProjectLearning/service"
	"testing"
)

type fakeService struct {
}

func NewFakeService() *fakeService {
	return &fakeService{}
}

func (s *fakeService) ListPosts() ([]*model.Post, error) {
	postList := []*model.Post{}
	post1 := &model.Post{
		Name:    "post1",
		Content: "content1",
	}
	post2 := &model.Post{
		Name:    "post2",
		Content: "content2",
	}
	postList = append(postList, post1, post2)
	return postList, nil
}

func TestListPosts(t *testing.T) {
	fake := NewFakeService()
	if _, err := service.ListPosts(fake); err != nil {
		t.Fatal("list posts failed")
	}
}
