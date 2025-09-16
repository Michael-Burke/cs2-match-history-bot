package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/olekukonko/tablewriter"
)

// This struct is used to read the player names from the json file
type FACEITPlayerNames struct {
	PlayerName []string `json:"players"`
}

type FACEITPlayerID struct {
	PlayerID string `json:"player_id"`
}

// This struct is a combination of the player names and player IDs. Player name is from FACEITPlayerNames struct, and the ID is read in from a function that parses the FACEIT API response
type FACEITPlayers struct {
	PlayerName string
	PlayerID   string
}

// function to query the FACEIT API and return the response
func QueryFACEITAPI(endpoint string, params map[string]interface{}) *http.Response {
	baseURL := "https://open.faceit.com/data/v4"
	u, err := url.Parse(baseURL + endpoint)
	if err != nil {
		log.Fatal("Error parsing URL: ", err)
	}

	q := u.Query()
	for key, val := range params {
		switch v := val.(type) {
		case string:
			q.Set(key, v)
		case int:
			q.Set(key, strconv.Itoa(v))
		case int64:
			q.Set(key, strconv.FormatInt(v, 10))
		case float64:
			q.Set(key, strconv.FormatFloat(v, 'f', -1, 64))
		case bool:
			q.Set(key, strconv.FormatBool(v))
		default:
			log.Fatalf("Unsupported type: %T", v)
		}
	}
	u.RawQuery = q.Encode()
	// log.Printf("Querying FACEIT API: %s", u.String())
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Fatal("Error creating request: ", err)
	}

	req.Header.Add("Authorization", "Bearer "+faceitAPIKey)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", "faceit-integration/1.0 (+https://open.faceit.com)")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error querying FACEIT API: ", err)
	}
	return resp
}

// Load the faceit player nicknames from data/faceit_player_ids.json
func loadPlayerJSON() FACEITPlayerNames {
	// log.Println("Loading player names from data/faceit_player_names.json")
	players, err := os.ReadFile("data/faceit_player_names.json")
	if err != nil {
		log.Fatal("Error reading faceit_player_names.json: ", err)
	}

	var faceitPlayerNames FACEITPlayerNames = FACEITPlayerNames{PlayerName: []string{}}
	err = json.Unmarshal(players, &faceitPlayerNames)
	if err != nil {
		log.Fatal("Error unmarshalling player names: ", err)
	}
	return faceitPlayerNames
}

func parseFACEITResponsePlayerNamesToIDs(response *http.Response) FACEITPlayerID {
	// Read in the body of a response that is returning a players json object
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal("Error reading response body: ", err)
	}
	response.Body.Close()
	// log.Printf("Response body: %s", body)
	var faceitPlayerIDs FACEITPlayerID
	err = json.Unmarshal(body, &faceitPlayerIDs)
	if err != nil {
		log.Fatal("Error unmarshalling player IDs: ", err)
	}
	return faceitPlayerIDs
}

func getPlayerIDs() []FACEITPlayers {
	players := loadPlayerJSON()
	var faceitPlayers []FACEITPlayers
	for _, player := range players.PlayerName {
		response := QueryFACEITAPI("/players", map[string]interface{}{"nickname": player})
		if response.StatusCode != 200 {
			log.Printf("Error getting player name for: %s ... Continuing", player)
			continue
		}
		parsedValues := parseFACEITResponsePlayerNamesToIDs(response)
		if parsedValues.PlayerID == "" {
			log.Printf("No player ID found for: %s ... Continuing", player)
			continue
		}
		faceitPlayers = append(faceitPlayers, FACEITPlayers{PlayerName: player, PlayerID: parsedValues.PlayerID})
	}
	return faceitPlayers
}

type Stats struct {
	Game                string  `json:"Game"`
	Team                string  `json:"Team"`
	Assists             int     `json:"Assists,string"`
	Rounds              int     `json:"Rounds,string"`
	OvertimeScore       int     `json:"Overtime score,string"`
	GameMode            string  `json:"Game Mode"`
	FinalScore          int     `json:"Final Score,string"`
	Map                 string  `json:"Map"`
	MatchID             string  `json:"Match Id"`
	Headshots           int     `json:"Headshots,string"`
	Nickname            string  `json:"Nickname"`
	CompetitionID       string  `json:"Competition Id"`
	Result              int     `json:"Result,string"`
	CreatedAt           string  `json:"Created At"`
	Score               string  `json:"Score"`
	Deaths              int     `json:"Deaths,string"`
	TripleKills         int     `json:"Triple Kills,string"`
	KDRatio             float64 `json:"K/D Ratio,string"`
	UpdatedAt           string  `json:"Updated At"`
	PentaKills          int     `json:"Penta Kills,string"`
	FirstHalfScore      int     `json:"First Half Score,string"`
	PlayerID            string  `json:"Player Id"`
	SecondHalfScore     int     `json:"Second Half Score,string"`
	Winner              string  `json:"Winner"`
	ADR                 float64 `json:"ADR,string"`
	KRRatio             float64 `json:"K/R Ratio,string"`
	HeadshotsPercentage float64 `json:"Headshots %,string"`
	MatchFinishedAt     int64   `json:"Match Finished At"`
	DoubleKills         int     `json:"Double Kills,string"`
	Region              string  `json:"Region"`
	Kills               int     `json:"Kills,string"`
	MVPs                int     `json:"MVPs,string"`
	MatchRound          int     `json:"Match Round,string"`
	BestOf              int     `json:"Best Of,string"`
	QuadroKills         int     `json:"Quadro Kills,string"`
}
type MatchHistory struct {
	Nickname            string
	Team                string
	Total_Wins          int
	Total_Losses        int
	Total_Matches       int
	Total_KDRatio       float64
	Total_HS_Percentage float64
}

