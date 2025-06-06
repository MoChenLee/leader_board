package model

type RankInfo struct {
	PlayerID  string
	Score     int
	Rank      int
	Timestamp int64
}

type PlayerEntry struct {
	PlayerID  string
	Score     int
	Timestamp int64
}
