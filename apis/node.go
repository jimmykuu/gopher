package apis

import (
	"fmt"
	"github.com/jimmykuu/gopher/models"
)

type NodeList struct {
	Base
}

type Node struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (a *NodeList) Get() interface{} {
	fmt.Println(a.User)
	c := a.DB.C(models.NODES)

	var nodes []models.Node
	c.Find(nil).All(&nodes)

	var result []Node

	for _, node := range nodes {
		result = append(result, Node{
			Id:   node.Id_.Hex(),
			Name: node.Name,
		})
	}

	return result
}