// Wrapper for the FACEIT stats list response
type playerStatsList struct {
	Items []struct {
		Stats Stats `json:"stats"`
	} `json:"items"`
	Start int `json:"start"`
	End   int `json:"end"`
}

func getMatchHistory(start, end int64, human_start, human_end string) string {
	var discordMessage string
	discordMessage += "**Match History**: " + human_start + " -> " + human_end + "\n\n"
	faceitPlayers := getPlayerIDs()
	// Endpoint is /players/{player_id}/games/cs2/stats?from=<INTEGER>&to=<INTEGER>&offset=0&limit=30
	log.Println("Getting match history for", len(faceitPlayers), "players")

	type runningTotals struct {
		kills     int
		deaths    int
		headshots int
	}

	aggregates := make(map[string]*MatchHistory) // key: PlayerID
	sums := make(map[string]*runningTotals)      // key: PlayerID

	for _, player := range faceitPlayers {
		// Ensure players with zero matches still appear in aggregates/sums
		if _, ok := aggregates[player.PlayerID]; !ok {
			aggregates[player.PlayerID] = &MatchHistory{Nickname: player.PlayerName}
		}
		if _, ok := sums[player.PlayerID]; !ok {
			sums[player.PlayerID] = &runningTotals{}
		}
		response := QueryFACEITAPI("/players/"+player.PlayerID+"/games/cs2/stats", map[string]interface{}{
			"from":   start,
			"to":     end,
			"offset": 0,
			"limit":  30,
		})
		body, err := io.ReadAll(response.Body)
		if err != nil {
			log.Fatal("Error reading response body: ", err)
		}
		response.Body.Close()

		var list playerStatsList
		if err := json.Unmarshal(body, &list); err != nil {
			log.Printf("Failed to unmarshal player %s stats: %v", player.PlayerName, err)
			continue
		}
		for _, it := range list.Items {
			s := it.Stats
			// If TEAM_NAME is set, discard stats for that team. This will filter out league games and only include pugs
			if teamName != "" && s.Team == teamName {
				continue
			}
			// Initialize on first encounter
			if _, ok := aggregates[s.PlayerID]; !ok {
				aggregates[s.PlayerID] = &MatchHistory{Nickname: s.Nickname, Team: s.Team}
			}
			if _, ok := sums[s.PlayerID]; !ok {
				sums[s.PlayerID] = &runningTotals{}
			}

			mh := aggregates[s.PlayerID]
			rt := sums[s.PlayerID]

			mh.Total_Matches++
			if s.Result == 1 {
				mh.Total_Wins++
			} else {
				mh.Total_Losses++
			}

			rt.kills += s.Kills
			rt.deaths += s.Deaths
			rt.headshots += s.Headshots
		}
	}

	//Sort the playerID by Total Matches
	playerIDs := make([]string, 0, len(aggregates))
	for playerID := range aggregates {
		playerIDs = append(playerIDs, playerID)
	}

	// Sort by Total_Matches desc, then Nickname case-insensitive asc
	sort.Slice(playerIDs, func(i, j int) bool {
		ai := aggregates[playerIDs[i]]
		aj := aggregates[playerIDs[j]]
		if ai.Total_Matches != aj.Total_Matches {
			return ai.Total_Matches > aj.Total_Matches
		}
		return strings.ToLower(ai.Nickname) < strings.ToLower(aj.Nickname)
	})

	var builder bytes.Buffer
	builder.Reset()
	table := tablewriter.NewTable(&builder)
	table.Header([]string{"NAME", "MATCHES", "W-L", "KD", "HS%"})

	// Finalize computed ratios and log
	for _, playerID := range playerIDs {
		mh := aggregates[playerID]
		rt := sums[playerID]
		if rt != nil {
			if rt.deaths > 0 {
				mh.Total_KDRatio = float64(rt.kills) / float64(rt.deaths)
			}
			if rt.kills > 0 {
				mh.Total_HS_Percentage = (float64(rt.headshots) / float64(rt.kills)) * 100.0
			}
		}
		if mh.Total_Matches > 0 {
			table.Append([]string{mh.Nickname, strconv.Itoa(mh.Total_Matches), fmt.Sprintf("%d-%d", mh.Total_Wins, mh.Total_Losses), fmt.Sprintf("%.2f", mh.Total_KDRatio), fmt.Sprintf("%.1f", mh.Total_HS_Percentage)})
			// discordMessage += fmt.Sprintf("**%s**: Matches %d, W-L %d-%d, KD %.2f, HS%% %.1f\n",
			// mh.Nickname, mh.Total_Matches, mh.Total_Wins, mh.Total_Losses, mh.Total_KDRatio, mh.Total_HS_Percentage)
		} else {
			table.Append([]string{mh.Nickname, "0", "0-0", "0.00", "0.0"})
			// discordMessage += fmt.Sprintf("**%s**: No matches played this week\n", mh.Nickname)
		}

	}
	table.Render()
	table = nil
	return builder.String()
}

