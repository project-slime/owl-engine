package job

import (
	"errors"
	"sync"

	"github.com/robfig/cron/v3"
)

type CronTab struct {
	inner *cron.Cron
	ids   map[string]cron.EntryID
	mutex sync.Mutex
}

func NewCronTab() *CronTab {
	return &CronTab{
		inner: cron.New(cron.WithParser(cron.NewParser(
			cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
		))),
		ids: make(map[string]cron.EntryID),
	}
}

// ids
func (c *CronTab) IDs() []string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	validIDs := make([]string, 0, len(c.ids))
	invalidIDs := make([]string, 0)
	for sid, eid := range c.ids {
		if e := c.inner.Entry(eid); e.ID != eid {
			invalidIDs = append(invalidIDs, sid)
			continue
		}

		validIDs = append(validIDs, sid)
	}

	for _, id := range invalidIDs {
		delete(c.ids, id)
	}

	return validIDs
}

// start the crontab engine
func (c *CronTab) Start() {
	c.inner.Start()
}

// stop the crontab engine
func (c *CronTab) Stop() {
	c.inner.Stop()
}

// DelByID remove one crontab task
func (c *CronTab) DelByID(id string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	eid, ok := c.ids[id]
	if !ok {
		return
	}
	c.inner.Remove(eid)
	delete(c.ids, id)
}

// AddByID add one crontab task
// id is unique
// spec is the crontab expression
func (c *CronTab) AddByID(id string, spec string, cmd cron.Job) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, ok := c.ids[id]; ok {
		return errors.New("crontab id exists")
	}
	eid, err := c.inner.AddJob(spec, cmd)
	if err != nil {
		return err
	}
	c.ids[id] = eid
	return nil
}

// AddByFunc add function as crontab task
func (c *CronTab) AddByFunc(id string, spec string, f func()) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, ok := c.ids[id]; ok {
		return errors.New("crontab id exists")
	}
	eid, err := c.inner.AddFunc(spec, f)
	if err != nil {
		return err
	}
	c.ids[id] = eid
	return nil
}

// IsExists check the crontab task whether existed with job id
func (c *CronTab) IsExists(jid string) bool {
	_, exist := c.ids[jid]
	return exist
}
