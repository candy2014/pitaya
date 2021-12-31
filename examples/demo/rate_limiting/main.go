package main

import (
	"flag"
	"fmt"
	"github.com/topfreegames/pitaya/conn/message"
	"github.com/topfreegames/pitaya/serialize/papertiger"
	"time"

	"strings"

	"github.com/spf13/viper"
	"github.com/topfreegames/pitaya"
	"github.com/topfreegames/pitaya/acceptor"
	"github.com/topfreegames/pitaya/acceptorwrapper"
	"github.com/topfreegames/pitaya/component"
	"github.com/topfreegames/pitaya/config"
	"github.com/topfreegames/pitaya/examples/demo/rate_limiting/services"
)

func configureFrontend(port int) {
	room := services.NewRoom()
	pitaya.Register(room,
		component.WithName("room"),
		component.WithNameFunc(strings.ToLower))

	// 5 requests in 1 minute. Doesn't make sense, just to test
	// rate limiting
	vConfig := viper.New()
	vConfig.Set("pitaya.conn.ratelimiting.limit", 5)
	vConfig.Set("pitaya.conn.ratelimiting.interval", time.Minute)
	pConfig := config.NewConfig(vConfig)

	tcp := acceptor.NewTCPAcceptor(fmt.Sprintf(":%d", port))
	wrapped := acceptorwrapper.WithWrappers(
		tcp,
		acceptorwrapper.NewRateLimitingWrapper(pConfig))
	pitaya.AddAcceptor(wrapped)

}

func main() {
	defer pitaya.Shutdown()

	port := flag.Int("port", 3250, "the port to listen")
	svType := "room"

	flag.Parse()
	configureFrontend(*port)

	pitaya.Configure(true, svType, pitaya.Cluster, map[string]string{})
	pitaya.SetSerializer(papertiger.NewPaperTigerBinary())
	pitaya.SetPacketDecoder(NewPaperTigerPacketDecoder())
	pitaya.SetPacketEncoder(NewPaperTigerPacketEncoder())
	pitaya.SetMessageEncoder(message.NewForwardMessageEncoder(false))
	pitaya.Start()
}
