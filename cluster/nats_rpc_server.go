// Copyright (c) TFG Co. All Rights Reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package cluster

import (
	"context"
	"errors"
	"fmt"
	context2 "github.com/topfreegames/pitaya/context"
	"hash/fnv"
	"math/rand"

	//"github.com/topfreegames/pitaya/service"
	"math"
	"time"

	"github.com/golang/protobuf/proto"
	nats "github.com/nats-io/nats.go"
	"github.com/topfreegames/pitaya/config"
	"github.com/topfreegames/pitaya/constants"
	e "github.com/topfreegames/pitaya/errors"
	"github.com/topfreegames/pitaya/logger"
	"github.com/topfreegames/pitaya/metrics"
	"github.com/topfreegames/pitaya/protos"
	"github.com/topfreegames/pitaya/session"
	"github.com/topfreegames/pitaya/util"
)

// NatsRPCServer struct
type NatsRPCServer struct {
	connString             string
	connectionTimeout      time.Duration
	maxReconnectionRetries int
	server                 *Server
	conn                   *nats.Conn
	pushBufferSize         int
	messagesBufferSize     int
	config                 *config.Config
	stopChan               chan bool
	subChan                chan *nats.Msg // subChan is the channel used by the server to receive network messages addressed to itself
	bindingsChan           chan *nats.Msg // bindingsChan receives notify from other servers on every user bind to session
	unhandledReqCh         []chan *protos.Request
	responses              []*protos.Response
	requests               []*protos.Request
	userPushCh             chan *protos.Push
	userKickCh             chan *protos.KickMsg
	broadcastPushCh        chan *nats.Msg
	sub                    *nats.Subscription
	dropped                int
	pitayaServer           protos.PitayaServer
	metricsReporters       []metrics.Reporter
	appDieChan             chan bool
}

// NewNatsRPCServer ctor
func NewNatsRPCServer(
	config *config.Config,
	server *Server,
	metricsReporters []metrics.Reporter,
	appDieChan chan bool,
) (*NatsRPCServer, error) {
	dispatchThreadNum := config.GetInt("pitaya.concurrency.remote.service")
	chLocals := make([]chan *protos.Request, dispatchThreadNum)
	for i := 0; i < dispatchThreadNum; i++ {
		chLocals[i] = make(chan *protos.Request, dispatchThreadNum)
	}

	ns := &NatsRPCServer{
		config:            config,
		server:            server,
		stopChan:          make(chan bool),
		unhandledReqCh:    chLocals,
		dropped:           0,
		metricsReporters:  metricsReporters,
		appDieChan:        appDieChan,
		connectionTimeout: nats.DefaultTimeout,
	}
	if err := ns.configure(); err != nil {
		return nil, err
	}

	return ns, nil
}

func (ns *NatsRPCServer) configure() error {
	ns.connString = ns.config.GetString("pitaya.cluster.rpc.server.nats.connect")
	if ns.connString == "" {
		return constants.ErrNoNatsConnectionString
	}
	ns.connectionTimeout = ns.config.GetDuration("pitaya.cluster.rpc.server.nats.connectiontimeout")
	ns.maxReconnectionRetries = ns.config.GetInt("pitaya.cluster.rpc.server.nats.maxreconnectionretries")
	ns.messagesBufferSize = ns.config.GetInt("pitaya.buffer.cluster.rpc.server.nats.messages")
	if ns.messagesBufferSize == 0 {
		return constants.ErrNatsMessagesBufferSizeZero
	}
	ns.pushBufferSize = ns.config.GetInt("pitaya.buffer.cluster.rpc.server.nats.push")
	if ns.pushBufferSize == 0 {
		return constants.ErrNatsPushBufferSizeZero
	}
	ns.subChan = make(chan *nats.Msg, ns.messagesBufferSize)
	ns.bindingsChan = make(chan *nats.Msg, ns.messagesBufferSize)
	// the reason this channel is buffered is that we can achieve more performance by not
	// blocking producers on a massive push
	ns.userPushCh = make(chan *protos.Push, ns.pushBufferSize)
	ns.userKickCh = make(chan *protos.KickMsg, ns.messagesBufferSize)
	ns.broadcastPushCh = make(chan *nats.Msg, ns.pushBufferSize)
	return nil
}

// GetBindingsChannel gets the channel that will receive all bindings
func (ns *NatsRPCServer) GetBindingsChannel() chan *nats.Msg {
	return ns.bindingsChan
}

