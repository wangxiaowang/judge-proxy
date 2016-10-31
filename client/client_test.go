package client_test

import (
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/zhexuany/judge-proxy/client"
	"github.com/zhexuany/wordGenerator"
)

func getRandomConfig(length int) client.Config {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)
	addrs := make([]string, length)
	var localhost string
	localhost = "http://192.168.80.158"
	for i := 0; i < length; i++ {
		url := localhost + ":" + string((r.Intn(10000) + 100))
		addrs[i] = url
	}

	config := client.Config{
		Addrs: addrs,
	}
	return config
}

func TestClient_GetNode(t *testing.T) {
	//create four virtual address
	config := getRandomConfig((20))

	c, err := client.NewClient(config)

	if err != nil {
		t.Errorf("Failed to create client %v", err)
	}

	prefix := "cpu"

	times := make(map[string]float64)

	for i := 0; i < 1000; i++ {
		measurement := prefix + wordGenerator.GetWord(20)
		url, _ := c.GetNode(measurement)
		times[url] += 1
	}

	avergae := 0.0
	length := float64(len(times))

	for _, v := range times {
		avergae += v
	}

	avergae /= length

	for _, v := range times {
		diff := math.Abs(float64(v) - avergae)
		if diff > 1000*0.3 {
			t.Errorf(" Evenly distributing keys failed. Differnce: %f. Average: %f", diff, avergae)
			// break
		}
	}
}

func TestClient_ResetConfig(t *testing.T) {
	oldConfig := getRandomConfig(4)
	c, _ := client.NewClient(oldConfig)

	newConfig := getRandomConfig(4)
	// go func() {
	// for i := 0; i < 10; i++ {
	// node, _ := c.GetNode(wordGenerator.GetWord(4))
	// }
	// }()

	err, ok := c.ResetConfig(newConfig)

	if !ok {
		t.Errorf("Failed to reset client's config", err)
	}
}

func difference(a []string, b []string) []string {
	diffStrs := []string{}
	m := map[string]int{}

	for _, str := range a {
		m[str] = 1
	}

	for _, str := range b {
		m[str] = m[str] + 1
	}

	for k, v := range m {
		if v == 1 {
			diffStrs = append(diffStrs, k)
		}
	}

	return diffStrs
}

func TestClient_RemoveNode(t *testing.T) {
	config := getRandomConfig(10)
	c, _ := client.NewClient(config)

	measurements := wordGenerator.GetWords(1000, 20)
	measurementsMap := make(map[string][]string)

	for _, measurement := range measurements {
		nodes := measurementsMap[measurement]
		if node, ok := c.GetNode(measurement); ok {
			nodes = append(nodes, node)
		}
	}
	// s := rand.NewSource(time.Now().UnixNano())
	// r := rand.New(s)
	c.RemoveNode(config.Addrs[0])

	newMeasurementsMap := make(map[string][]string)
	for _, measurement := range measurements {
		nodes := newMeasurementsMap[measurement]
		if node, ok := c.GetNode(measurement); ok {
			nodes = append(nodes, node)
		}
	}

	//TODO: compare the difference between measurements
	for _, measurement := range measurements {
		if len(difference(measurementsMap[measurement], newMeasurementsMap[measurement])) > 10 {
			t.Errorf("Failed to verify remove node")
		}
	}

}
