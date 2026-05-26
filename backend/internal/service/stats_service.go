package service

import (
	"context"
	"errors"
	"math"
	"sort"
	"strings"
	"time"
)

var ErrInvalidStatsInput = errors.New("invalid stats input")

const maxStatsRange = 366 * 24 * time.Hour

type StatsEntryQuery struct {
	StartTime time.Time
	EndTime   time.Time
	Project   string
}

type StatsRange struct {
	StartTime int64
	EndTime   int64
}

type ReviewStatsLogEntry struct {
	ProjectID         uint
	ProjectName       string
	Author            string
	AuthorIdentity    string
	AuthorDisplayName string
	Score             int
	Additions         int
	Deletions         int
	CreatedAt         int64
}

type StatsOverview struct {
	ActiveProjects           int                   `json:"activeProjects"`
	Contributors             int                   `json:"contributors"`
	TotalCommits             int64                 `json:"totalCommits"`
	AverageScore             float64               `json:"averageScore"`
	ProjectCommitStats       []ProjectStatsItem    `json:"projectCommitStats"`
	ProjectAverageScoreStats []ProjectStatsItem    `json:"projectAverageScoreStats"`
	AuthorCommitStats        []AuthorStatsItem     `json:"authorCommitStats"`
	AuthorAverageScoreStats  []AuthorStatsItem     `json:"authorAverageScoreStats"`
	ProjectCodeChangeStats   []CodeChangeStatsItem `json:"projectCodeChangeStats"`
	AuthorCodeChangeStats    []CodeChangeStatsItem `json:"authorCodeChangeStats"`
}

type ProjectStatsItem struct {
	ProjectName  string  `json:"projectName"`
	CommitCount  int64   `json:"commitCount"`
	AverageScore float64 `json:"averageScore"`
}

type AuthorStatsItem struct {
	Author       string  `json:"author"`
	CommitCount  int64   `json:"commitCount"`
	AverageScore float64 `json:"averageScore"`
}

type CodeChangeStatsItem struct {
	Name      string `json:"name"`
	Additions int64  `json:"additions"`
	Deletions int64  `json:"deletions"`
}

type MemberCommitSummaryQuery struct {
	StartTime int64
	EndTime   int64
	Project   string
	Page      int
	Size      int
}

type MemberCommitStatsPage struct {
	Items []MemberCommitStats `json:"items"`
	Total int64               `json:"total"`
	Page  int                 `json:"page"`
	Size  int                 `json:"size"`
}

type MemberCommitStats struct {
	Author            string              `json:"author"`
	AuthorIdentity    string              `json:"authorIdentity"`
	AuthorDisplayName string              `json:"authorDisplayName"`
	AuthorDisplayText string              `json:"authorDisplayText"`
	Summary           MemberCommitSummary `json:"summary"`
	DailyCommits      []DailyCommitStats  `json:"dailyCommits"`
}

type MemberCommitSummary struct {
	ProjectCount         int     `json:"projectCount"`
	CommitCount          int64   `json:"commitCount"`
	AvgCommitCountPerDay float64 `json:"avgCommitCountPerDay"`
	AvgScore             float64 `json:"avgScore"`
	AdditionCount        int64   `json:"additionCount"`
	DeletionCount        int64   `json:"deletionCount"`
}

type DailyCommitStats struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type StatsRepository interface {
	ListReviewStatsEntries(ctx context.Context, query StatsEntryQuery) ([]ReviewStatsLogEntry, error)
}

type StatsService struct {
	stats StatsRepository
}

func NewStatsService(stats StatsRepository) *StatsService {
	return &StatsService{stats: stats}
}

func (s *StatsService) GetStats(ctx context.Context, query StatsRange) (*StatsOverview, error) {
	entryQuery, err := normalizeStatsEntryQuery(query.StartTime, query.EndTime, "")
	if err != nil {
		return nil, err
	}
	rows, err := s.stats.ListReviewStatsEntries(ctx, entryQuery)
	if err != nil {
		return nil, err
	}
	overview := buildStatsOverview(rows)
	return &overview, nil
}