// GetUserMessagesTopic get the topic for user
func GetUserMessagesTopic(uid string, svType string) string {
	return fmt.Sprintf("pitaya/%s/user/%s/push", svType, uid)
}

// GetUserKickTopic get the topic for kicking an user
func GetUserKickTopic(uid string, svType string) string {
	return fmt.Sprintf("pitaya/%s/user/%s/kick", svType, uid)
}

// GetBindBroadcastTopic gets the topic on which bind events will be broadcasted
func GetBindBroadcastTopic(svType string) string {
	return fmt.Sprintf("pitaya/%s/bindings", svType)
}

// GetPushBroadcastTopic gets the topic on which broadcast events will be push
func GetPushBroadcastTopic(svType string) string {
	return fmt.Sprintf("pitaya/%s/broadcasted", svType)
}

// onSessionBind should be called on each session bind
func (ns *NatsRPCServer) onSessionBind(ctx context.Context, s *session.Session) error {
	if ns.server.Frontend {
		subu, err := ns.subscribeToUserMessages(s.UID(), ns.server.Type)
		if err != nil {
			return err
		}
		subk, err := ns.subscribeToUserKickChannel(s.UID(), ns.server.Type)
		if err != nil {
			return err
		}
		s.Subscriptions = []*nats.Subscription{subu, subk}
	}
	return nil
}

// SetPitayaServer sets the pitaya server
func (ns *NatsRPCServer) SetPitayaServer(ps protos.PitayaServer) {
	ns.pitayaServer = ps
}

func (ns *NatsRPCServer) subscribeToBindingsChannel() error {
	_, err := ns.conn.ChanSubscribe(GetBindBroadcastTopic(ns.server.Type), ns.bindingsChan)
	return err
}

func (ns *NatsRPCServer) subscribeToPushBroadcastChannel() error {
	_, err := ns.conn.ChanSubscribe(GetPushBroadcastTopic(ns.server.Type), ns.broadcastPushCh)
	return err
}

func (ns *NatsRPCServer) subscribeToUserKickChannel(uid string, svType string) (*nats.Subscription, error) {
	sub, err := ns.conn.Subscribe(GetUserKickTopic(uid, svType), func(msg *nats.Msg) {
		kick := &protos.KickMsg{}
		err := proto.Unmarshal(msg.Data, kick)
		if err != nil {
			logger.Log.Error("error unrmarshalling push: ", err.Error())
		}
		ns.userKickCh <- kick
	})
	return sub, err
}

func (ns *NatsRPCServer) subscribeToUserMessages(uid string, svType string) (*nats.Subscription, error) {
	sub, err := ns.conn.Subscribe(GetUserMessagesTopic(uid, svType), func(msg *nats.Msg) {
		push := &protos.Push{}
		err := proto.Unmarshal(msg.Data, push)
		if err != nil {
			logger.Log.Error("error unmarshalling push:", err.Error())
		}
		ns.userPushCh <- push
	})
	if err != nil {
		return nil, err
	}
	return sub, nil
}

func (ns *NatsRPCServer) handleMessages() {
	defer (func() {
		ns.conn.Drain()
		for i := range ns.unhandledReqCh {
			close(ns.unhandledReqCh[i])
		}
		close(ns.subChan)
		close(ns.bindingsChan)
	})()
	maxPending := float64(0)
	for {
		select {
		case msg := <-ns.subChan:
			ns.reportMetrics()
			dropped, err := ns.sub.Dropped()
			if err != nil {
				logger.Log.Errorf("error getting number of dropped messages: %s", err.Error())
			}
			if dropped > ns.dropped {
				logger.Log.Warnf("[rpc server] some messages were dropped! numDropped: %d", dropped)
				ns.dropped = dropped
			}
			subsChanLen := float64(len(ns.subChan))
			maxPending = math.Max(float64(maxPending), subsChanLen)
			logger.Log.Debugf("subs channel size: %d, max: %d, dropped: %d", subsChanLen, maxPending, dropped)
			req := &protos.Request{}
			// TODO: Add tracing here to report delay to start processing message in spans
			err = proto.Unmarshal(msg.Data, req)
			if err != nil {
				// should answer rpc with an error
				logger.Log.Error("error unmarshalling rpc message:", err.Error())
				continue
			}

			var threadId int64 = -1
			var sessionUid string
			if req.Session != nil {
				sessionUid = req.Session.Uid
			} else {
				ctx, _ := util.GetContextFromRequest(req, "")
				if id := context2.GetFromPropagateCtx(ctx, constants.SessionIdCtxKey); id != nil {
					if uid, ok := id.(string); ok {
						sessionUid = uid
					}
				}
			}
			dispatchThreadNum := ns.config.GetInt("pitaya.concurrency.remote.service")

			random := func() {
				threadId = rand.Int63n(int64(dispatchThreadNum))
			}

			if sessionUid != "" {
				sessionId := int64(hash(sessionUid))
				threadId = sessionId % int64(dispatchThreadNum)
			}

			if threadId < 0 || threadId >= int64(dispatchThreadNum) {
				random()
			}
			logger.Log.Errorf("当前的sessionUid %s threadId %d route %s", sessionUid, threadId, req.Msg.Route)
			req.Msg.Reply = msg.Reply
			ns.unhandledReqCh[int(threadId)] <- req
		case <-ns.stopChan:
			return
		}
	}
}

