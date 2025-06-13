package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/bulatorr/go-yaynison/ynison"
	"github.com/gin-gonic/gin"
)

const AUTH_HEADER = "Authorization"

func main() {
	r := gin.Default()
	r.GET("/get_current_track_alpha", tokenGIN)
	r.Run(":8080")
}

func tokenGIN(c *gin.Context) {
	token := c.Request.Header.Get(AUTH_HEADER)

	// чуть чуть проверки токена
	token, ok := strings.CutPrefix(token, "OAuth ")
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "Incorrect token entered",
		})
		return
	}

	// канал, куда будем отправлять данные по готовности
	done := make(chan ynison.PutYnisonStateResponse, 1)

	y := ynison.NewClient(token)
	defer y.Close()
	y.OnMessage(func(pysr ynison.PutYnisonStateResponse) {
		done <- pysr
	})

	err := y.Connect()
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	select {
	case data := <-done:
		if len(data.PlayerState.PlayerQueue.PlayableList) > 0 {
			index := data.PlayerState.PlayerQueue.CurrentPlayableIndex
			c.JSON(http.StatusOK, gin.H{
				"paused":      data.PlayerState.Status.Paused,
				"duration_ms": data.PlayerState.Status.DurationMs,
				"progress_ms": data.PlayerState.Status.ProgressMs,
				"entity_id":   data.PlayerState.PlayerQueue.EntityID,
				"entity_type": data.PlayerState.PlayerQueue.EntityType,
				"track_id":    data.PlayerState.PlayerQueue.PlayableList[index].PlayableID,
			})
			return
		} else {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "PlayerQueue information missing",
			})
			return
		}
	case <-time.After(10 * time.Second):
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "Failed to retrieve data"})
		return
	case <-c.Request.Context().Done():
		c.AbortWithStatus(http.StatusGatewayTimeout)
		return
	}
}
