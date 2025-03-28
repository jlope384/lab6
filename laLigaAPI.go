package main

import (
	"database/sql"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/cors"
	_ "github.com/mattn/go-sqlite3"
)

type Match struct {
	MatchID   string `json:"match_id"`
	HomeTeam  string `json:"home_team"`
	AwayTeam  string `json:"away_team"`
	Date      string `json:"date"`
}

const (
	databasePath = "./data/laliga.db"
)

var db *sql.DB

func main() {
	setupDatabase()

	router := gin.Default()

	// Configuración CORS robusta con rs/cors
	corsConfig := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // En producción, reemplaza con tus dominios específicos
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposedHeaders:   []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})

	// Aplicar el middleware CORS
	router.Use(func(c *gin.Context) {
		corsConfig.HandlerFunc(c.Writer, c.Request)
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// Endpoints
	router.GET("/ping", ping)
	router.GET("/matches", getMatches)
	router.GET("/matches/:id", getMatch)
	router.PUT("/matches/:id", updateMatch)
	router.POST("/matches", createMatch)
	router.DELETE("/matches/:id", deleteMatch)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

func setupDatabase() {
	if err := os.MkdirAll("./data", 0755); err != nil {
		panic(err)
	}

	var err error
	db, err = sql.Open("sqlite3", databasePath)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS matches (
			match_id TEXT PRIMARY KEY,
			home_team TEXT NOT NULL,
			away_team TEXT NOT NULL,
			date TEXT NOT NULL
		)
	`)
	if err != nil {
		panic(err)
	}
}

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

func getMatches(c *gin.Context) {
	rows, err := db.Query("SELECT match_id, home_team, away_team, date FROM matches")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve matches"})
		return
	}
	defer rows.Close()

	var matches []Match
	for rows.Next() {
		var match Match
		if err := rows.Scan(&match.MatchID, &match.HomeTeam, &match.AwayTeam, &match.Date); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not read matches"})
			return
		}
		matches = append(matches, match)
	}

	c.JSON(http.StatusOK, matches)
}

func getMatch(c *gin.Context) {
	matchID := c.Param("id")

	var match Match
	err := db.QueryRow(
		"SELECT match_id, home_team, away_team, date FROM matches WHERE match_id = ?", 
		matchID,
	).Scan(&match.MatchID, &match.HomeTeam, &match.AwayTeam, &match.Date)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"message": "Match not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve match"})
		return
	}

	c.JSON(http.StatusOK, match)
}

func createMatch(c *gin.Context) {
	var newMatch Match
	if err := c.ShouldBindJSON(&newMatch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	_, err := db.Exec(
		"INSERT INTO matches (match_id, home_team, away_team, date) VALUES (?, ?, ?, ?)",
		newMatch.MatchID, newMatch.HomeTeam, newMatch.AwayTeam, newMatch.Date,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create match"})
		return
	}

	c.JSON(http.StatusCreated, newMatch)
}

func updateMatch(c *gin.Context) {
	matchID := c.Param("id")

	var updatedMatch Match
	if err := c.ShouldBindJSON(&updatedMatch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	result, err := db.Exec(
		"UPDATE matches SET home_team = ?, away_team = ?, date = ? WHERE match_id = ?",
		updatedMatch.HomeTeam, updatedMatch.AwayTeam, updatedMatch.Date, matchID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update match"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Match not found"})
		return
	}

	c.JSON(http.StatusOK, updatedMatch)
}

func deleteMatch(c *gin.Context) {
	matchID := c.Param("id")

	result, err := db.Exec("DELETE FROM matches WHERE match_id = ?", matchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete match"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Match not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Match deleted"})
}