func hash(s string) uint64 {
	h := fnv.New64()
	h.Write([]byte(s))
	return h.Sum64()
}

// GetUnhandledRequestsChannel gets the unhandled requests channel from nats rpc server
func (ns *NatsRPCServer) GetUnhandledRequestsChannel(threadId int) chan *protos.Request {
	return ns.unhandledReqCh[threadId]
}

func (ns *NatsRPCServer) getUserPushChannel() chan *protos.Push {
	return ns.userPushCh
}

func (ns *NatsRPCServer) getBroadcastPushChChannel() chan *nats.Msg {
	return ns.broadcastPushCh
}

func (ns *NatsRPCServer) getUserKickChannel() chan *protos.KickMsg {
	return ns.userKickCh
}

func (ns *NatsRPCServer) marshalResponse(res *protos.Response) ([]byte, error) {
	p, err := proto.Marshal(res)
	if err != nil {
		res := &protos.Response{
			Error: &protos.Error{
				Code: e.ErrUnknownCode,
				Msg:  err.Error(),
			},
		}
		p, _ = proto.Marshal(res)
	}

	if err == nil && res.Error != nil {
		err = errors.New(res.Error.Msg)
	}
	return p, err
}

func (ns *NatsRPCServer) processMessages(threadID int) {
	for ns.requests[threadID] = range ns.GetUnhandledRequestsChannel(threadID) {
		logger.Log.Debugf("(%d) processing message %v", threadID, ns.requests[threadID].GetMsg().GetId())
		ctx, err := util.GetContextFromRequest(ns.requests[threadID], ns.server.ID)
		if err != nil {
			ns.responses[threadID] = &protos.Response{
				Error: &protos.Error{
					Code: e.ErrInternalCode,
					Msg:  err.Error(),
				},
			}
		} else {
			ns.responses[threadID], _ = ns.pitayaServer.Call(ctx, ns.requests[threadID])
		}
		p, err := ns.marshalResponse(ns.responses[threadID])
		err = ns.conn.Publish(ns.requests[threadID].GetMsg().GetReply(), p)
		if err != nil {
			logger.Log.Error("error sending message response")
		}
	}
}

func (ns *NatsRPCServer) processSessionBindings() {
	for bind := range ns.bindingsChan {
		b := &protos.BindMsg{}
		err := proto.Unmarshal(bind.Data, b)
		if err != nil {
			logger.Log.Errorf("error processing binding msg: %v", err)
			continue
		}
		ns.pitayaServer.SessionBindRemote(context.Background(), b)
	}
}

func (ns *NatsRPCServer) processPushes() {
	for push := range ns.getUserPushChannel() {
		logger.Log.Debugf("sending push to user %s: %v", push.GetUid(), string(push.Data))
		_, err := ns.pitayaServer.PushToUser(context.Background(), push)
		if err != nil {
			logger.Log.Errorf("error sending push to user: %v", err)
		}
	}
}

func (ns *NatsRPCServer) processBroadcastPush() {
	for push := range ns.getBroadcastPushChChannel() {
		b := &protos.Push{}
		err := proto.Unmarshal(push.Data, b)
		logger.Log.Debugf("broadcasts push to user %s: %v", b.GetUid(), string(push.Data))
		_, err = ns.pitayaServer.BroadcastToUser(context.Background(), b)
		if err != nil {
			logger.Log.Errorf("error sending push to user: %v", err)
		}
	}
}

