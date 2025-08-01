package wsserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"twitter-clone/internal/domain/cache"
	"twitter-clone/internal/domain/config"
	"twitter-clone/internal/domain/twitter"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type WebSocketServer struct {
	clients map[int64]*websocket.Conn
	cache   cache.Cache
	ctx     context.Context

	server *http.Server

	// TODO move it to config
	apiPath string
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024, // classic params needs to be checked
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewWebSocketServer(cache cache.Cache, config config.WSServerConfig) *WebSocketServer {
	commonAddress := fmt.Sprintf("%s:%d", config.WSServerHost(), config.WSServerPort())
	router := mux.NewRouter()
	webSocketServer := &WebSocketServer{
		clients: make(map[int64]*websocket.Conn),
		cache:   cache,
		server: &http.Server{
			Addr:    commonAddress, // Configurable port
			Handler: router,
		},
		apiPath: config.WSServerAPIPath(),
	}
	webSocketServer.registerRoutes()
	return webSocketServer
}

func (ws *WebSocketServer) registerRoutes() {
	router := ws.server.Handler.(*mux.Router)
	router.HandleFunc("/ws", ws.handleConnections)
}

func (ws *WebSocketServer) Start() error {
	return ws.server.ListenAndServe()
}

func (ws *WebSocketServer) Stop(ctx context.Context) error {
	return ws.server.Shutdown(ctx)
}

func (ws *WebSocketServer) Info() string {
	return ""
}

// Handle WebSocket connections
func (ws *WebSocketServer) handleConnections(w http.ResponseWriter, r *http.Request) {
	var (
		userStr string
		userID  int64
		err     error
	)
	ctx := r.Context()
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		return
	}
	defer conn.Close()

	if userStr = r.URL.Query().Get("user_id"); userStr == "" {
		return
	}
	if userID, err = strconv.ParseInt(userStr, 10, 64); err != nil {
		return
	}

	// firstly we have correctly set user, then -- add to the table
	ws.handleNewUser(ctx, userID)
	// I forgot how to store it not in a memory map
	ws.clients[userID] = conn
}

func (ws *WebSocketServer) handleNewUser(ctx context.Context, userID int64) {
	var err error
	var exists bool
	var timeline []twitter.Tweet
	if err := ws.checkSetFollowers(ctx, userID); err != nil {
		log.Printf("Error fetching followers: %v", err)
		return
	}
	// there could be no followers for use we have to distinguish it using errors

	if exists, err = ws.cache.CheckUserTimelineExists(ctx, userID); err != nil {
		log.Printf("Error fetching timeline: %v", err)
		return
	}
	if !exists {
		// API request to fetch tweets and store them in cache
		if timeline, err = ws.getTimeline(ctx, userID); err != nil {
			log.Printf("Error fetching timeline: %v", err)
			return
		}
		if err = ws.storeTimeline(ctx, userID, timeline); err != nil {
			log.Printf("Error storing timeline: %v", err)
			return
		}
	}

	// write to user it's timeline back
	// err = client.WriteMessage(websocket.TextMessage, timeline)
}

func (ws *WebSocketServer) getTimeline(ctx context.Context, userID int64) ([]twitter.Tweet, error) {
	panic("not implemented")
}

func (ws *WebSocketServer) storeTimeline(ctx context.Context, userID int64, timeline []twitter.Tweet) error {
	panic("not implemented")
}

func (ws *WebSocketServer) checkSetFollowers(ctx context.Context, userID int64) error {
	var (
		err       error
		req       *http.Request
		resp      *http.Response
		user      twitter.User
		followers []twitter.User
	)
	// ok, right now I don't have enough time, but it should be implemented a
	// API connector using interfaces and under the hood it should request everything
	// rn it will be pure requests, I hope i'll hv time soon
	userPath := fmt.Sprintf("%s/api/v1/get_user?user=%d", ws.apiPath, userID)
	if req, err = http.NewRequestWithContext(ctx, "GET", userPath, nil); err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}
	if resp, err = http.DefaultClient.Do(req); err != nil {
		log.Printf("Error making request: %v", err)
		return err
	}
	defer func() {
		_ = resp.Body.Close() // linter issue
	}()
	_ = json.NewDecoder(resp.Body).Decode(&user)

	if user.ID == 0 {
		return fmt.Errorf("invalid user ID")
	}

	followersPath := fmt.Sprintf("%s/api/v1/followers?user=%d", ws.apiPath, user.ID)
	if req, err = http.NewRequestWithContext(ctx, "GET", followersPath, nil); err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}
	if resp, err = http.DefaultClient.Do(req); err != nil {
		log.Printf("Error making request: %v", err)
		return err
	}
	defer func() {
		_ = resp.Body.Close() // linter issue
	}()
	_ = json.NewDecoder(resp.Body).Decode(&followers)

	if len(followers) == 0 {
		return fmt.Errorf("no followers found for user ID %d", userID)
	}

	if err = ws.cache.SetFollowers(ctx, userID, followers); err != nil {
		return err
	}

	return nil
}

func (ws *WebSocketServer) HandleTweets(ctx context.Context) {
	pubsub, err := ws.cache.SubscribeToTweetsChannel(ws.ctx, "tweets")
	if err != nil {
		log.Printf("Error subscribing to tweets channel: %v", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-pubsub:
			var tweet twitter.ChannelTweet
			if err := json.Unmarshal([]byte(msg), &tweet); err != nil {
				log.Printf("Unmarshal error: %v", err)
				continue
			}
			if conn, ok := ws.clients[tweet.UserID]; ok {
				var tweetMarshalled []byte
				tweetMarshalled, _ = json.Marshal(tweet.Tweet)
				err = conn.WriteMessage(websocket.TextMessage, tweetMarshalled)
				if err != nil {
					log.Printf("Error writing message to client: %v", err)
					delete(ws.clients, tweet.UserID)
				}
			} else {
				log.Printf("User %d is disconnected", tweet.UserID)
			}
		}
	}
}
