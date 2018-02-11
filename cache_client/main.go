package main

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"fmt"
	"os"
	"log"
	"errors"
	"google.golang.org/grpc"
	"golang.org/x/net/context"
	"github.com/spf13/cast"
	cli "gopkg.in/urfave/cli.v1"
	pb "cache/proto"
)

const address = "localhost:50051"

var client pb.CacheServiceClient
func main() {

	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close()

	client = pb.NewCacheServiceClient(conn)

	app := cli.NewApp()
	app.Name = "cache-client"
	app.Description = "CLI client for cache-server"
	app.Version = "1.0"

	app.Commands = []cli.Command{
    {
      Name:    "newcache",
			Usage:   "Create a new cache with unique name ex. newcache my-cache",
			Flags: []cli.Flag {
				cli.StringFlag{
					Name: "expire, e",
					Value: "10",
					Usage: "Default expire for items in minutes",
				},
				cli.StringFlag{
					Name: "gc-interval, gc",
					Value: "30",
					Usage: "Clear interval for garbage collection in minutes",
				},
			},
      Action:  func(c *cli.Context) error {
				var name string
				if name = c.Args().Get(0); len(name) == 0 {
					return errors.New("Cache name is required")
				}
				r, err := client.MakeCache(context.Background(), &pb.Cache{
					Name: name,
					Duration: cast.ToInt64(c.String("e")),
					Interval: cast.ToInt64(c.String("gc")),
				})
				if err != nil {
					return err
				}
				fmt.Println(r.Message)
				return nil
      },
		},
		{
      Name:    "listcache",
      Usage:   "Show the list existed caches",
      Action:  func(c *cli.Context) error {
			
				r, err := client.ListCache(context.Background(), &pb.Empty{})
				if err != nil {
					return err
				}
				if len(r.Caches) > 0 {
					for _, c := range r.Caches {
						fmt.Printf("%s\n\r", c.Name)
					}
				}
				return nil
      },
		},
		{
      Name:    "delcache",
      Usage:   "Delete cache by name",
      Action:  func(c *cli.Context) error {
				var name string
				if name = c.Args().Get(0); len(name) == 0 {
					return errors.New("Cache name is required")
				}
				r, err := client.DelCache(context.Background(), &pb.Cache{Name: name})
				if err != nil {
					return err
				}
				fmt.Println(r.Message)
				return nil
      },
		},
		{
      Name:    "cache",
			Usage:   "Some operations with cache",
			Subcommands: []cli.Command{
				{
					Name:    "add",
					Usage:   "cache add CACHENAME KEY VALUE [-d='MINUTES']",
					Flags: []cli.Flag {
						cli.StringFlag{
							Name: "duration, d",
							Value: "10",
							Usage: "Storage time in seconds",
						},
					},
					Action:  func(c *cli.Context) error {
						if c.NArg() < 3 {
							return errors.New("Need more options, use -h for help")
						}
						args := c.Args()
						r, err := client.CacheAdd(context.Background(), &pb.CacheObj{
							Cache: &pb.Cache{
								Name: args.Get(0),
							},
							Key: args.Get(1),
							Value: makeAny(args.Get(2)),
							Duration: cast.ToInt32(c.String("d")),
						})
						if err != nil {
							return err
						}
						fmt.Println(r.Message)
						return nil
					},
				},
				{
					Name:    "get",
					Usage:   "cache get CACHENAME ..KEYS",
					Action:  func(c *cli.Context) error {
						if c.NArg() < 1 {
							return errors.New("Need more options, use -h for help")
						}
						//collect keys and make special message to send them
						args := c.Args()
						pbkeys := &pb.CacheKeys{Keys: []string{}}
						for _ ,val := range args.Tail() {
							pbkeys.Keys = append(pbkeys.Keys, val)
						}
						value, _ := ptypes.MarshalAny(pbkeys)
						// make request
						r, err := client.CacheGet(context.Background(), &pb.CacheObj{
							Cache: &pb.Cache{
								Name: args.Get(0),
							},
							Value: value,
						})
						if err != nil {
							return err
						}
						fmt.Println(r.Value)
						return nil
					},
				},
				{
					Name:    "set",
					Usage:   "cache set CACHENAME KEYS VALUE [-d='MINUTES']",
					Flags: []cli.Flag {
						cli.StringFlag{
							Name: "duration, d",
							Value: "-1",
							Usage: "Storage time in seconds",
						},
					},
					Action:  func(c *cli.Context) error {
						if c.NArg() < 3 {
							return errors.New("Need more options, use -h for help")
						}
						args := c.Args()
						r, err := client.CacheSet(context.Background(), &pb.CacheObj{
							Cache: &pb.Cache{Name: args.Get(0)},
							Key: args.Get(1),
							Value: makeAny(args.Get(2)),
							Duration: cast.ToInt32(c.String("d")),
						})
						if err != nil {
							return err
						}
						fmt.Println(r.Message)
						return nil
					},
				},
				{
					Name:    "update",
					Usage:   "cache update CACHENAME KEYS VALUE [-d='MINUTES']",
					Flags: []cli.Flag {
						cli.StringFlag{
							Name: "duration, d",
							Value: "-1",
							Usage: "Storage time in seconds",
						},
					},
					Action:  func(c *cli.Context) error {
						if c.NArg() < 3 {
							return errors.New("Need more options, use -h for help")
						}
						args := c.Args()
						r, err := client.CacheUpdate(context.Background(), &pb.CacheObj{
							Cache: &pb.Cache{Name: args.Get(0)},
							Key: args.Get(1),
							Value: makeAny(args.Get(2)),
							Duration: cast.ToInt32(c.String("d")),
						})
						if err != nil {
							return err
						}
						fmt.Println(r.Message)
						return nil
					},
				},
				{
					Name:    "delete",
					Usage:   "cache delete CACHENAME KEY",
					Action:  func(c *cli.Context) error {
						if c.NArg() < 2 {
							return errors.New("Need more options, use -h for help")
						}
						args := c.Args()
						r, err := client.CacheDelete(context.Background(), &pb.CacheObj{
							Cache: &pb.Cache{Name: args.Get(0)},
							Key: args.Get(1),
						})
						if err != nil {
							return err
						}
						fmt.Println(r.Message)
						return nil
					},
				},
			},
    },
  }

	app.Run(os.Args)

}

func makeAny(val string) *any.Any {
	v := &any.Any{Value: []byte(val)}
	value, _ := ptypes.MarshalAny(v)
	return value
}