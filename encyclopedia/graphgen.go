package encyclopedia

import (
	"fmt"
	"strings"

	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
	"github.com/tgascoigne/ragekit/jenkins"
	"github.com/tgascoigne/ragekit/resource/types"
)

type dbConfig struct {
	driver bolt.DriverPool
}

type DbConn struct {
	conn bolt.Conn
}

var DB dbConfig

func ConnectDb(addr string) {
	pool, err := bolt.NewDriverPool(addr, 30)
	if err != nil {
		panic(err)
	}
	DB = dbConfig{
		driver: pool,
	}
}

func NewConn() *DbConn {
	conn, err := DB.driver.OpenPool()
	if err != nil {
		panic(err)
	}

	return &DbConn{
		conn: conn,
	}
}

func (c *DbConn) Graph(nodes []Node) {
	for _, node := range nodes {
		c.GraphNode(node)
	}
}

func (c *DbConn) Close() {
	c.conn.Close()
}

func (c *DbConn) GraphNode(node Node) int64 {
	//CREATE (f:FOO {a: {a}, b: {b}, c: {c}, d: {d}, e: {e}, f: {f}, g: {g}, h: {h}})-[b:BAR]->(c:BAZ)
	stmt := fmt.Sprintf("MERGE (n:%v {%v}) RETURN ID(n)", node.Label(), createPropertyList(node.Properties()))

	//fmt.Printf("node is %#v\n", node)
	//fmt.Printf("statement is %v\n", stmt)

	conn := c.conn

	result, err := conn.QueryNeo(stmt, typeConvPropertyValues(node.Properties()))
	if err != nil {
		panic(err)
	}

	row, _, _ := result.NextNeo()
	nodeId := row[0].(int64)

	result.Close()

	for name, value := range node.Properties() {
		if value, ok := value.(jenkins.Jenkins32); ok && value != jenkins.Jenkins32(0) {
			asset := Asset{value}
			assetId := c.GraphNode(asset)
			c.GraphRelationship(nodeId, assetId, ":"+name)
		}
	}

	return nodeId
}

func (c *DbConn) nodeId(node Node) int64 {
	stmt := fmt.Sprintf("MATCH (n:%v {%v}) RETURN ID(n)", node.Label(), createPropertyList(node.Properties()))
	params := typeConvPropertyValues(node.Properties())

	result, err := c.conn.QueryNeo(stmt, params)
	if err != nil {
		panic(err)
	}

	row, _, _ := result.NextNeo()
	id := row[0].(int64)

	result.Close()

	return id
}

func (c *DbConn) GraphRelationship(fromId, toId int64, label string) {
	stmt := fmt.Sprintf(`MATCH (from) WHERE ID(from) = %v
MATCH (to) WHERE ID(to) = %v
CREATE (from)-[%v]->(to)`, fromId, toId, label)

	_, err := c.conn.ExecNeo(stmt, nil)
	if err != nil {
		panic(err)
	}
}

func typeConvPropertyValues(properties map[string]interface{}) map[string]interface{} {
	newProps := make(map[string]interface{})
	for key, value := range properties {
		newProps[key] = typeConvPropertyValue(value)
	}

	return newProps
}

func typeConvPropertyValue(value interface{}) interface{} {
	switch value := value.(type) {
	case jenkins.Jenkins32:
		return value.String()

	case types.Float32:
		return float32(value)

	case types.Vec4f:
		return []interface{}{float32(value[0]), float32(value[1]), float32(value[2]), float32(value[3])}

	case types.Unknown32:
		return uint32(value)

	default:
		if value == nil {
			return "nil"
		}
		return value
	}
}

func createPropertyList(props map[string]interface{}) string {
	results := make([]string, 0)

	for name, _ := range props {
		results = append(results, fmt.Sprintf("%v: {%v}", name, name))
	}

	return strings.Join(results, ", ")
}
