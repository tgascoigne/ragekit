package encyclopedia

import (
	"fmt"
	"strings"
	"sync"

	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"
	"github.com/tgascoigne/ragekit/jenkins"
	"github.com/tgascoigne/ragekit/resource/types"
)

type dbConn struct {
	driver bolt.Driver
	conn   bolt.Conn
	lock   sync.Mutex
}

var DB dbConn

func ConnectDb(addr string) {
	DB = dbConn{
		driver: bolt.NewDriver(),
	}

	DB.lock.Lock()
	defer DB.lock.Unlock()
	driver := DB.driver

	var err error
	DB.conn, err = driver.OpenNeo("bolt://neo4j:jetpack@mimas:7687")
	if err != nil {
		panic(err)
	}
}

func Graph(nodes []Node) {
	DB.lock.Lock()
	defer DB.lock.Unlock()

	for _, node := range nodes {
		GraphNode(node)
	}
}

func GraphNode(node Node) {
	//CREATE (f:FOO {a: {a}, b: {b}, c: {c}, d: {d}, e: {e}, f: {f}, g: {g}, h: {h}})-[b:BAR]->(c:BAZ)
	stmt := fmt.Sprintf("MERGE (n:%v {%v}) RETURN ID(n)", node.Label(), createPropertyList(node.Properties()))

	//fmt.Printf("node is %#v\n", node)
	//fmt.Printf("statement is %v\n", stmt)
	conn := DB.conn
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
			GraphNode(asset)
			GraphRelationship(nodeId, asset, ":"+name)
		}
	}
}

func nodeId(node Node) int64 {
	stmt := fmt.Sprintf("MATCH (n:%v {%v}) RETURN ID(n)", node.Label(), createPropertyList(node.Properties()))
	params := typeConvPropertyValues(node.Properties())

	result, err := DB.conn.QueryNeo(stmt, params)
	if err != nil {
		panic(err)
	}

	row, _, _ := result.NextNeo()
	id := row[0].(int64)

	result.Close()

	return id
}

func GraphRelationship(fromId int64, to Node, label string) {
	toId := nodeId(to)

	stmt := fmt.Sprintf(`MATCH (from) WHERE ID(from) = %v
MATCH (to) WHERE ID(to) = %v
CREATE (from)-[%v]->(to)`, fromId, toId, label)

	_, err := DB.conn.ExecNeo(stmt, nil)
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

func CloseDb() {
	DB.conn.Close()
}