func (ns *NatsRPCServer) processKick() {
	for kick := range ns.getUserKickChannel() {
		logger.Log.Debugf("Sending kick to user %s: %v", kick.GetUserId())
		_, err := ns.pitayaServer.KickUser(context.Background(), kick)
		if err != nil {
			logger.Log.Errorf("error sending kick to user: %v", err)
		}
	}
}

// Init inits nats rpc server
func (ns *NatsRPCServer) Init() error {
	ns.responses = make([]*protos.Response, ns.config.GetInt("pitaya.concurrency.remote.service"))
	ns.requests = make([]*protos.Request, ns.config.GetInt("pitaya.concurrency.remote.service"))
	// TODO should we have concurrency here? it feels like we should
	go ns.handleMessages()

	logger.Log.Debugf("connecting to nats (server) with timeout of %s", ns.connectionTimeout)
	conn, err := setupNatsConn(
		ns.connString,
		ns.appDieChan,
		nats.MaxReconnects(ns.maxReconnectionRetries),
		nats.Timeout(ns.connectionTimeout),
	)
	if err != nil {
		return err
	}
	ns.conn = conn
	if ns.sub, err = ns.subscribe(getChannel(ns.server.Type, ns.server.ID)); err != nil {
		return err
	}

	err = ns.subscribeToBindingsChannel()
	if err != nil {
		return err
	}

	err = ns.subscribeToPushBroadcastChannel()
	if err != nil {
		return err
	}
	// this handles remote messages
	for i := 0; i < ns.config.GetInt("pitaya.concurrency.remote.service"); i++ {
		go ns.processMessages(i)
	}

	session.OnSessionBind(ns.onSessionBind)

	// this should be so fast that we shoudn't need concurrency
	go ns.processPushes()
	go ns.processBroadcastPush()
	go ns.processSessionBindings()
	go ns.processKick()

	return nil
}

// AfterInit runs after initialization
func (ns *NatsRPCServer) AfterInit() {}

// BeforeShutdown runs before shutdown
func (ns *NatsRPCServer) BeforeShutdown() {}

// Shutdown stops nats rpc server
func (ns *NatsRPCServer) Shutdown() error {
	close(ns.stopChan)
	return nil
}

func (ns *NatsRPCServer) subscribe(topic string) (*nats.Subscription, error) {
	return ns.conn.ChanSubscribe(topic, ns.subChan)
}

func (ns *NatsRPCServer) stop() {
}

func (ns *NatsRPCServer) reportMetrics() {
	if ns.metricsReporters != nil {
		for _, mr := range ns.metricsReporters {
			if err := mr.ReportGauge(metrics.DroppedMessages, map[string]string{}, float64(ns.dropped)); err != nil {
				logger.Log.Warnf("failed to report dropped message: %s", err.Error())
			}

			// subchan
			subChanCapacity := ns.messagesBufferSize - len(ns.subChan)
			if subChanCapacity == 0 {
				logger.Log.Warn("subChan is at maximum capacity")
			}
			if err := mr.ReportGauge(metrics.ChannelCapacity, map[string]string{"channel": "rpc_server_subchan"}, float64(subChanCapacity)); err != nil {
				logger.Log.Warnf("failed to report subChan queue capacity: %s", err.Error())
			}

			// bindingschan
			bindingsChanCapacity := ns.messagesBufferSize - len(ns.bindingsChan)
			if bindingsChanCapacity == 0 {
				logger.Log.Warn("bindingsChan is at maximum capacity")
			}
			if err := mr.ReportGauge(metrics.ChannelCapacity, map[string]string{"channel": "rpc_server_bindingschan"}, float64(bindingsChanCapacity)); err != nil {
				logger.Log.Warnf("failed to report bindingsChan capacity: %s", err.Error())
			}

			// userpushch
			userPushChanCapacity := ns.messagesBufferSize - len(ns.bindingsChan)
			if userPushChanCapacity == 0 {
				logger.Log.Warn("userPushChan is at maximum capacity")
			}
			if err := mr.ReportGauge(metrics.ChannelCapacity, map[string]string{"channel": "rpc_server_userpushchan"}, float64(userPushChanCapacity)); err != nil {
				logger.Log.Warnf("failed to report userPushCh capacity: %s", err.Error())
			}
		}
	}
}
