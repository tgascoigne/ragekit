package encyclopedia

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/tgascoigne/ragekit/jenkins"
	"github.com/tgascoigne/ragekit/resource/types"
	_ "gopkg.in/cq.v1"
)

type dbConfig struct {
	addr string
}

type DbConn struct {
	conn *sql.DB
}

var DB dbConfig

func ConnectDb(addr string) {
	DB = dbConfig{
		addr: addr,
	}
}

func NewConn() *DbConn {
	db, err := sql.Open("neo4j-cypher", DB.addr)
	if err != nil {
		log.Fatal(err)
	}

	return &DbConn{
		conn: db,
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
	properties := typeConvPropertyValues(node.Properties())
	propsFmt, propsList := createPropertyList(properties)
	stmt := fmt.Sprintf("MERGE (n:%v {%v}) RETURN ID(n)", node.Label(), propsFmt)

	//fmt.Printf("node is %#v\n", node)
	//fmt.Printf("statement is %v\n", stmt)

	conn := c.conn
	rows, err := conn.Query(stmt, propsList...)
	if err != nil {
		panic(err)
	}

	var nodeId int64
	rows.Next()
	err = rows.Scan(&nodeId)
	if err != nil {
		panic(err)
	}

	rows.Close()

	for name, value := range node.Properties() {
		if value, ok := value.(jenkins.Jenkins32); ok && value != jenkins.Jenkins32(0) {
			asset := Asset{value}
			assetId := c.GraphNode(asset)
			c.GraphRelationship(nodeId, assetId, ":"+name)
		}
	}

	return nodeId
}

func (c *DbConn) GraphRelationship(fromId, toId int64, label string) {
	stmt := fmt.Sprintf(`MATCH (from) WHERE ID(from) = %v
MATCH (to) WHERE ID(to) = %v
CREATE (from)-[%v]->(to)`, fromId, toId, label)
	_, err := c.conn.Exec(stmt)
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
		if math.IsNaN(float64(value)) {
			return float64(0)
		}
		return float64(value)

	case types.Vec4f:
		return []float64{typeConvPropertyValue(value[0]).(float64),
			typeConvPropertyValue(value[1]).(float64),
			typeConvPropertyValue(value[2]).(float64),
			typeConvPropertyValue(value[3]).(float64)}

	case types.Unknown32:
		return int64(value)

	default:
		if value == nil {
			return "nil"
		}
		return value
	}
}

func createPropertyList(props map[string]interface{}) (string, []interface{}) {
	results := make([]string, 0)

	values := make([]interface{}, 0)
	i := 0
	for name, v := range props {
		results = append(results, fmt.Sprintf("%v: {%v}", name, i))
		values = append(values, v)
		i++
	}

	return strings.Join(results, ", "), values
}