// ------------------------------------------------------------
// Discord Slash Commands
// ------------------------------------------------------------

// get a player's detailed stats over the last 7 days
// get a player's detailed league stats over the last 3 months
func ListPlayers() string {
	var players []FACEITPlayers = []FACEITPlayers{}
	players = getPlayerIDs()
	// Sort players by PlayerName case-insensitive asc
	sort.Slice(players, func(i, j int) bool {
		return strings.ToLower(players[i].PlayerName) < strings.ToLower(players[j].PlayerName)
	})
	var builder bytes.Buffer
	builder.Reset()
	table := tablewriter.NewTable(&builder)
	table.Header([]string{"NAME", "ID"})
	for _, player := range players {
		table.Append([]string{player.PlayerName, player.PlayerID})
	}
	table.Render()
	table = nil
	return "```" + builder.String() + "```"
}

func AddPlayer(playerName string) string {
	// Load existing names in the expected structure
	names := loadPlayerJSON()

	// Check if the player already exists
	for _, name := range names.PlayerName {
		if name == playerName {
			return "Player already exists: " + playerName
		}
	}

	// Append and sort case-insensitively
	names.PlayerName = append(names.PlayerName, playerName)
	sort.Slice(names.PlayerName, func(i, j int) bool {
		return strings.ToLower(names.PlayerName[i]) < strings.ToLower(names.PlayerName[j])
	})

	// Persist to data/faceit_player_names.json with the correct structure
	data, err := json.MarshalIndent(names, "", "    ")
	if err != nil {
		log.Fatal("Error marshalling player names: ", err)
	}
	if err := os.WriteFile("data/faceit_player_names.json", data, 0644); err != nil {
		log.Fatal("Error writing faceit_player_names.json: ", err)
	}

	return "Player added: " + playerName
}

func RemovePlayer(playerName string) string {
	names := loadPlayerJSON()
	for i, name := range names.PlayerName {
		if name == playerName {
			names.PlayerName = append(names.PlayerName[:i], names.PlayerName[i+1:]...)
			break
		}
	}
	data, err := json.MarshalIndent(names, "", "    ")
	if err != nil {
		log.Fatal("Error marshalling player names: ", err)
	}
	if err := os.WriteFile("data/faceit_player_names.json", data, 0644); err != nil {
		log.Fatal("Error writing faceit_player_names.json: ", err)
	}
	return "Player removed: " + playerName
}

func FACEITInit(s *discordgo.Session) string {

	// LAST WEEK
	start, end, human_start, human_end := CurrentWeekWindow(time.Now().AddDate(0, 0, -7))
	log.Printf("Last Week: %s -> %s", human_start, human_end)
	discordMessage := getMatchHistory(start, end, human_start, human_end)
	marker := "**Last Week -- Match History**: " + human_start + " -> " + human_end
	msg := marker + "\n\n" + "```" + discordMessage + "```"
	UpdateMessage(s, msg, marker)
	// UpdatePresence(s, msg, marker)

	// CURRENT WEEK
	start, end, human_start, human_end = CurrentWeekWindow(time.Now())
	log.Printf("Current Week: %s -> %s", human_start, human_end)
	discordMessage = getMatchHistory(start, end, human_start, human_end)
	marker = "**Current Week -- Match History**: " + human_start + " -> " + human_end
	msg = marker + "\n\n" + "```" + discordMessage + "```"
	UpdateMessage(s, msg, marker)
	// UpdatePresence(s, msg, marker)
	discordMessage = ""
	msg = ""
	log.Println("Listening and READY")
	return "Refreshed!"
}

// StartFACEITRefresher runs FACEIT refresh immediately and then every hour until stopCh is closed.
func StartFACEITRefresher(s *discordgo.Session, stopCh <-chan struct{}) {
	run := func() {
		FACEITInit(s)
	}

	// Run once immediately
	run()

	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			log.Println("Hourly FACEIT refresh")
			run()
		case <-stopCh:
			log.Println("Stopping FACEIT refresher")
			return
		}
	}
}
