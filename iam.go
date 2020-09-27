package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/goccy/go-yaml"
	"github.com/rs/zerolog/log"
)

const groupIdPrefix = "AGPA"
const userIdPrefix = "AIDA"

var alnum = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

var iamService = &Service{
	Name: "iam",
}

func randomAlnum(l int) string {
	b := make([]byte, l)
	for i := 0; i < l; i++ {
		b[i] = alnum[rand.Intn(len(alnum))]
	}
	return string(b)
}

func generateGroupId() string {
	return groupIdPrefix + randomAlnum(17)
}

func generateUserId() string {
	return userIdPrefix + randomAlnum(17)
}

type IAMGroup struct {
	Id        string     `yaml:"id"`
	Name      string     `yaml:"name"`
	CreatedAt time.Time  `yaml:"created_at"`
	Path      string     `yaml:"path"`
	Members   []*IAMUser `yaml:"-"`
}

func (g *IAMGroup) BuildArn(accountId string) string {
	path := strings.TrimPrefix(strings.TrimRight(g.Path, "/"), "/")
	var slash string
	if path != "" {
		slash = "/"
	}
	return fmt.Sprintf("arn:aws:iam::%s:group/%s%s%s", accountId, path, slash, g.Name)
}

type IAMUser struct {
	Id        string    `yaml:"id"`
	Name      string    `yaml:"name"`
	CreatedAt time.Time `yaml:"created_at"`
	Path      string    `yaml:"path"`
}

func (u *IAMUser) BuildArn(accountId string) string {
	path := strings.TrimPrefix(strings.TrimRight(u.Path, "/"), "/")
	var slash string
	if path != "" {
		slash = "/"
	}
	return fmt.Sprintf("arn:aws:iam::%s:user/%s%s%s", accountId, path, slash, u.Name)
}

type IAMRegistry interface {
	GetGroupByName(string) (*IAMGroup, bool, error)
	GetUserByName(string) (*IAMUser, bool, error)
	GetUsers() ([]*IAMUser, error)
	GetGroups() ([]*IAMGroup, error)
}

func registerAPISet(reg IAMRegistry) {
	iamAPISet := NewAPISet("2010-05-08", "https://iam.amazonaws.com/doc/2010-05-08/")
	iamAPISet.RegisterHandler(
		&QueryOperationHandler{
			Name_: "GetGroup",
			Proto: iam.GetGroupInput{},
			Handler: func(req *aws.Request) (*aws.Response, error) {
				accountId := getAccountId(req)
				params := req.Params.(*iam.GetGroupInput)
				if params.GroupName == nil {
					return nil, &SenderFault{
						Code_:    "MissingParameter",
						Message_: "GroupName",
					}
				}
				g, ok, err := reg.GetGroupByName(*params.GroupName)
				if err != nil {
					return nil, err
				}
				if !ok {
					return nil, &SenderFault{
						Code_:    "NoSuchEntity",
						Message_: fmt.Sprintf("The group with name %s cannot be found.", *params.GroupName),
					}
				}

				out := &iam.GetGroupOutput{
					Group: &iam.Group{
						Arn:        aws.String(g.BuildArn(accountId)),
						CreateDate: aws.Time(g.CreatedAt),
						GroupId:    aws.String(g.Id),
						GroupName:  aws.String(g.Name),
						Path:       aws.String(g.Path),
					},
					IsTruncated: aws.Bool(false),
				}
				out.Users = make([]iam.User, len(g.Members))
				for i, u := range g.Members {
					out.Users[i] = iam.User{
						Arn:        aws.String(u.BuildArn(accountId)),
						CreateDate: aws.Time(u.CreatedAt),
						UserId:     aws.String(u.Id),
						UserName:   aws.String(u.Name),
						Path:       aws.String(u.Path),
					}
				}
				return &aws.Response{
					Request: &aws.Request{
						Data: out,
					},
				}, nil
			},
		},
	)
	iamAPISet.RegisterHandler(
		&QueryOperationHandler{
			Name_: "GetUser",
			Proto: iam.GetUserInput{},
			Handler: func(req *aws.Request) (*aws.Response, error) {
				accountId := getAccountId(req)
				params := req.Params.(*iam.GetUserInput)
				if params.UserName == nil {
					return nil, &SenderFault{
						Code_:    "MissingParameter",
						Message_: "UserName",
					}
				}
				u, ok, err := reg.GetUserByName(*params.UserName)
				if err != nil {
					return nil, err
				}
				if !ok {
					return nil, &SenderFault{
						Code_:    "NoSuchEntity",
						Message_: fmt.Sprintf("The user with name %s cannot be found.", *params.UserName),
					}
				}

				out := &iam.GetUserOutput{
					User: &iam.User{
						Arn:        aws.String(u.BuildArn(accountId)),
						CreateDate: aws.Time(u.CreatedAt),
						UserId:     aws.String(u.Id),
						UserName:   aws.String(u.Name),
						Path:       aws.String(u.Path),
					},
				}
				return &aws.Response{
					Request: &aws.Request{
						Data: out,
					},
				}, nil
			},
		},
	)
	iamAPISet.RegisterHandler(
		&QueryOperationHandler{
			Name_: "ListUsers",
			Proto: iam.ListUsersInput{},
			Handler: func(req *aws.Request) (*aws.Response, error) {
				accountId := getAccountId(req)

				out := &iam.ListUsersOutput{
					IsTruncated: aws.Bool(false),
				}
				users, err := reg.GetUsers()
				if err != nil {
					return nil, err
				}

				out.Users = make([]iam.User, len(users))
				for i, u := range users {
					out.Users[i] = iam.User{
						Arn:        aws.String(u.BuildArn(accountId)),
						CreateDate: aws.Time(u.CreatedAt),
						UserId:     aws.String(u.Id),
						UserName:   aws.String(u.Name),
						Path:       aws.String(u.Path),
					}
				}
				return &aws.Response{
					Request: &aws.Request{
						Data: out,
					},
				}, nil
			},
		},
	)
	iamAPISet.RegisterHandler(
		&QueryOperationHandler{
			Name_: "ListGroups",
			Proto: iam.ListGroupsInput{},
			Handler: func(req *aws.Request) (*aws.Response, error) {
				accountId := getAccountId(req)

				out := &iam.ListGroupsOutput{
					IsTruncated: aws.Bool(false),
				}
				groups, err := reg.GetGroups()
				if err != nil {
					return nil, err
				}

				out.Groups = make([]iam.Group, len(groups))
				for i, u := range groups {
					out.Groups[i] = iam.Group{
						Arn:        aws.String(u.BuildArn(accountId)),
						CreateDate: aws.Time(u.CreatedAt),
						GroupId:    aws.String(u.Id),
						GroupName:  aws.String(u.Name),
						Path:       aws.String(u.Path),
					}
				}
				return &aws.Response{
					Request: &aws.Request{
						Data: out,
					},
				}, nil
			},
		},
	)

	iamService.AddAPISet(iamAPISet)
}

