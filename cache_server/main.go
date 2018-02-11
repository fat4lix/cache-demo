package main

import (
	"github.com/spf13/cast"
	"time"
	"net"
	"errors"
	"log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "cache/proto"
	"cache/cache"
	"google.golang.org/grpc/reflection"
	"github.com/golang/protobuf/ptypes"
	"strings"
	"github.com/golang/protobuf/ptypes/any"
)

const (
	port = ":50051"
)

type server struct{}

var cs *cache.CacheService

func(s *server) MakeCache(ctx context.Context, in *pb.Cache) (*pb.CacheResponse, error) {
	cs.Add(in.Name, cache.Config{
		time.Duration(in.Duration) * time.Minute,
		time.Duration(in.Interval) * time.Minute,
	})
	return &pb.CacheResponse{Message: "true"}, nil
}

func(s *server) ListCache(ctx context.Context, in *pb.Empty) (*pb.CacheList, error) {
	var l []*pb.Cache
	for _, cache := range cs.List() {
		l = append(l, &pb.Cache{Name: cache.Name})
	}
	return &pb.CacheList{Caches: l}, nil
}

func(s *server) DelCache(ctx context.Context, in *pb.Cache) (*pb.CacheResponse, error) {
	status := cs.Del(in.Name)
	if status {
		return &pb.CacheResponse{Message: "true"}, nil
	}
	return nil, errors.New("Cache with given name do not exist")
}
func(s *server) CacheAdd(ctx context.Context, in *pb.CacheObj) (*pb.CacheResponse, error) {
	if !cs.Has(in.Cache.Name) {
		return nil, errors.New("Cache not exist")
	}
	cache := cs.Get(in.Cache.Name)
	if cache.Has(in.Key) {
		return nil, errors.New("Key is already exist in the cache")
	}

	value := &any.Any{}
	
	ptypes.UnmarshalAny(in.Value, value)

	_, err := cache.Add(in.Key, value.Value, time.Duration(in.Duration) * time.Minute)

	if err != nil { 
		return nil, err
	}
	return &pb.CacheResponse{Message: "true"}, nil
}
func(s *server) CacheGet(ctx context.Context, in *pb.CacheObj) (*pb.CacheValue, error) {
	if !cs.Has(in.Cache.Name) {
		return nil, errors.New("Cache not exist")
	}
	cache := cs.Get(in.Cache.Name)
	keys := &pb.CacheKeys{}
	ptypes.UnmarshalAny(in.Value, keys)

	values, ok := cache.Get(keys.Keys)

	var stringValues []string

	if ok {
		for _, v := range values {
			stringValues = append(stringValues, cast.ToString(v))
		}
	}
	value := strings.Join(stringValues, ", ")
	return &pb.CacheValue{Value: value}, nil
}
func(s *server) CacheSet(ctx context.Context, in *pb.CacheObj) (*pb.CacheResponse, error) {
	if !cs.Has(in.Cache.Name) {
		return nil, errors.New("Cache not exist")
	}
	cache := cs.Get(in.Cache.Name)

	value := &any.Any{}
	
	ptypes.UnmarshalAny(in.Value, value)

	var duration time.Duration = -1

	if in.Duration != -1 {
		duration = time.Duration(in.Duration) * time.Minute
	}
	if cache.Has(in.Key) {
		cache.Update(in.Key, value.Value, duration)
		return &pb.CacheResponse{Message: "true"}, nil
	}
	_, err := cache.Add(in.Key, value.Value, duration)

	if err != nil { 
		return nil, err
	}
	return &pb.CacheResponse{Message: "true"}, nil
}
func(s *server) CacheUpdate(ctx context.Context, in *pb.CacheObj) (*pb.CacheResponse, error) {
	
	if !cs.Has(in.Cache.Name) {
		return nil, errors.New("Cache not exist")
	}
	cache := cs.Get(in.Cache.Name)

	value := &any.Any{}
	
	ptypes.UnmarshalAny(in.Value, value)

	var duration time.Duration = -1

	if in.Duration != -1 {
		duration = time.Duration(in.Duration) * time.Minute
	}
	if cache.Has(in.Key) {
		cache.Update(in.Key, value.Value, duration)
		return &pb.CacheResponse{Message: "true"}, nil
	}
	return nil, errors.New("Key is not exist")
}
func(s *server) CacheDelete(ctx context.Context, in *pb.CacheObj) (*pb.CacheResponse, error) {

	if !cs.Has(in.Cache.Name) {
		return nil, errors.New("Cache not exist")
	}
	cache := cs.Get(in.Cache.Name)
	if !cache.Has(in.Key) {
		return nil, errors.New("Key is not exist")
	}
	cache.Del(in.Key)

	return &pb.CacheResponse{Message: "true"}, nil
}
func main() {

	cs = cache.NewCacheService()

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s:= grpc.NewServer()
	pb.RegisterCacheServiceServer(s, &server{})
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}