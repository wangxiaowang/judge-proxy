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
	localhost = "0.0.0.0"
	for i := 0; i < 4; i++ {
		url := localhost + ":" + string(r.Intn(10)*1000)
		addrs[i] = url
	}

	var config client.Config
	copy(config.Addrs, addrs)

	return config
}

func TestClient_GetNode(t *testing.T) {
	//create four virtual address
	c, _ := client.NewClient(getRandomConfig((4)))

	prefix := "cpu"

	wordGen := wordGenerator.New()

	times := make(map[string]float64)

	for i := 0; i < 1000; i++ {
		url, _ := c.GetNode(prefix + wordGen.GetWord(20))
		times[url] = times[url] + 1
	}

	avergae := 0.0
	length := float64(len(times))
	for _, v := range times {
		avergae += v
	}

	avergae /= length

	ok := true

	for _, v := range times {
		if math.Abs(float64(v)-avergae) > 30 {
			ok = false
			break
		}
	}

	if !ok {
		t.Errorf("Consistent hashing can not distributed incoming request evenly")
	}
}

func TestClient_ResetConfig(t *testing.T) {
	oldConfig := getRandomConfig(4)
	c, _ := client.NewClient(oldConfig)

	newConfig := getRandomConfig(4)
	//TODO: zhexuany, need to simulate the concurrent behavior
	//GetNode and ResetConfig are happending at the same time.
	err, ok := c.ResetConfig(newConfig)

	if !ok {
		t.Errorf("Failed to reset client's config", err)
	}
}