func (s *StatsService) GetMemberCommitSummary(ctx context.Context, query MemberCommitSummaryQuery) (*MemberCommitStatsPage, error) {
	query.Page, query.Size = normalizeStatsPage(query.Page, query.Size)
	entryQuery, err := normalizeStatsEntryQuery(query.StartTime, query.EndTime, query.Project)
	if err != nil {
		return nil, err
	}
	rows, err := s.stats.ListReviewStatsEntries(ctx, entryQuery)
	if err != nil {
		return nil, err
	}
	items := buildMemberCommitStats(rows, entryQuery.StartTime, entryQuery.EndTime)
	total := int64(len(items))
	start := (query.Page - 1) * query.Size
	if start > len(items) {
		start = len(items)
	}
	end := start + query.Size
	if end > len(items) {
		end = len(items)
	}
	return &MemberCommitStatsPage{Items: items[start:end], Total: total, Page: query.Page, Size: query.Size}, nil
}

func normalizeStatsEntryQuery(startTime int64, endTime int64, project string) (StatsEntryQuery, error) {
	if startTime <= 0 || endTime <= 0 || startTime >= endTime {
		return StatsEntryQuery{}, ErrInvalidStatsInput
	}
	start := time.UnixMilli(startTime)
	end := time.UnixMilli(endTime)
	if end.Sub(start) > maxStatsRange {
		return StatsEntryQuery{}, ErrInvalidStatsInput
	}
	return StatsEntryQuery{
		StartTime: start,
		EndTime:   end,
		Project:   strings.TrimSpace(project),
	}, nil
}

func buildStatsOverview(rows []ReviewStatsLogEntry) StatsOverview {
	projectIDs := map[uint]struct{}{}
	authors := map[string]struct{}{}
	projectStats := map[string]*statsAccumulator{}
	authorStats := map[string]*statsAccumulator{}
	var totalScore int64
	var scoredCount int64

	for _, row := range rows {
		projectIDs[row.ProjectID] = struct{}{}
		authorKey := authorIdentity(row)
		authors[authorKey] = struct{}{}
		if row.Score > 0 {
			totalScore += int64(row.Score)
			scoredCount++
		}

		project := projectStats[row.ProjectName]
		if project == nil {
			project = &statsAccumulator{}
			projectStats[row.ProjectName] = project
		}
		project.add(row.Score, row.Additions, row.Deletions)

		author := authorStats[authorKey]
		if author == nil {
			author = &statsAccumulator{displayName: row.AuthorDisplayName}
			authorStats[authorKey] = author
		}
		if author.displayName == "" && row.AuthorDisplayName != "" {
			author.displayName = row.AuthorDisplayName
		}
		author.add(row.Score, row.Additions, row.Deletions)
	}

	stats := StatsOverview{
		ActiveProjects: len(projectIDs),
		Contributors:   len(authors),
		TotalCommits:   int64(len(rows)),
		AverageScore:   averageScore(totalScore, scoredCount),
	}
	stats.ProjectCommitStats = buildProjectStatsItems(projectStats)
	stats.ProjectAverageScoreStats = append([]ProjectStatsItem(nil), stats.ProjectCommitStats...)
	stats.AuthorCommitStats = buildAuthorStatsItems(authorStats)
	stats.AuthorAverageScoreStats = append([]AuthorStatsItem(nil), stats.AuthorCommitStats...)
	stats.ProjectCodeChangeStats = buildProjectCodeChangeItems(projectStats)
	stats.AuthorCodeChangeStats = buildAuthorCodeChangeItems(authorStats)
	return stats
}

type statsAccumulator struct {
	count       int64
	scoredCount int64
	scoreTotal  int64
	additions   int64
	deletions   int64
	displayName string
}

func (a *statsAccumulator) add(score int, additions int, deletions int) {
	a.count++
	if score > 0 {
		a.scoredCount++
		a.scoreTotal += int64(score)
	}
	a.additions += int64(additions)
	a.deletions += int64(deletions)
}

func buildProjectStatsItems(stats map[string]*statsAccumulator) []ProjectStatsItem {
	items := make([]ProjectStatsItem, 0, len(stats))
	for name, item := range stats {
		items = append(items, ProjectStatsItem{ProjectName: name, CommitCount: item.count, AverageScore: averageScore(item.scoreTotal, item.scoredCount)})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].CommitCount == items[j].CommitCount {
			return items[i].ProjectName < items[j].ProjectName
		}
		return items[i].CommitCount > items[j].CommitCount
	})
	return items
}

func buildAuthorStatsItems(stats map[string]*statsAccumulator) []AuthorStatsItem {
	items := make([]AuthorStatsItem, 0, len(stats))
	for author, item := range stats {
		items = append(items, AuthorStatsItem{Author: authorDisplayText(author, item.displayName), CommitCount: item.count, AverageScore: averageScore(item.scoreTotal, item.scoredCount)})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].CommitCount == items[j].CommitCount {
			return items[i].Author < items[j].Author
		}
		return items[i].CommitCount > items[j].CommitCount
	})
	return items
}