type BasicIAMRegistry struct {
	groups map[string]*IAMGroup
	users  map[string]*IAMUser
}

func (reg *BasicIAMRegistry) GetGroupByName(name string) (*IAMGroup, bool, error) {
	g, ok := reg.groups[name]
	return g, ok, nil
}

func (reg *BasicIAMRegistry) GetUserByName(name string) (*IAMUser, bool, error) {
	u, ok := reg.users[name]
	return u, ok, nil
}

func (reg *BasicIAMRegistry) GetUsers() ([]*IAMUser, error) {
	users := make([]*IAMUser, 0, len(reg.users))
	for _, u := range reg.users {
		users = append(users, u)
	}
	return users, nil
}

func (reg *BasicIAMRegistry) GetGroups() ([]*IAMGroup, error) {
	groups := make([]*IAMGroup, 0, len(reg.groups))
	for _, u := range reg.groups {
		groups = append(groups, u)
	}
	return groups, nil
}

var epoch = time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)

func buildRegistryFromYAML(yamlBytes []byte) (*BasicIAMRegistry, error) {
	var y struct {
		Groups []struct {
			IAMGroup `yaml:",inline"`
			Members  []string `yaml:"members"`
		} `yaml:"groups"`
		Users []IAMUser `yaml:"users"`
	}
	err := yaml.Unmarshal(yamlBytes, &y)
	if err != nil {
		return nil, err
	}

	r := &BasicIAMRegistry{
		groups: make(map[string]*IAMGroup),
		users:  make(map[string]*IAMUser),
	}

	log.Info().Msg(fmt.Sprintf("%d groups and %d users found", len(y.Groups), len(y.Users)))

	for i, _ := range y.Users {
		u := &y.Users[i]
		r.users[u.Name] = u
		if u.CreatedAt.IsZero() {
			u.CreatedAt = epoch
		}
		if u.Id == "" {
			u.Id = generateGroupId()
		}
		if u.Path == "" {
			u.Path = "/"
		}
	}

	for i, _ := range y.Groups {
		g := &y.Groups[i]
		var members []*IAMUser
		log.Info().Msg(fmt.Sprintf("group %s has %d members", g.Name, len(g.Members)))
		for _, m := range g.Members {
			u, ok := r.users[m]
			if !ok {
				return nil, fmt.Errorf("unknown user %s among the members of %s", m, g.Name)
			}
			members = append(members, u)
		}
		if g.CreatedAt.IsZero() {
			g.CreatedAt = epoch
		}
		if g.Path == "" {
			g.Path = "/"
		}
		if g.Id == "" {
			g.Id = generateGroupId()
		}
		g.IAMGroup.Members = members
		r.groups[g.Name] = &g.IAMGroup
	}

	return r, nil
}
