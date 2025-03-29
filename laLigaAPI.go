package main

import (
	"database/sql"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/cors"
)

type Match struct {
	MatchID      string `json:"match_id"`
	HomeTeam     string `json:"home_team"`
	AwayTeam     string `json:"away_team"`
	Date         string `json:"date"`
	HomeGoals    int    `json:"home_goals,omitempty"`
	AwayGoals    int    `json:"away_goals,omitempty"`
	YellowCards  int    `json:"yellow_cards,omitempty"`
	RedCards     int    `json:"red_cards,omitempty"`
	ExtraMinutes int    `json:"extra_minutes,omitempty"`
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
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           int((12 * time.Hour).Seconds()),
	})

	// Apply the CORS middleware
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

	// New patch endpoints
	router.PATCH("/matches/:id/goals", updateGoals)
	router.PATCH("/matches/:id/yellowcards", updateYellowCards)
	router.PATCH("/matches/:id/redcards", updateRedCards)
	router.PATCH("/matches/:id/extratime", updateExtraTime)

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
			date TEXT NOT NULL,
			home_goals INTEGER DEFAULT 0,
			away_goals INTEGER DEFAULT 0,
			yellow_cards INTEGER DEFAULT 0,
			red_cards INTEGER DEFAULT 0,
			extra_minutes INTEGER DEFAULT 0
		)
	`)
	if err != nil {
		panic(err)
	}

	// Verificar si la tabla está vacía e insertar datos iniciales
	insertInitialData()
}

func insertInitialData() {
	// Verificar si la tabla está vacía
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM matches").Scan(&count)
	if err != nil {
		panic(err)
	}

	if count == 0 {
		// Datos iniciales de partidos de LaLiga
		initialMatches := []Match{
			{
				MatchID:      "1",
				HomeTeam:     "Real Madrid",
				AwayTeam:     "Barcelona",
				Date:         "2023-10-28",
				HomeGoals:    2,
				AwayGoals:    1,
				YellowCards:  3,
				RedCards:     1,
				ExtraMinutes: 5,
			},
			{
				MatchID:      "2",
				HomeTeam:     "Atletico Madrid",
				AwayTeam:     "Sevilla",
				Date:         "2023-10-29",
				HomeGoals:    1,
				AwayGoals:    0,
				YellowCards:  2,
				RedCards:     0,
				ExtraMinutes: 3,
			},
			{
				MatchID:      "3",
				HomeTeam:     "Valencia",
				AwayTeam:     "Villarreal",
				Date:         "2023-10-30",
				HomeGoals:    0,
				AwayGoals:    0,
				YellowCards:  1,
				RedCards:     0,
				ExtraMinutes: 0,
			},
			{
				MatchID:      "4",
				HomeTeam:     "Athletic Bilbao",
				AwayTeam:     "Real Sociedad",
				Date:         "2023-10-31",
				HomeGoals:    1,
				AwayGoals:    1,
				YellowCards:  4,
				RedCards:     0,
				ExtraMinutes: 0,
			},
			{
				MatchID:      "5",
				HomeTeam:     "Betis",
				AwayTeam:     "Espanyol",
				Date:         "2023-11-01",
				HomeGoals:    3,
				AwayGoals:    2,
				YellowCards:  2,
				RedCards:     0,
				ExtraMinutes: 0,
			},
		}

		// Insertar los partidos iniciales
		tx, err := db.Begin()
		if err != nil {
			panic(err)
		}

		for _, match := range initialMatches {
			_, err = tx.Exec(
				"INSERT INTO matches (match_id, home_team, away_team, date, home_goals, away_goals, yellow_cards, red_cards, extra_minutes) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
				match.MatchID, match.HomeTeam, match.AwayTeam, match.Date, match.HomeGoals, match.AwayGoals, match.YellowCards, match.RedCards, match.ExtraMinutes,
			)
			if err != nil {
				tx.Rollback()
				panic(err)
			}
		}

		err = tx.Commit()
		if err != nil {
			panic(err)
		}
	}
}

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

func getMatches(c *gin.Context) {
	rows, err := db.Query("SELECT match_id, home_team, away_team, date, home_goals, away_goals, yellow_cards, red_cards, extra_minutes FROM matches")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve matches"})
		return
	}
	defer rows.Close()

	var matches []Match
	for rows.Next() {
		var match Match
		if err := rows.Scan(
			&match.MatchID,
			&match.HomeTeam,
			&match.AwayTeam,
			&match.Date,
			&match.HomeGoals,
			&match.AwayGoals,
			&match.YellowCards,
			&match.RedCards,
			&match.ExtraMinutes,
		); err != nil {
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
		"SELECT match_id, home_team, away_team, date, home_goals, away_goals, yellow_cards, red_cards, extra_minutes FROM matches WHERE match_id = ?",
		matchID,
	).Scan(
		&match.MatchID,
		&match.HomeTeam,
		&match.AwayTeam,
		&match.Date,
		&match.HomeGoals,
		&match.AwayGoals,
		&match.YellowCards,
		&match.RedCards,
		&match.ExtraMinutes,
	)

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
		"INSERT INTO matches (match_id, home_team, away_team, date, home_goals, away_goals, yellow_cards, red_cards, extra_minutes) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		newMatch.MatchID,
		newMatch.HomeTeam,
		newMatch.AwayTeam,
		newMatch.Date,
		newMatch.HomeGoals,
		newMatch.AwayGoals,
		newMatch.YellowCards,
		newMatch.RedCards,
		newMatch.ExtraMinutes,
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
		"UPDATE matches SET home_team = ?, away_team = ?, date = ?, home_goals = ?, away_goals = ?, yellow_cards = ?, red_cards = ?, extra_minutes = ? WHERE match_id = ?",
		updatedMatch.HomeTeam,
		updatedMatch.AwayTeam,
		updatedMatch.Date,
		updatedMatch.HomeGoals,
		updatedMatch.AwayGoals,
		updatedMatch.YellowCards,
		updatedMatch.RedCards,
		updatedMatch.ExtraMinutes,
		matchID,
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

func updateGoals(c *gin.Context) {
	matchID := c.Param("id")

	var goals struct {
		HomeGoals int `json:"home_goals"`
		AwayGoals int `json:"away_goals"`
	}

	if err := c.ShouldBindJSON(&goals); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	result, err := db.Exec(
		"UPDATE matches SET home_goals = ?, away_goals = ? WHERE match_id = ?",
		goals.HomeGoals, goals.AwayGoals, matchID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update goals"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Match not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Goals updated"})
}

func updateYellowCards(c *gin.Context) {
	matchID := c.Param("id")

	var yellowCards struct {
		YellowCards int `json:"yellow_cards"`
	}

	if err := c.ShouldBindJSON(&yellowCards); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	result, err := db.Exec(
		"UPDATE matches SET yellow_cards = ? WHERE match_id = ?",
		yellowCards.YellowCards, matchID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update yellow cards"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Match not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Yellow cards updated"})
}

func updateRedCards(c *gin.Context) {
	matchID := c.Param("id")

	var redCards struct {
		RedCards int `json:"red_cards"`
	}

	if err := c.ShouldBindJSON(&redCards); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	result, err := db.Exec(
		"UPDATE matches SET red_cards = ? WHERE match_id = ?",
		redCards.RedCards, matchID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update red cards"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Match not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Red cards updated"})
}

func updateExtraTime(c *gin.Context) {
	matchID := c.Param("id")

	var extraTime struct {
		ExtraMinutes int `json:"extra_minutes"`
	}

	if err := c.ShouldBindJSON(&extraTime); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	result, err := db.Exec(
		"UPDATE matches SET extra_minutes = ? WHERE match_id = ?",
		extraTime.ExtraMinutes, matchID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update extra time"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Match not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Extra time updated"})
}