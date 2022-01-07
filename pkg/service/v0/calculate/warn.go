package calculate

import (
	"fmt"
	"owl-engine/pkg/client/database"
	"owl-engine/pkg/util"
	"sync"
	"time"

	"owl-engine/pkg/lib/job"
	"owl-engine/pkg/xlogs"

	uuid "github.com/satori/go.uuid"
)

type RuleWarn struct{}

func Warn(stopCh <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	cronTab := job.NewCronTab()

	id := uuid.NewV4().String()
	warn := new(RuleWarn)
	if err := cronTab.AddByID(id, "* * * * *", warn); err != nil {
		xlogs.Errorf("failed to add rule closing alarm timing task, %s", err.Error())
		return
	} else {
		xlogs.Info("succeed to add rule closing alarm timing task")
	}

	cronTab.Start()

	select {
	case <-stopCh:
		ids := cronTab.IDs()
		for _, id := range ids {
			cronTab.DelByID(id)
		}
		cronTab.Stop()

		xlogs.Info("stop scheduled reminder task")
		return
	}
}

func (r *RuleWarn) Run() {
	var warnRules = make([]string, 0)
	timePast := util.DateTimeToString(time.Now().Add(-2 * time.Hour))

	for _, table := range []string{"engine_tbl_rules", "engine_tbl_logger_rules"} {
		var names = make([]string, 0)
		sql := fmt.Sprintf("SELECT `name` FROM `%s` WHERE `inuse` = %d AND `switch` = %d AND `updated_at` <= '%s';",
			table, 1, 2, timePast)

		err := database.DB.Raw(sql).Scan(&names).Error
		if err == nil {
			warnRules = append(warnRules, names...)
		}
	}

	if len(warnRules) > 0 {
		// 发送告警
		
	}
}