func buildProjectCodeChangeItems(stats map[string]*statsAccumulator) []CodeChangeStatsItem {
	items := make([]CodeChangeStatsItem, 0, len(stats))
	for name, item := range stats {
		items = append(items, CodeChangeStatsItem{Name: name, Additions: item.additions, Deletions: item.deletions})
	}
	sortCodeChangeItems(items)
	return items
}

func buildAuthorCodeChangeItems(stats map[string]*statsAccumulator) []CodeChangeStatsItem {
	items := make([]CodeChangeStatsItem, 0, len(stats))
	for author, item := range stats {
		items = append(items, CodeChangeStatsItem{Name: authorDisplayText(author, item.displayName), Additions: item.additions, Deletions: item.deletions})
	}
	sortCodeChangeItems(items)
	return items
}

func sortCodeChangeItems(items []CodeChangeStatsItem) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Additions == items[j].Additions {
			return items[i].Name < items[j].Name
		}
		return items[i].Additions > items[j].Additions
	})
}

func buildMemberCommitStats(rows []ReviewStatsLogEntry, start time.Time, end time.Time) []MemberCommitStats {
	type memberAccumulator struct {
		displayName string
		projects    map[string]struct{}
		daily       map[string]int
		statsAccumulator
	}
	members := map[string]*memberAccumulator{}
	for _, row := range rows {
		author := authorIdentity(row)
		member := members[author]
		if member == nil {
			member = &memberAccumulator{displayName: row.AuthorDisplayName, projects: map[string]struct{}{}, daily: map[string]int{}}
			members[author] = member
		}
		if member.displayName == "" && row.AuthorDisplayName != "" {
			member.displayName = row.AuthorDisplayName
		}
		member.projects[row.ProjectName] = struct{}{}
		member.add(row.Score, row.Additions, row.Deletions)
		member.daily[time.UnixMilli(row.CreatedAt).Format("2006-01-02")]++
	}

	days := daysInclusive(start, end)
	items := make([]MemberCommitStats, 0, len(members))
	for author, member := range members {
		items = append(items, MemberCommitStats{
			Author:            author,
			AuthorIdentity:    author,
			AuthorDisplayName: member.displayName,
			AuthorDisplayText: authorDisplayText(author, member.displayName),
			Summary: MemberCommitSummary{
				ProjectCount:         len(member.projects),
				CommitCount:          member.count,
				AvgCommitCountPerDay: round2(float64(member.count) / float64(days)),
				AvgScore:             averageScore(member.scoreTotal, member.scoredCount),
				AdditionCount:        member.additions,
				DeletionCount:        member.deletions,
			},
			DailyCommits: buildDailyCommitStats(member.daily, start, end),
		})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Summary.CommitCount == items[j].Summary.CommitCount {
			return items[i].AuthorIdentity < items[j].AuthorIdentity
		}
		return items[i].Summary.CommitCount > items[j].Summary.CommitCount
	})
	return items
}

func buildDailyCommitStats(counts map[string]int, start time.Time, end time.Time) []DailyCommitStats {
	startDay := dayStart(start)
	endDay := dayStart(end)
	items := make([]DailyCommitStats, 0, daysInclusive(start, end))
	for current := startDay; !current.After(endDay); current = current.AddDate(0, 0, 1) {
		date := current.Format("2006-01-02")
		items = append(items, DailyCommitStats{Date: date, Count: counts[date]})
	}
	return items
}

func daysInclusive(start time.Time, end time.Time) int {
	days := int(dayStart(end).Sub(dayStart(start)).Hours()/24) + 1
	if days < 1 {
		return 1
	}
	return days
}

func dayStart(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, value.Location())
}

func authorIdentity(row ReviewStatsLogEntry) string {
	if strings.TrimSpace(row.AuthorIdentity) != "" {
		return strings.TrimSpace(row.AuthorIdentity)
	}
	if strings.TrimSpace(row.Author) != "" {
		return strings.TrimSpace(row.Author)
	}
	return strings.TrimSpace(row.AuthorDisplayName)
}

func averageScore(total int64, count int64) float64 {
	if count == 0 {
		return 0
	}
	return round2(float64(total) / float64(count))
}

func normalizeStatsPage(page int, size int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	if size > 200 {
		size = 200
	}
	return page, size
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}
