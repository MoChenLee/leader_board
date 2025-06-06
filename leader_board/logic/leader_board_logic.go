package logic

import (
	"leader_board/leader_board/model"
	"sort"
	"sync"
)

type LeaderboardService struct {
	mu      sync.RWMutex
	players map[string]*model.PlayerEntry
	ranking []*model.PlayerEntry
}

// NewLeaderboardService 创建排行榜服务
func NewLeaderboardService() *LeaderboardService {
	return &LeaderboardService{
		players: make(map[string]*model.PlayerEntry),
		ranking: make([]*model.PlayerEntry, 0),
	}
}

// 更新玩家分数
func (lb *LeaderboardService) UpdateScore(playerId string, score int, timestamp int64) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	entry, exists := lb.players[playerId]
	if exists {
		// 先移除旧的
		lb.removeFromRanking(entry)
		entry.Score = score
		entry.Timestamp = timestamp
	} else {
		entry = &model.PlayerEntry{PlayerID: playerId, Score: score, Timestamp: timestamp}
		lb.players[playerId] = entry
	}
	lb.insertToRanking(entry)
}

// 获取玩家当前排名
func (lb *LeaderboardService) GetPlayerRank(playerId string) *model.RankInfo {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	entry, exists := lb.players[playerId]
	if !exists {
		return nil
	}
	rank := lb.findRank(entry)
	return &model.RankInfo{
		PlayerID:  entry.PlayerID,
		Score:     entry.Score,
		Rank:      rank + 1,
		Timestamp: entry.Timestamp,
	}
}

// 获取前N名
func (lb *LeaderboardService) GetTopN(n int) []model.RankInfo {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	res := make([]model.RankInfo, 0, n)
	for i := 0; i < n && i < len(lb.ranking); i++ {
		entry := lb.ranking[i]
		res = append(res, model.RankInfo{
			PlayerID:  entry.PlayerID,
			Score:     entry.Score,
			Rank:      i + 1,
			Timestamp: entry.Timestamp,
		})
	}
	return res
}

// 获取玩家周边排名
func (lb *LeaderboardService) GetPlayerRankRange(playerId string, rng int) []model.RankInfo {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	entry, exists := lb.players[playerId]
	if !exists {
		return nil
	}
	rank := lb.findRank(entry)
	start := rank - rng
	if start < 0 {
		start = 0
	}
	end := rank + rng
	if end >= len(lb.ranking) {
		end = len(lb.ranking) - 1
	}
	res := make([]model.RankInfo, 0, end-start+1)
	for i := start; i <= end; i++ {
		e := lb.ranking[i]
		res = append(res, model.RankInfo{
			PlayerID:  e.PlayerID,
			Score:     e.Score,
			Rank:      i + 1,
			Timestamp: e.Timestamp,
		})
	}
	return res
}

// 内部方法：插入到排行榜
func (lb *LeaderboardService) insertToRanking(entry *model.PlayerEntry) {
	idx := sort.Search(len(lb.ranking), func(i int) bool {
		if lb.ranking[i].Score < entry.Score {
			return true
		}
		if lb.ranking[i].Score == entry.Score {
			return lb.ranking[i].Timestamp > entry.Timestamp
		}
		return false
	})
	lb.ranking = append(lb.ranking, nil)
	copy(lb.ranking[idx+1:], lb.ranking[idx:])
	lb.ranking[idx] = entry
}

// 内部方法：移除旧排名
func (lb *LeaderboardService) removeFromRanking(entry *model.PlayerEntry) {
	for i, e := range lb.ranking {
		if e == entry {
			lb.ranking = append(lb.ranking[:i], lb.ranking[i+1:]...)
			return
		}
	}
}

// 内部方法：查找排名
func (lb *LeaderboardService) findRank(entry *model.PlayerEntry) int {
	for i, e := range lb.ranking {
		if e == entry {
			return i
		}
	}
	return -1
}

// ------------------------------------以下密集排名的方法------------------------------------

func (lb *LeaderboardService) findDenseRank(entry *model.PlayerEntry) int {
	rank := 1
	prevScore := -1
	for i, e := range lb.ranking {
		if i == 0 || e.Score != prevScore {
			if i != 0 {
				rank++
			}
			prevScore = e.Score
		}
		if e == entry {
			return rank
		}
	}
	return -1
}

func (lb *LeaderboardService) GetDenseTopN(n int) []model.RankInfo {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	res := make([]model.RankInfo, 0, n)
	rank := 1
	prevScore := -1
	for i := 0; i < n && i < len(lb.ranking); i++ {
		entry := lb.ranking[i]
		if i == 0 || entry.Score != prevScore {
			if i != 0 {
				rank++
			}
			prevScore = entry.Score
		}
		res = append(res, model.RankInfo{
			PlayerID:  entry.PlayerID,
			Score:     entry.Score,
			Rank:      rank,
			Timestamp: entry.Timestamp,
		})
	}
	return res
}

func (lb *LeaderboardService) GetPlayerDenseRankRange(playerId string, rng int) []model.RankInfo {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	entry, exists := lb.players[playerId]
	if !exists {
		return nil
	}

	denseRanks := make([]int, len(lb.ranking))
	rank := 1
	prevScore := -1
	for i, e := range lb.ranking {
		if i == 0 || e.Score != prevScore {
			if i != 0 {
				rank++
			}
			prevScore = e.Score
		}
		denseRanks[i] = rank
	}
	//idx := 0 // 以玩家为中心前后rng个 才会使用这个
	playerDenseRank := 0
	for i, e := range lb.ranking {
		if e == entry {
			//idx = i
			playerDenseRank = denseRanks[i]
			break
		}
	}

	res := make([]model.RankInfo, 0, 2*rng+1)
	for i, r := range denseRanks {
		if r >= playerDenseRank-rng && r <= playerDenseRank+rng {
			e := lb.ranking[i]
			res = append(res, model.RankInfo{
				PlayerID:  e.PlayerID,
				Score:     e.Score,
				Rank:      r,
				Timestamp: e.Timestamp,
			})
		}
	}
	return res
}
