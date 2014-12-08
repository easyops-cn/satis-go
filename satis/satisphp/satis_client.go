package satisphp

import (
	"errors"
	"github.com/benschw/satis-go/satis/satisphp/api"
	"github.com/benschw/satis-go/satis/satisphp/db"
	"github.com/benschw/satis-go/satis/satisphp/job"
	"log"
)

var _ = log.Print

var ErrRepoNotFound = errors.New("Repository Not Found")

type SatisClient struct {
	Jobs   chan job.SatisJob
	DbPath string
}

func (s *SatisClient) FindAllRepos() ([]api.Repo, error) {
	j := job.NewFindAllJob(s.DbPath)

	err := s.performJob(j)

	repos := <-j.ReposResp

	rs := make([]api.Repo, len(repos), len(repos))
	for i, repo := range repos {
		rs[i] = *api.NewRepo(repo.Type, repo.Url)
	}

	return rs, err
}

func (s *SatisClient) SaveRepo(repo *api.Repo) error {
	repoEntity := db.SatisRepository{
		Type: repo.Type,
		Url:  repo.Url,
	}
	j := job.NewSaveRepoJob(s.DbPath, repoEntity, true)
	return s.performJob(j)
}

func (s *SatisClient) DeleteRepo(id string) error {
	var toDelete api.Repo

	repos, err := s.FindAllRepos()
	if err != nil {
		return err
	}

	found := false
	for _, r := range repos {
		if r.Id == id {
			found = true
			toDelete = r
		}
	}

	if found {
		j := job.NewDeleteRepoJob(s.DbPath, toDelete.Url, true)
		err = s.performJob(j)

		switch err {
		case job.ErrRepoNotFound:
			return ErrRepoNotFound
		default:
			return err
		}

	} else {
		return ErrRepoNotFound
	}
}

func (s *SatisClient) GenerateSatisWeb() error {
	j := job.NewGenerateJob()
	return s.performJob(j)
}

func (s *SatisClient) Shutdown() error {
	j := job.NewExitJob()
	return s.performJob(j)
}

func (s *SatisClient) performJob(j job.SatisJob) error {
	s.Jobs <- j

	return <-j.ExitChan()
}